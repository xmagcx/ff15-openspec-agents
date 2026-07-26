package ff15

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedSyncAssetsAvailable(t *testing.T) {
	assets := syncAssetsFS()

	for _, path := range []string{
		"AGENTS.md",
		"agents/noctis.agent.md",
		"prompts/noctis.prompt.md",
		"docs/deployment-policy.md",
	} {
		content, err := fs.ReadFile(assets, path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		if len(content) == 0 {
			t.Fatalf("expected embedded asset %q to be non-empty", path)
		}
	}
}

func TestRunSyncList(t *testing.T) {
	var out bytes.Buffer

	if err := Run([]string{"sync", "--list"}, bytes.NewBuffer(nil), &out, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{"Available Agents:", "noctis", "Available Prompts:", "Available Docs:", "deployment-policy.md"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, text)
		}
	}
}

func TestRunSyncDryRunSelectiveAcrossSelectedEcosystems(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer

	err := Run([]string{"sync", "--target", target, "--ecosystems", "claude,kiro,pi,opencode", "--agents", "noctis,gladiolus", "--prompts", "noctis", "--docs", "development-policy.md", "--dry-run", "--force"}, bytes.NewBuffer(nil), &out, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	text := out.String()
	for _, want := range []string{
		"=== Syncing Claude assets (" + filepath.ToSlash(filepath.Join(target, ".claude", "agents")) + ") ===",
		"=== Syncing Kiro assets (" + filepath.ToSlash(filepath.Join(target, ".kiro", "steering")) + ") ===",
		"=== Syncing OpenCode assets (" + filepath.ToSlash(filepath.Join(target, ".opencode", "agents")) + ") ===",
		"[SYNC] gladiolus.agent.md -> " + filepath.Join(target, ".claude", "agents", "gladiolus.agent.md"),
		"[SYNC] noctis.prompt.md -> " + filepath.Join(target, ".kiro", "steering", "noctis.prompt.md"),
		"[SYNC] development-policy.md -> " + filepath.Join(target, ".opencode", "agents", "development-policy.md"),
		"[CREATE] " + filepath.Join(target, "CLAUDE.md") + " (with policy section)",
		"[CREATE] " + filepath.Join(target, ".kiro", "steering", "ff15-openspec-agents.md") + " (with policy section)",
		"[CREATE] " + filepath.Join(target, "AGENTS.md") + " (with policy section)",
		"[CREATE] " + filepath.Join(target, ".pi", "AGENTS.md") + " (with policy section)",
		"[CREATE] " + filepath.Join(target, ".opencode", "agents", "ff15-openspec-agents.md") + " (with policy section)",
		"=== Total files synced: 17 ===",
		"(Dry run - no files were actually modified)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, text)
		}
	}

	for _, path := range []string{
		filepath.Join(target, ".claude", "agents", "noctis.agent.md"),
		filepath.Join(target, ".kiro", "steering", "noctis.prompt.md"),
		filepath.Join(target, ".opencode", "agents", "development-policy.md"),
		filepath.Join(target, "AGENTS.md"),
		filepath.Join(target, "CLAUDE.md"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected dry run to leave %s absent, got err=%v", path, err)
		}
	}
}

func TestRunSyncForceUpdatesManagedBlocksAndEcosystemTargets(t *testing.T) {
	target := t.TempDir()
	for _, path := range []string{
		filepath.Join(target, "AGENTS.md"),
		filepath.Join(target, "CLAUDE.md"),
	} {
		original := strings.Join([]string{
			"# Team Notes",
			"",
			syncStartMarker,
			"old policy",
			syncEndMarker,
			"",
			"Keep this footer.",
		}, "\n")
		if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	var out bytes.Buffer
	err := Run([]string{"sync", "--target", target, "--ecosystems", "claude,pi,opencode", "--agents", "noctis", "--prompts", "noctis", "--docs", "development-policy.md", "--force"}, bytes.NewBuffer(nil), &out, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	for _, path := range []string{
		filepath.Join(target, ".claude", "agents", "noctis.agent.md"),
		filepath.Join(target, ".claude", "agents", "noctis.prompt.md"),
		filepath.Join(target, ".claude", "agents", "development-policy.md"),
		filepath.Join(target, ".opencode", "agents", "noctis.agent.md"),
		filepath.Join(target, ".opencode", "agents", "noctis.prompt.md"),
		filepath.Join(target, ".opencode", "agents", "development-policy.md"),
		filepath.Join(target, ".pi", "AGENTS.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected synced file %s, got err=%v", path, err)
		}
	}

	for _, path := range []string{filepath.Join(target, "AGENTS.md"), filepath.Join(target, "CLAUDE.md")} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		text := string(content)
		if strings.Contains(text, "old policy") {
			t.Fatalf("expected managed block to be replaced in %s, got:\n%s", path, text)
		}
		for _, want := range []string{"# Team Notes", "Keep this footer.", "docs/development-policy.md"} {
			if !strings.Contains(text, want) {
				t.Fatalf("expected %s to contain %q, got:\n%s", path, want, text)
			}
		}
	}
}

func TestRunSyncOnlyModesAndAppendWithoutForce(t *testing.T) {
	tests := []struct {
		name       string
		flag       string
		wantOnly   string
		wantAbsent string
		wantFile   string
	}{
		{name: "agents only", flag: "--agents-only", wantOnly: ".claude/agents", wantAbsent: "noctis.prompt.md", wantFile: filepath.Join(".claude", "agents", "gladiolus.agent.md")},
		{name: "prompts only", flag: "--prompts-only", wantOnly: ".claude/agents", wantAbsent: "deployment-policy.md", wantFile: filepath.Join(".claude", "agents", "noctis.prompt.md")},
		{name: "docs only", flag: "--docs-only", wantOnly: ".claude/agents", wantAbsent: "gladiolus.agent.md", wantFile: filepath.Join(".claude", "agents", "deployment-policy.md")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := t.TempDir()
			existing := filepath.Join(target, tt.wantFile)
			if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
				t.Fatalf("MkdirAll() error = %v", err)
			}
			if err := os.WriteFile(existing, []byte("keep me"), 0o644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			claudePath := filepath.Join(target, "CLAUDE.md")
			if err := os.WriteFile(claudePath, []byte("# Notes"), 0o644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}

			var out bytes.Buffer
			err := Run([]string{"sync", "--target", target, "--ecosystems", "claude", tt.flag}, bytes.NewBuffer(nil), &out, &out)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			text := out.String()
			if !strings.Contains(text, tt.wantOnly) {
				t.Fatalf("expected output to contain %q, got:\n%s", tt.wantOnly, text)
			}
			if strings.Contains(text, tt.wantAbsent) {
				t.Fatalf("did not expect output to contain %q, got:\n%s", tt.wantAbsent, text)
			}
			if !strings.Contains(text, "[SKIP] "+existing) {
				t.Fatalf("expected output to skip existing file, got:\n%s", text)
			}
			if !strings.Contains(text, "[APPEND] "+claudePath+" (add policy section at end)") {
				t.Fatalf("expected CLAUDE.md append output, got:\n%s", text)
			}

			content, err := os.ReadFile(existing)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}
			if string(content) != "keep me" {
				t.Fatalf("expected existing file to remain unchanged, got %q", string(content))
			}

			claudeContent, err := os.ReadFile(claudePath)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}
			if !strings.Contains(string(claudeContent), syncStartMarker) || !strings.Contains(string(claudeContent), "docs/review-policy.md") {
				t.Fatalf("expected CLAUDE.md append content, got:\n%s", string(claudeContent))
			}
		})
	}
}
