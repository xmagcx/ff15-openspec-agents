package ff15

import (
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
