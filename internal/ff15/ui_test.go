package ff15

import (
	"strings"
	"testing"
)

func TestWizardViewIncludesCoverFallback(t *testing.T) {
	model := newWizardModel(Config{TargetRoot: "."}, PlatformLinux)
	model.cover = CoverInfo{Path: coverAssetPath, Exists: false}
	view := stripANSI(model.View())
	if view == "" {
		t.Fatalf("expected view output")
	}
	if !strings.Contains(view, "wizard uses text styling only") {
		t.Fatalf("expected cover fallback in view: %s", view)
	}
}

func TestWizardViewShowsRenderedCoverAndTheme(t *testing.T) {
	model := newWizardModel(Config{TargetRoot: "."}, PlatformLinux)
	model.cover = CoverInfo{
		Path:            coverAssetPath,
		Exists:          true,
		Width:           128,
		Height:          128,
		TerminalPreview: true,
		PreviewMode:     "ANSI truecolor preview",
		Preview:         "preview-line-1\npreview-line-2",
		Colorized:       true,
	}
	view := stripANSI(model.View())
	for _, want := range []string{"FF15 INIT WIZARD", "preview-line-1", "preview-line-2", "Journey", "Choose ecosystems", "FFXV-inspired install flow"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected themed wizard to contain %q, got:\n%s", want, view)
		}
	}
}
