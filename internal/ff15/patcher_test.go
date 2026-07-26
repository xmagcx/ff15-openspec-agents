package ff15

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyManagedPatchCreatesNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	if err := applyManagedPatch(path, "# Test\n\n- item"); err != nil {
		t.Fatalf("applyManagedPatch() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "<!-- FF15:MANAGED:START -->") {
		t.Fatalf("expected managed block in %s", text)
	}
}

func TestApplyManagedPatchReplacesManagedBlockOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	original := "# CLAUDE\n\nUser content\n\n<!-- FF15:MANAGED:START -->\nold\n<!-- FF15:MANAGED:END -->\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := applyManagedPatch(path, "new guidance"); err != nil {
		t.Fatalf("applyManagedPatch() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "User content") {
		t.Fatalf("expected unmanaged content preserved: %s", text)
	}
	if strings.Contains(text, "old") || !strings.Contains(text, "new guidance") {
		t.Fatalf("expected managed block replaced: %s", text)
	}
}

func TestExecutePlanSyncsEmbeddedAssetsForInit(t *testing.T) {
	targetDir := filepath.Join(t.TempDir(), ".claude", "agents")
	plan := Plan{
		Platform: PlatformLinux,
		Steps: []Step{{
			Kind:     StepSync,
			Title:    "Sync Claude agent assets",
			FilePath: targetDir,
		}},
	}
	if err := ExecutePlan(plan, io.Discard); err != nil {
		t.Fatalf("ExecutePlan() error = %v", err)
	}
	for _, path := range []string{
		filepath.Join(targetDir, "gladiolus.agent.md"),
		filepath.Join(targetDir, "ignis.agent.md"),
		filepath.Join(targetDir, "iris.agent.md"),
		filepath.Join(targetDir, "lunafreya.agent.md"),
		filepath.Join(targetDir, "noctis.agent.md"),
		filepath.Join(targetDir, "prompto.agent.md"),
		filepath.Join(targetDir, "noctis.prompt.md"),
		filepath.Join(targetDir, "development-policy.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected synced asset %s: %v", path, err)
		}
	}
}
