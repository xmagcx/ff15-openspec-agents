package ff15

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	TargetRoot    string
	DryRun        bool
	AutoApprove   bool
	Ecosystems    []Ecosystem
	OptionalTools []ToolName
}

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	_ = stderr
	if len(args) > 0 {
		switch args[0] {
		case "init":
			return runInit(args[1:], stdin, stdout)
		case "sync":
			return runSync(args[1:], stdin, stdout)
		case "help":
			printUsage(stdout)
			return nil
		}
	}

	fs := flag.NewFlagSet("ff15", flag.ContinueOnError)
	fs.SetOutput(stdout)
	_ = fs.Bool("init", false, "start the FF15 init flow explicitly")
	target := fs.String("target", ".", "target project directory")
	dryRun := fs.Bool("dry-run", false, "print checkpoints without applying changes")
	yes := fs.Bool("yes", false, "approve the generated plan without prompting")
	ecosystemsCSV := fs.String("ecosystems", "", "comma-separated ecosystems: claude,kiro,pi,opencode")
	optionalToolsCSV := fs.String("optional-tools", "", "comma-separated optional tools: headroom,rtk")
	fs.Usage = func() {
		printUsage(stdout)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Flags:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	return executeInit(Config{
		TargetRoot:    *target,
		DryRun:        *dryRun,
		AutoApprove:   *yes,
		Ecosystems:    parseEcosystems(*ecosystemsCSV),
		OptionalTools: parseOptionalTools(*optionalToolsCSV),
	}, stdin, stdout)
}

func printUsage(stdout io.Writer) {
	fmt.Fprintln(stdout, "Usage:")
	fmt.Fprintln(stdout, "  ff15 [flags]")
	fmt.Fprintln(stdout, "  ff15 init [flags]")
	fmt.Fprintln(stdout, "  ff15 sync [flags]")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Commands:")
	fmt.Fprintln(stdout, "  init    start the FF15 init flow")
	fmt.Fprintln(stdout, "  sync    sync FF15 agents, prompts, docs, and AGENTS.md")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Running `ff15` without a command starts the init flow by default.")
}

func runInit(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("ff15 init", flag.ContinueOnError)
	fs.SetOutput(stdout)
	target := fs.String("target", ".", "target project directory")
	dryRun := fs.Bool("dry-run", false, "print checkpoints without applying changes")
	yes := fs.Bool("yes", false, "approve the generated plan without prompting")
	ecosystemsCSV := fs.String("ecosystems", "", "comma-separated ecosystems: claude,kiro,pi,opencode")
	optionalToolsCSV := fs.String("optional-tools", "", "comma-separated optional tools: headroom,rtk")
	fs.Usage = func() {
		printUsage(stdout)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Flags:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	return executeInit(Config{
		TargetRoot:    *target,
		DryRun:        *dryRun,
		AutoApprove:   *yes,
		Ecosystems:    parseEcosystems(*ecosystemsCSV),
		OptionalTools: parseOptionalTools(*optionalToolsCSV),
	}, stdin, stdout)
}

func executeInit(cfg Config, stdin io.Reader, stdout io.Writer) error {
	platform, err := DetectPlatform("")
	if err != nil {
		return err
	}

	cfg.TargetRoot = filepath.Clean(cfg.TargetRoot)
	if shouldRunWizard(cfg) {
		if !supportsInteractiveWizard(stdin, stdout) {
			return fmt.Errorf("interactive mode requires a terminal; pass --ecosystems, --optional-tools, and --yes for non-interactive use")
		}
		result, err := RunWizard(stdin, stdout, cfg, platform)
		if err != nil {
			return err
		}
		if result.Cancelled {
			fmt.Fprintln(stdout, "Aborted before execution.")
			return nil
		}
		cfg.Ecosystems = result.Ecosystems
		cfg.OptionalTools = result.OptionalTools
		cfg.AutoApprove = result.Approved
	}

	plan, err := BuildPlan(cfg.TargetRoot, platform, cfg.Ecosystems, cfg.OptionalTools)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, renderPlanSummary(plan, DetectCoverInfo()))
	if cfg.DryRun {
		fmt.Fprintln(stdout, "Dry run complete. No changes applied.")
		return nil
	}
	if !cfg.AutoApprove {
		return fmt.Errorf("plan approval required; rerun with --yes or use the interactive wizard in a terminal")
	}

	return ExecutePlan(plan, stdout)
}

func shouldRunWizard(cfg Config) bool {
	return len(cfg.Ecosystems) == 0 || (!cfg.AutoApprove && !cfg.DryRun)
}

func supportsInteractiveWizard(stdin io.Reader, stdout io.Writer) bool {
	inFile, inOK := stdin.(*os.File)
	outFile, outOK := stdout.(*os.File)
	if !inOK || !outOK {
		return false
	}
	inStat, err := inFile.Stat()
	if err != nil {
		return false
	}
	outStat, err := outFile.Stat()
	if err != nil {
		return false
	}
	return (inStat.Mode()&os.ModeCharDevice) != 0 && (outStat.Mode()&os.ModeCharDevice) != 0
}

func renderPlanSummary(plan Plan, cover CoverInfo) string {
	lines := []string{fmt.Sprintf("FF15 init plan for %s (%s)", plan.TargetRoot, plan.Platform)}
	if cover.Exists {
		lines = append(lines, "Cover: "+cover.Summary())
	}
	lines = append(lines, renderPlanSummaryLines(plan)...)
	return strings.Join(lines, "\n")
}

func renderPlanSummaryLines(plan Plan) []string {
	lines := []string{
		fmt.Sprintf("Ecosystems: %s", joinEcosystems(plan.Ecosystems)),
		fmt.Sprintf("Tools: %s", joinTools(plan.SelectedTools)),
		fmt.Sprintf("Steps: %d", len(plan.Steps)),
	}
	for i, step := range plan.Steps {
		summary := step.Title
		if step.Kind == StepPatch {
			summary = filepath.ToSlash(step.FilePath)
		}
		lines = append(lines, fmt.Sprintf("%d. [%s] %s", i+1, step.Kind, summary))
	}
	return lines
}

func parseEcosystems(raw string) []Ecosystem {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]Ecosystem, 0, len(parts))
	seen := map[Ecosystem]bool{}
	for _, part := range parts {
		if ecosystem, ok := ParseEcosystem(part); ok && !seen[ecosystem] {
			result = append(result, ecosystem)
			seen[ecosystem] = true
		}
	}
	return result
}

func parseOptionalTools(raw string) []ToolName {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]ToolName, 0, len(parts))
	seen := map[ToolName]bool{}
	for _, part := range parts {
		if tool, ok := ParseToolName(part); ok && isOptionalTool(tool) && !seen[tool] {
			result = append(result, tool)
			seen[tool] = true
		}
	}
	return result
}
