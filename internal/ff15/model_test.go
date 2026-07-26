package ff15

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Platform
		wantErr bool
	}{
		{name: "linux", input: "linux", want: PlatformLinux},
		{name: "windows", input: "windows", want: PlatformWindows},
		{name: "unsupported", input: "darwin", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectPlatform(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("DetectPlatform() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("DetectPlatform() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWizardModelRequiresOneEcosystem(t *testing.T) {
	model := newWizardModel(Config{TargetRoot: "."}, PlatformLinux)
	for i := range model.ecosystemItems {
		model.ecosystemItems[i].selected = false
	}

	updated, _ := model.Update(testKey("enter"))
	got := updated.(wizardModel)
	if got.stage != stageEcosystems {
		t.Fatalf("stage = %v, want ecosystems", got.stage)
	}
	if got.err == "" {
		t.Fatalf("expected validation error")
	}
}

func TestWizardModelBuildsReviewPlan(t *testing.T) {
	model := newWizardModel(Config{TargetRoot: ".", Ecosystems: []Ecosystem{EcosystemClaude}, OptionalTools: []ToolName{ToolRTK}}, PlatformLinux)
	model.stage = stageOptionalTools

	updated, _ := model.Update(testKey("enter"))
	got := updated.(wizardModel)
	if got.stage != stageReview {
		t.Fatalf("stage = %v, want review", got.stage)
	}
	if len(got.plan.Steps) == 0 {
		t.Fatalf("expected plan steps")
	}
	if !containsTool(got.plan.SelectedTools, ToolRTK) {
		t.Fatalf("expected RTK in selected tools: %v", got.plan.SelectedTools)
	}
}

func TestWizardModelCanApproveFromReview(t *testing.T) {
	model := newWizardModel(Config{TargetRoot: ".", Ecosystems: []Ecosystem{EcosystemClaude}}, PlatformLinux)
	model.stage = stageReview
	model.plan = Plan{TargetRoot: ".", Platform: PlatformLinux, Ecosystems: []Ecosystem{EcosystemClaude}, SelectedTools: normalizeTools(nil)}

	updated, cmd := model.Update(testKey("y"))
	got := updated.(wizardModel)
	if !got.approved {
		t.Fatalf("expected approval")
	}
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
}

func testKey(value string) tea.KeyMsg {
	runes := []rune(value)
	if value == "enter" {
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
}

func containsTool(tools []ToolName, want ToolName) bool {
	for _, tool := range tools {
		if tool == want {
			return true
		}
	}
	return false
}
