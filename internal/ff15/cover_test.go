package ff15

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateProjectCover(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chocobo.png")
	if err := GenerateProjectCover(path); err != nil {
		t.Fatalf("GenerateProjectCover() error = %v", err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer file.Close()
	cfg, err := png.DecodeConfig(file)
	if err != nil {
		t.Fatalf("DecodeConfig() error = %v", err)
	}
	if cfg.Width != 128 || cfg.Height != 128 {
		t.Fatalf("cover size = %dx%d, want 128x128", cfg.Width, cfg.Height)
	}
}

func TestRenderCoverPreviewSupportsANSIAndGlyphFallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chocobo.png")
	if err := GenerateProjectCover(path); err != nil {
		t.Fatalf("GenerateProjectCover() error = %v", err)
	}

	tests := []struct {
		name      string
		colorized bool
		want      string
		notWant   string
	}{
		{name: "ansi", colorized: true, want: "\x1b[48;2;", notWant: "text fallback only"},
		{name: "glyph", colorized: false, want: "\n", notWant: "\x1b[48;2;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview, err := RenderCoverPreview(path, CoverRenderOptions{Width: 24, Colorized: tt.colorized, TrueColor: tt.colorized})
			if err != nil {
				t.Fatalf("RenderCoverPreview() error = %v", err)
			}
			if !strings.Contains(preview, tt.want) {
				t.Fatalf("expected preview to contain %q, got %q", tt.want, preview)
			}
			if strings.Contains(preview, tt.notWant) {
				t.Fatalf("expected preview not to contain %q, got %q", tt.notWant, preview)
			}
		})
	}
}

func TestRenderCoverPreviewCentersASCIIBanner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ascii.txt")
	content := "CHOCOBO\nBANNER"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	preview, err := RenderCoverPreview(path, CoverRenderOptions{Width: 20})
	if err != nil {
		t.Fatalf("RenderCoverPreview() error = %v", err)
	}
	if !strings.Contains(preview, "      CHOCOBO") {
		t.Fatalf("expected centered banner, got %q", preview)
	}
	if !strings.Contains(preview, "       BANNER") {
		t.Fatalf("expected centered banner second line, got %q", preview)
	}
}
