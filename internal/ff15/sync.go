package ff15

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	syncStartMarker = "<!-- SOFTWARE DEVELOPMENT POLICIES:START -->"
	syncEndMarker   = "<!-- SOFTWARE DEVELOPMENT POLICIES:END -->"
)

type SyncConfig struct {
	TargetRoot      string
	Ecosystems      []Ecosystem
	SelectedAgents  []string
	SelectedPrompts []string
	SelectedDocs    []string
	AgentsOnly      bool
	PromptsOnly     bool
	DocsOnly        bool
	Force           bool
	DryRun          bool
	List            bool
}

func runSync(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("ff15 sync", flag.ContinueOnError)
	fs.SetOutput(stdout)
	target := fs.String("target", "", "target project directory")
	ecosystemsCSV := fs.String("ecosystems", "", "comma-separated ecosystems: claude,kiro,pi,opencode (default: all supported ecosystems)")
	agentsCSV := fs.String("agents", "", "comma-separated list of agents to sync")
	promptsCSV := fs.String("prompts", "", "comma-separated list of prompts to sync")
	docsCSV := fs.String("docs", "", "comma-separated list of docs to sync")
	agentsOnly := fs.Bool("agents-only", false, "sync only agents and managed guidance files")
	promptsOnly := fs.Bool("prompts-only", false, "sync only prompts and managed guidance files")
	docsOnly := fs.Bool("docs-only", false, "sync only docs and managed guidance files")
	force := fs.Bool("force", false, "overwrite existing files without confirmation")
	dryRun := fs.Bool("dry-run", false, "show what would be synced without making changes")
	list := fs.Bool("list", false, "list available agents, prompts, and docs")
	fs.Usage = func() {
		printUsage(stdout)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Sync Flags:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	return executeSync(SyncConfig{
		TargetRoot:      *target,
		Ecosystems:      parseEcosystems(*ecosystemsCSV),
		SelectedAgents:  parseCSV(*agentsCSV),
		SelectedPrompts: parseCSV(*promptsCSV),
		SelectedDocs:    parseCSV(*docsCSV),
		AgentsOnly:      *agentsOnly,
		PromptsOnly:     *promptsOnly,
		DocsOnly:        *docsOnly,
		Force:           *force,
		DryRun:          *dryRun,
		List:            *list,
	}, stdin, stdout)
}

func executeSync(cfg SyncConfig, stdin io.Reader, stdout io.Writer) error {
	assets := syncAssetsFS()

	if cfg.List {
		listSyncItems(assets, stdout)
		return nil
	}

	if strings.TrimSpace(cfg.TargetRoot) == "" {
		return fmt.Errorf("--target is required for sync operations")
	}

	targetRoot := filepath.Clean(cfg.TargetRoot)
	stat, err := os.Stat(targetRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target directory does not exist: %s", targetRoot)
		}
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("target path is not a directory: %s", targetRoot)
	}

	overwrite := &overwritePrompter{stdin: stdin, stdout: stdout}
	totalSynced := 0
	ecosystems := defaultSyncEcosystems(cfg.Ecosystems)

	for _, ecosystem := range ecosystems {
		assetDir := ecosystemAssetDir(targetRoot, ecosystem)
		if assetDir == "" {
			continue
		}

		fmt.Fprintf(stdout, "\n=== Syncing %s assets (%s) ===\n", ecosystemDisplayName(ecosystem), filepath.ToSlash(assetDir))
		ecosystemSynced := 0

		if !cfg.PromptsOnly && !cfg.DocsOnly {
			synced, err := syncFiles(syncFilesConfig{
				SourceFS:  assets,
				SourceDir: "agents",
				TargetDir: assetDir,
				Pattern:   ".agent.md",
				Selected:  cfg.SelectedAgents,
				Force:     cfg.Force,
				DryRun:    cfg.DryRun,
				Prompt:    overwrite,
				Stdout:    stdout,
			})
			if err != nil {
				return err
			}
			ecosystemSynced += synced
		}

		if !cfg.AgentsOnly && !cfg.DocsOnly {
			synced, err := syncFiles(syncFilesConfig{
				SourceFS:  assets,
				SourceDir: "prompts",
				TargetDir: assetDir,
				Pattern:   ".prompt.md",
				Selected:  cfg.SelectedPrompts,
				Force:     cfg.Force,
				DryRun:    cfg.DryRun,
				Prompt:    overwrite,
				Stdout:    stdout,
			})
			if err != nil {
				return err
			}
			ecosystemSynced += synced
		}

		if !cfg.AgentsOnly && !cfg.PromptsOnly {
			synced, err := syncFiles(syncFilesConfig{
				SourceFS:  assets,
				SourceDir: "docs",
				TargetDir: assetDir,
				Pattern:   ".md",
				Selected:  cfg.SelectedDocs,
				Force:     cfg.Force,
				DryRun:    cfg.DryRun,
				Prompt:    overwrite,
				Stdout:    stdout,
			})
			if err != nil {
				return err
			}
			ecosystemSynced += synced
		}

		fmt.Fprintf(stdout, "%s assets synced: %d\n", ecosystemDisplayName(ecosystem), ecosystemSynced)
		totalSynced += ecosystemSynced
	}

	fmt.Fprintln(stdout, "\n=== Syncing managed guidance files ===")
	managedContent, err := fs.ReadFile(assets, "AGENTS.md")
	if err != nil {
		return err
	}
	managedFiles := map[string]bool{}
	for _, ecosystem := range ecosystems {
		for _, targetFile := range ecosystemManagedPaths(targetRoot, ecosystem) {
			if managedFiles[targetFile] {
				continue
			}
			managedFiles[targetFile] = true
			synced, err := syncManagedMarkdown(targetFile, string(managedContent), cfg.DryRun, stdout)
			if err != nil {
				return err
			}
			if synced {
				totalSynced++
			}
		}
	}

	fmt.Fprintf(stdout, "\n=== Total files synced: %d ===\n", totalSynced)
	if cfg.DryRun {
		fmt.Fprintln(stdout, "\n(Dry run - no files were actually modified)")
	}
	return nil
}

