package ff15

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildPlanIncludesMandatoryToolsAndSelectedEcosystems(t *testing.T) {
	targetRoot := t.TempDir()
	plan, err := BuildPlan(targetRoot, PlatformLinux, []Ecosystem{EcosystemClaude, EcosystemPi, EcosystemOpenCode}, []ToolName{ToolRTK})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	wantTools := map[ToolName]bool{ToolEngram: true, ToolCodeGraph: true, ToolOpenSpec: true, ToolRTK: true}
	for _, tool := range plan.SelectedTools {
		delete(wantTools, tool)
	}
	if len(wantTools) != 0 {
		t.Fatalf("missing tools: %v", wantTools)
	}

	wantFiles := map[string]bool{
		filepath.Join(targetRoot, "CLAUDE.md"):                                      false,
		filepath.Join(targetRoot, "AGENTS.md"):                                      false,
		filepath.Join(targetRoot, ".pi", "AGENTS.md"):                               false,
		filepath.Join(targetRoot, ".opencode", "agents", "ff15-openspec-agents.md"): false,
	}
	for _, step := range plan.Steps {
		if step.Kind == StepPatch {
			if _, ok := wantFiles[step.FilePath]; ok {
				wantFiles[step.FilePath] = true
			}
		}
	}
	for path, seen := range wantFiles {
		if !seen {
			t.Fatalf("expected patch step for %s", path)
		}
	}
}

func TestInstallStepForToolUsesRealCommands(t *testing.T) {
	tests := []struct {
		name    string
		tool    ToolName
		wantSub string
		verify  string
	}{
		{name: "engram", tool: ToolEngram, wantSub: "go install github.com/Gentleman-Programming/engram/cmd/engram@latest", verify: "engram version"},
		{name: "codegraph", tool: ToolCodeGraph, wantSub: "npm install -g @colbymchenry/codegraph@latest", verify: "codegraph --version"},
		{name: "openspec", tool: ToolOpenSpec, wantSub: "npm install -g @fission-ai/openspec@latest", verify: "openspec --version"},
		{name: "rtk", tool: ToolRTK, wantSub: "cargo install --git https://github.com/rtk-ai/rtk rtk", verify: "rtk --help"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := installStepForTool(tt.tool, PlatformLinux)
			if len(step.Commands) == 0 || !strings.Contains(step.Commands[0], tt.wantSub) {
				t.Fatalf("expected command containing %q, got %v", tt.wantSub, step.Commands)
			}
			if step.Verify != tt.verify {
				t.Fatalf("expected verify %q, got %q", tt.verify, step.Verify)
			}
		})
	}
}

func TestSharedGuidanceIncludesRequestedProtocols(t *testing.T) {
	body := sharedGuidance([]ToolName{ToolEngram, ToolCodeGraph, ToolOpenSpec, ToolRTK})
	checks := []string{
		"# Engram Documentation & Protocols",
		"Artifact Storage Policy",
		"Sub-Agent Context Protocol",
		"CodeGraph",
		"RTK (Rust Token Killer)",
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected guidance to contain %q", check)
		}
	}
}
