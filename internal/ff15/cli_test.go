package ff15

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStartsInitByDefault(t *testing.T) {
	var out bytes.Buffer

	err := Run(nil, bytes.NewBuffer(nil), &out, &out)
	if err == nil {
		t.Fatalf("expected error for non-interactive default run")
	}
	if !strings.Contains(err.Error(), "interactive mode requires a terminal") {
		t.Fatalf("expected interactive mode error, got %v", err)
	}
	if strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected init flow attempt instead of usage output, got:\n%s", out.String())
	}
}

func TestRunHelpPrintsCommandsAndFlags(t *testing.T) {
	var out bytes.Buffer

	if err := Run([]string{"--help"}, bytes.NewBuffer(nil), &out, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{"Usage:", "ff15 [flags]", "ff15 sync [flags]", "Commands:", "init", "sync", "Flags:", "-target"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected help output to contain %q, got:\n%s", want, text)
		}
	}
}

func TestRunInitFlagStillWorks(t *testing.T) {
	var out bytes.Buffer

	err := Run([]string{"--init", "--dry-run", "--yes", "--ecosystems", "claude,pi", "--optional-tools", "rtk"}, bytes.NewBuffer(nil), &out, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{"FF15 init plan", "Dry run complete. No changes applied.", "[install]", "[patch]"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, text)
		}
	}
}

func TestExecuteInitDryRunWithFlags(t *testing.T) {
	var out bytes.Buffer
	cfg := Config{
		TargetRoot:    filepath.Join("project"),
		DryRun:        true,
		AutoApprove:   true,
		Ecosystems:    []Ecosystem{EcosystemClaude, EcosystemPi},
		OptionalTools: []ToolName{ToolRTK},
	}
	if err := executeInit(cfg, bytes.NewBuffer(nil), &out); err != nil {
		t.Fatalf("executeInit() error = %v", err)
	}
	text := out.String()
	for _, want := range []string{"FF15 init plan", "Dry run complete. No changes applied.", "[install]", "[patch]"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, text)
		}
	}
}