type overwritePrompter struct {
	stdin  io.Reader
	stdout io.Writer
}

func (o *overwritePrompter) Confirm(target string) (bool, error) {
	if o.stdin == nil {
		return false, nil
	}
	fmt.Fprintf(o.stdout, "Overwrite %s? [y/N] ", target)
	reader := bufio.NewReader(o.stdin)
	response, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(response), "y"), nil
}

type syncFilesConfig struct {
	SourceFS  fs.FS
	SourceDir string
	TargetDir string
	Pattern   string
	Selected  []string
	Force     bool
	DryRun    bool
	Prompt    interface{ Confirm(string) (bool, error) }
	Stdout    io.Writer
}

func syncFiles(cfg syncFilesConfig) (int, error) {
	entries, err := fs.ReadDir(cfg.SourceFS, cfg.SourceDir)
	if err != nil {
		return 0, err
	}

	selected := make(map[string]bool, len(cfg.Selected))
	for _, name := range cfg.Selected {
		selected[name] = true
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, cfg.Pattern) {
			continue
		}
		stem := strings.TrimSuffix(name, cfg.Pattern)
		if len(selected) > 0 && !selected[stem] && !selected[name] {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)

	synced := 0
	for _, name := range files {
		sourceFile := filepath.ToSlash(filepath.Join(cfg.SourceDir, name))
		targetFile := filepath.Join(cfg.TargetDir, name)

		if info, err := os.Stat(targetFile); err == nil && !info.IsDir() && !cfg.Force {
			if cfg.DryRun {
				fmt.Fprintf(cfg.Stdout, "[SKIP] %s (exists, use --force to overwrite)\n", targetFile)
				continue
			}
			ok, err := cfg.Prompt.Confirm(targetFile)
			if err != nil {
				return synced, err
			}
			if !ok {
				fmt.Fprintf(cfg.Stdout, "[SKIP] %s\n", targetFile)
				continue
			}
		} else if err != nil && !os.IsNotExist(err) {
			return synced, err
		}

		if cfg.DryRun {
			fmt.Fprintf(cfg.Stdout, "[SYNC] %s -> %s\n", name, targetFile)
			synced++
			continue
		}

		if err := copyFile(cfg.SourceFS, sourceFile, targetFile); err != nil {
			return synced, err
		}
		fmt.Fprintf(cfg.Stdout, "[SYNC] %s -> %s\n", name, targetFile)
		synced++
	}

	return synced, nil
}

