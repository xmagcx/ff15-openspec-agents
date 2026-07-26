package ff15

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	coverAssetPath  = "skin/chocobo.png"
	bannerAssetPath = "skin/ascii.txt"
)

type CoverInfo struct {
	Path            string
	Width           int
	Height          int
	Exists          bool
	TerminalPreview bool
	TerminalHint    string
	Preview         string
	PreviewMode     string
	Colorized       bool
}

type CoverRenderOptions struct {
	Width      int
	Colorized  bool
	TrueColor  bool
	Background color.RGBA
}

func DetectCoverInfo() CoverInfo {
	if path, ok := locateBannerAsset(); ok {
		info := CoverInfo{Path: filepath.ToSlash(path)}
		content, err := os.ReadFile(path)
		if err != nil {
			return info
		}
		lines := trimBannerLines(strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n"))
		if len(lines) == 0 {
			return info
		}
		info.Exists = true
		info.Height = len(lines)
		for _, line := range lines {
			if w := utf8.RuneCountInString(line); w > info.Width {
				info.Width = w
			}
		}
		info.Colorized, info.TerminalHint = detectTerminalColorSupport()
		info.PreviewMode = "ASCII banner"
		preview, err := RenderCoverPreview(path, CoverRenderOptions{Width: maxInt(72, info.Width+8)})
		if err == nil && strings.TrimSpace(preview) != "" {
			info.TerminalPreview = true
			info.Preview = preview
		}
		if !info.TerminalPreview {
			info.TerminalHint = "text fallback only"
		}
		return info
	}

	info := CoverInfo{Path: coverAssetPath}
	path, ok := locateCoverAsset()
	if !ok {
		return info
	}
	info.Path = filepath.ToSlash(path)
	file, err := os.Open(path)
	if err != nil {
		return info
	}
	defer file.Close()
	cfg, err := png.DecodeConfig(file)
	if err != nil {
		return info
	}
	info.Exists = true
	info.Width = cfg.Width
	info.Height = cfg.Height
	info.Colorized, info.TerminalHint = detectTerminalColorSupport()
	info.PreviewMode = "PNG metadata"
	preview, err := RenderCoverPreview(path, CoverRenderOptions{
		Width:      72,
		Colorized:  info.Colorized,
		TrueColor:  info.Colorized,
		Background: color.RGBA{R: 6, G: 9, B: 22, A: 255},
	})
	if err == nil && strings.TrimSpace(preview) != "" {
		info.TerminalPreview = true
		info.Preview = preview
		info.PreviewMode = "ANSI banner"
	} else {
		info.TerminalHint = "text fallback only"
		info.PreviewMode = "text fallback only"
	}
	return info
}

func (c CoverInfo) Summary() string {
	if !c.Exists {
		return bannerAssetPath + " unavailable"
	}
	status := c.PreviewMode
	if status == "" {
		status = "text fallback only"
	}
	return fmt.Sprintf("%s (%dx%d, %s)", c.Path, c.Width, c.Height, status)
}

func locateBannerAsset() (string, bool) {
	candidates := []string{bannerAssetPath}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, bannerAssetPath),
			filepath.Join(exeDir, "..", bannerAssetPath),
		)
	}
	return locateFirstFile(candidates)
}

func locateCoverAsset() (string, bool) {
	candidates := []string{coverAssetPath}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, coverAssetPath),
			filepath.Join(exeDir, "..", coverAssetPath),
		)
	}
	return locateFirstFile(candidates)
}

func locateFirstFile(candidates []string) (string, bool) {
	seen := map[string]bool{}
	for _, candidate := range candidates {
		cleaned := filepath.Clean(candidate)
		if seen[cleaned] {
			continue
		}
		seen[cleaned] = true
		if stat, err := os.Stat(cleaned); err == nil && !stat.IsDir() {
			return cleaned, true
		}
	}
	return "", false
}

func detectTerminalColorSupport() (bool, string) {
	if os.Getenv("NO_COLOR") != "" {
		return false, "NO_COLOR requested"
	}
	term := strings.ToLower(os.Getenv("TERM"))
	if term == "dumb" {
		return false, "TERM=dumb"
	}
	colorterm := strings.ToLower(os.Getenv("COLORTERM"))
	if strings.Contains(colorterm, "truecolor") || strings.Contains(colorterm, "24bit") {
		return true, "truecolor terminal detected"
	}
	if os.Getenv("WT_SESSION") != "" || os.Getenv("ANSICON") != "" || strings.EqualFold(os.Getenv("ConEmuANSI"), "ON") {
		return true, "ANSI-capable terminal detected"
	}
	if strings.Contains(term, "xterm") || strings.Contains(term, "screen") || strings.Contains(term, "tmux") || strings.Contains(term, "256color") || term == "linux" {
		return true, "ANSI-capable terminal detected"
	}
	return false, "text fallback only"
}