func copyFile(sourceFS fs.FS, sourceFile, targetFile string) error {
	content, err := fs.ReadFile(sourceFS, sourceFile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(targetFile, content, 0o644)
}

func syncManagedMarkdown(targetFile, sourceText string, dryRun bool, stdout io.Writer) (bool, error) {
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		if dryRun {
			fmt.Fprintf(stdout, "[CREATE] %s (with policy section)\n", targetFile)
			return true, nil
		}
		if err := os.MkdirAll(filepath.Dir(targetFile), 0o755); err != nil {
			return false, err
		}
		template := fmt.Sprintf("# Guidelines\n\n%s\n%s\n%s\n", syncStartMarker, sourceText, syncEndMarker)
		if err := os.WriteFile(targetFile, []byte(template), 0o644); err != nil {
			return false, err
		}
		fmt.Fprintf(stdout, "[CREATE] %s (with policy section)\n", targetFile)
		return true, nil
	} else if err != nil {
		return false, err
	}

	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		return false, err
	}
	targetText := string(targetContent)
	startIndex := strings.Index(targetText, syncStartMarker)
	endIndex := strings.Index(targetText, syncEndMarker)
	if startIndex >= 0 && endIndex > startIndex {
		if dryRun {
			fmt.Fprintf(stdout, "[UPDATE] %s (between markers)\n", targetFile)
			return true, nil
		}
		replacement := syncStartMarker + "\n" + sourceText + "\n" + syncEndMarker
		newText := targetText[:startIndex] + replacement + targetText[endIndex+len(syncEndMarker):]
		if err := os.WriteFile(targetFile, []byte(newText), 0o644); err != nil {
			return false, err
		}
		fmt.Fprintf(stdout, "[UPDATE] %s (between markers)\n", targetFile)
		return true, nil
	}

	if dryRun {
		fmt.Fprintf(stdout, "[APPEND] %s (add policy section at end)\n", targetFile)
		return true, nil
	}
	if !strings.HasSuffix(targetText, "\n") {
		targetText += "\n"
	}
	targetText += "\n" + syncStartMarker + "\n" + sourceText + "\n" + syncEndMarker + "\n"
	if err := os.WriteFile(targetFile, []byte(targetText), 0o644); err != nil {
		return false, err
	}
	fmt.Fprintf(stdout, "[APPEND] %s (add policy section at end)\n", targetFile)
	return true, nil
}

func listSyncItems(sourceFS fs.FS, stdout io.Writer) {
	fmt.Fprintln(stdout, "Available Agents:")
	fmt.Fprintln(stdout, "----------------------------------------")
	for _, agent := range availableSyncItems(sourceFS, "agents", ".agent.md", false) {
		fmt.Fprintf(stdout, "  - %s\n", agent)
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Available Prompts:")
	fmt.Fprintln(stdout, "----------------------------------------")
	for _, prompt := range availableSyncItems(sourceFS, "prompts", ".prompt.md", false) {
		fmt.Fprintf(stdout, "  - %s\n", prompt)
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Available Docs:")
	fmt.Fprintln(stdout, "----------------------------------------")
	for _, doc := range availableSyncItems(sourceFS, "docs", ".md", true) {
		fmt.Fprintf(stdout, "  - %s\n", doc)
	}
}

func availableSyncItems(sourceFS fs.FS, dir, suffix string, keepName bool) []string {
	entries, err := fs.ReadDir(sourceFS, dir)
	if err != nil {
		return nil
	}
	items := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		if keepName {
			items = append(items, name)
			continue
		}
		items = append(items, strings.TrimSuffix(name, suffix))
	}
	sort.Strings(items)
	return items
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}
	return result
}