func RenderCoverPreview(path string, opts CoverRenderOptions) (string, error) {
	if strings.HasSuffix(strings.ToLower(path), ".txt") {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		width := opts.Width
		if width <= 0 {
			width = 72
		}
		return renderCenteredASCIIBanner(string(content), width), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		return "", err
	}
	width := opts.Width
	if width <= 0 {
		width = 48
	}
	height := maxInt(12, width/4)
	if opts.Colorized {
		return renderANSIBanner(img, width, height, opts.TrueColor, opts.Background), nil
	}
	return renderGlyphCover(img, width, height), nil
}

func renderCenteredASCIIBanner(content string, width int) string {
	lines := trimBannerLines(strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n"))
	if len(lines) == 0 {
		return ""
	}
	maxWidth := 0
	for _, line := range lines {
		if w := utf8.RuneCountInString(line); w > maxWidth {
			maxWidth = w
		}
	}
	if width < maxWidth {
		width = maxWidth
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		pad := (width - utf8.RuneCountInString(line)) / 2
		if pad < 0 {
			pad = 0
		}
		out = append(out, strings.Repeat(" ", pad)+line)
	}
	return strings.Join(out, "\n")
}

func trimBannerLines(lines []string) []string {
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return lines[start:end]
}

func renderANSIBanner(img image.Image, width, height int, trueColor bool, background color.RGBA) string {
	previewHeight := maxInt(8, height-2)
	previewWidth := maxInt(16, previewHeight)
	if previewWidth > width/2 {
		previewWidth = maxInt(12, width/2)
	}
	leftWidth := maxInt(4, (width-previewWidth)/2)
	rightWidth := maxInt(4, width-previewWidth-leftWidth)

	var b strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < leftWidth; x++ {
			b.WriteString(renderBannerGlyph(sampleAverageColor(img, leftWidth, height, x, y), background, trueColor))
		}
		for x := 0; x < previewWidth; x++ {
			pixel := sampleAverageColor(img, previewWidth, previewHeight, x, minInt(y, previewHeight-1))
			if pixel.A == 0 {
				pixel = background
			}
			if trueColor {
				b.WriteString(ansiBGTrueColor(pixel, "  "))
			} else {
				b.WriteString("██")
			}
		}
		for x := 0; x < rightWidth; x++ {
			b.WriteString(renderBannerGlyph(sampleAverageColor(img, rightWidth, height, maxInt(0, previewWidth+x-previewWidth/3), y), background, trueColor))
		}
		b.WriteString(ansiReset())
		if y+1 < height {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func renderBannerGlyph(pixel, background color.RGBA, trueColor bool) string {
	if pixel.A == 0 {
		pixel = background
	}
	glyph := shadeGlyph(relativeLuminance(pixel))
	if !trueColor {
		return glyph
	}
	return ansiFGTrueColor(pixel, glyph)
}

func renderGlyphCover(img image.Image, width, height int) string {
	const ramp = " .'`^,:;Il!i~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"
	var b strings.Builder
	for y := 0; y < height; y += 2 {
		for x := 0; x < width; x++ {
			top := sampleAverageColor(img, width, height, x, y)
			bottom := sampleAverageColor(img, width, height, x, y+1)
			luma := (relativeLuminance(top) + relativeLuminance(bottom)) / 2
			index := int(luma * float64(len(ramp)-1) / 255)
			if index < 0 {
				index = 0
			}
			if index >= len(ramp) {
				index = len(ramp) - 1
			}
			b.WriteByte(ramp[index])
		}
		if y+2 < height {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func sampleAverageColor(img image.Image, cellsWide, cellsHigh, cellX, cellY int) color.RGBA {
	bounds := img.Bounds()
	x0 := bounds.Min.X + (cellX*bounds.Dx())/cellsWide
	x1 := bounds.Min.X + ((cellX+1)*bounds.Dx())/cellsWide
	y0 := bounds.Min.Y + (cellY*bounds.Dy())/cellsHigh
	y1 := bounds.Min.Y + ((cellY+1)*bounds.Dy())/cellsHigh
	if x1 <= x0 {
		x1 = x0 + 1
	}
	if y1 <= y0 {
		y1 = y0 + 1
	}
	var sumR, sumG, sumB, sumA uint64
	var count uint64
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			r, g, b, a := img.At(px, py).RGBA()
			sumR += uint64(r >> 8)
			sumG += uint64(g >> 8)
			sumB += uint64(b >> 8)
			sumA += uint64(a >> 8)
			count++
		}
	}
	if count == 0 {
		return color.RGBA{}
	}
	return color.RGBA{R: uint8(sumR / count), G: uint8(sumG / count), B: uint8(sumB / count), A: uint8(sumA / count)}
}

func ansiBGTrueColor(bg color.RGBA, text string) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm%s", bg.R, bg.G, bg.B, text)
}

func ansiFGTrueColor(fg color.RGBA, text string) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", fg.R, fg.G, fg.B, text)
}

func ansiReset() string {
	return "\x1b[0m"
}

func relativeLuminance(c color.RGBA) float64 {
	return 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
}

func shadeGlyph(luma float64) string {
	ramp := []string{"  ", "..", "::", "**", "##", "@@"}
	index := int(luma * float64(len(ramp)-1) / 255)
	if index < 0 {
		index = 0
	}
	if index >= len(ramp) {
		index = len(ramp) - 1
	}
	return ramp[index]
}

func GenerateProjectCover(path string) error {
	const (
		width  = 128
		height = 128
		scale  = 8
	)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	fillRect(img, 0, 0, width, height, rgba(6, 9, 22))
	for y := 0; y < 68; y++ {
		mix := uint8(18 + y/3)
		fillRect(img, 0, y, width, 1, rgba(12, 14+mix/4, 28+mix/2))
	}
	fillRect(img, 0, 68, width, 14, rgba(18, 24, 44))
	fillRect(img, 0, 82, width, 46, rgba(11, 13, 24))

	for _, star := range [][2]int{{10, 10}, {22, 16}, {48, 12}, {71, 20}, {96, 9}, {108, 18}, {119, 13}} {
		fillRect(img, star[0], star[1], 2, 2, rgba(220, 228, 255))
	}
	fillRect(img, 90, 18, 14, 14, rgba(234, 237, 255))
	fillRect(img, 94, 22, 6, 6, rgba(245, 247, 255))

	for _, mountain := range []struct {
		x0, y0, x1, y1 int
		c              color.Color
	}{
		{0, 86, 44, 58, rgba(22, 27, 52)},
		{24, 88, 76, 50, rgba(28, 34, 62)},
		{58, 84, 104, 54, rgba(20, 24, 48)},
		{84, 90, 128, 60, rgba(16, 19, 38)},
	} {
		fillTriangle(img, mountain.x0, mountain.y0, mountain.x1, mountain.y1, mountain.c)
	}

	chocoboGold := rgba(228, 190, 78)
	chocoboShade := rgba(176, 138, 44)
	chocoboDark := rgba(95, 68, 20)
	accent := rgba(122, 170, 255)

	pixels := []string{
		"................................",
		"................................",
		"...............gg...............",
		".............gggggg.............",
		".............ggggggg............",
		"............gggggggg............",
		"...........gggggggggg...........",
		"...........gggggggggg...........",
		"..........gggggggggggg..........",
		"..........gggssssggggg..........",
		".........gggssssssggggg.........",
		".........ggssssssssgggg.........",
		".........ggsssssssssggg.........",
		".........gggssssssssggg.........",
		".........gggggggggggggg.........",
		"........gggggggggggggggg........",
		".......gggggggggggggggggg.......",
		".......gggggggddddggggggg.......",
		".......ggggggdddddggggggg.......",
		".......gggggdddddddgggggg.......",
		"........ggggdddddddggggg........",
		"........ggggdddddddggggg........",
		".........gggdddddddgggg.........",
		".........gggdddddddgggg.........",
		".........ggggdddddggggg.........",
		".........ggggggdggggggg.........",
		".........gggggggggggggg.........",
		"........gggggggggggggggg........",
		"........gggggg..gggggggg........",
		".......gggggg....gggggggg.......",
		".......ggggg......ggggggg.......",
		".......gggg........gggggg.......",
		".......ggg..........ggggg.......",
		".......ddd..........dddgg.......",
		".......ddd..........dddgg.......",
		".......ddd..........dddgg.......",
		"........d............dgg........",
		"........d............dgg........",
		".......aa............aag........",
		".......aa............aag........",
	}
	paintSprite(img, 44, 42, scale, pixels, map[rune]color.Color{
		'g': chocoboGold,
		's': chocoboShade,
		'd': chocoboDark,
		'a': accent,
	})
	fillRect(img, 84, 64, 4, 4, rgba(255, 250, 240))
	fillRect(img, 86, 66, 2, 2, rgba(15, 18, 32))

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func paintSprite(img *image.RGBA, offsetX, offsetY, scale int, rows []string, palette map[rune]color.Color) {
	for y, row := range rows {
		for x, pixel := range row {
			if pixel == '.' {
				continue
			}
			fillRect(img, offsetX+x*scale, offsetY+y*scale, scale, scale, palette[pixel])
		}
	}
}

func fillTriangle(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	for x := x0; x < x1; x++ {
		progress := float64(x-x0) / float64(maxInt(1, x1-x0))
		top := y0 - int(progress*float64(y0-y1))
		fillRect(img, x, top, 1, img.Bounds().Dy()-top, c)
	}
}

func fillRect(img *image.RGBA, x, y, width, height int, c color.Color) {
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			if image.Pt(px, py).In(img.Bounds()) {
				img.Set(px, py, c)
			}
		}
	}
}

func rgba(r, g, b uint8) color.Color {
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
