package ff15

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

type wizardStage int

const (
	stageEcosystems wizardStage = iota
	stageOptionalTools
	stageReview
)

type optionItem struct {
	label    string
	detail   string
	selected bool
}

type WizardResult struct {
	Ecosystems    []Ecosystem
	OptionalTools []ToolName
	Approved      bool
	Cancelled     bool
}

type wizardModel struct {
	targetRoot string
	platform   Platform
	dryRun     bool
	cover      CoverInfo

	stage  wizardStage
	cursor int
	err    string

	ecosystemValues []Ecosystem
	ecosystemItems  []optionItem
	toolValues      []ToolName
	toolItems       []optionItem

	plan      Plan
	approved  bool
	cancelled bool
}

type colorRGB struct{ r, g, b uint8 }

type wizardTheme struct {
	ansi        bool
	panelBorder colorRGB
	title       colorRGB
	accent      colorRGB
	secondary   colorRGB
	gold        colorRGB
	error       colorRGB
	muted       colorRGB
}

func RunWizard(in io.Reader, out io.Writer, cfg Config, platform Platform) (WizardResult, error) {
	model := newWizardModel(cfg, platform)
	program := tea.NewProgram(model, tea.WithInput(in), tea.WithOutput(out), tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return WizardResult{}, err
	}
	resultModel, ok := finalModel.(wizardModel)
	if !ok {
		return WizardResult{}, fmt.Errorf("unexpected wizard model type %T", finalModel)
	}
	return WizardResult{
		Ecosystems:    resultModel.selectedEcosystems(),
		OptionalTools: resultModel.selectedOptionalTools(),
		Approved:      resultModel.approved,
		Cancelled:     resultModel.cancelled,
	}, nil
}

func newWizardModel(cfg Config, platform Platform) wizardModel {
	ecosystemValues := []Ecosystem{EcosystemClaude, EcosystemKiro, EcosystemPi, EcosystemOpenCode}
	ecosystemItems := []optionItem{
		{label: "Claude Code", detail: "Shared CLAUDE.md guidance and agent note."},
		{label: "Kiro", detail: "Steering file under .kiro/steering/."},
		{label: "Pi Agents", detail: "Root AGENTS.md plus .pi note."},
		{label: "OpenCode", detail: "Root AGENTS.md plus .opencode note."},
	}
	selectedEcosystems := map[Ecosystem]bool{}
	for _, ecosystem := range normalizeEcosystems(cfg.Ecosystems) {
		selectedEcosystems[ecosystem] = true
	}
	if len(selectedEcosystems) == 0 {
		for _, ecosystem := range ecosystemValues {
			selectedEcosystems[ecosystem] = true
		}
	}
	for i, ecosystem := range ecosystemValues {
		ecosystemItems[i].selected = selectedEcosystems[ecosystem]
	}

	toolValues := []ToolName{ToolHeadroom, ToolRTK, ToolSpek, ToolDossier}
	toolItems := []optionItem{
		{label: "Headroom", detail: "Optional Python-powered AI workflow helper."},
		{label: "RTK", detail: "Optional Cargo-installed command prefix helper."},
		{label: "Spek", detail: "Optional OpenSpec web viewer scaffolded locally under .ff15/tools/."},
		{label: "Dossier", detail: "Optional OpenSpec terminal viewer installed with Go."},
	}
	selectedTools := map[ToolName]bool{}
	for _, tool := range normalizeTools(cfg.OptionalTools) {
		if isOptionalTool(tool) {
			selectedTools[tool] = true
		}
	}
	for i, tool := range toolValues {
		toolItems[i].selected = selectedTools[tool]
	}

	return wizardModel{
		targetRoot:      cfg.TargetRoot,
		platform:        platform,
		dryRun:          cfg.DryRun,
		cover:           DetectCoverInfo(),
		stage:           stageEcosystems,
		ecosystemValues: ecosystemValues,
		ecosystemItems:  ecosystemItems,
		toolValues:      toolValues,
		toolItems:       toolItems,
	}
}

func (m wizardModel) Init() tea.Cmd {
	return nil
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case " ", "x":
			m.toggleCurrent()
		case "a":
			if m.stage == stageEcosystems {
				for i := range m.ecosystemItems {
					m.ecosystemItems[i].selected = true
				}
				m.err = ""
			}
		case "backspace", "esc":
			if m.stage == stageOptionalTools {
				m.stage = stageEcosystems
				m.cursor = 0
				m.err = ""
			} else if m.stage == stageReview {
				m.stage = stageOptionalTools
				m.cursor = 0
				m.err = ""
			}
		case "enter":
			return m.submit()
		case "y":
			if m.stage == stageReview {
				m.approved = true
				return m, tea.Quit
			}
		case "n":
			if m.stage == stageReview {
				m.cancelled = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m wizardModel) View() string {
	theme := newWizardTheme(m.cover.Colorized)
	sections := []string{
		renderHeroPanel(m, theme),
		renderStagePanel(m, theme),
	}
	if m.err != "" {
		sections = append(sections, renderPanel("Warning", []string{theme.paint("! "+m.err, theme.error, false)}, theme))
	}

	switch m.stage {
	case stageEcosystems:
		sections = append(sections,
			renderPanel("Choose ecosystems", renderOptions(m.ecosystemItems, m.cursor, theme), theme),
			renderPanel("Controls", []string{renderControls(theme, "↑/↓ move", "space toggle", "a select all", "enter continue", "q quit")}, theme),
		)
	case stageOptionalTools:
		sections = append(sections,
			renderPanel("Optional tools", renderOptions(m.toolItems, m.cursor, theme), theme),
			renderPanel("Controls", []string{renderControls(theme, "↑/↓ move", "space toggle", "enter review", "esc back", "q quit")}, theme),
		)
	case stageReview:
		action := "Apply plan"
		if m.dryRun {
			action = "Finish dry run"
		}
		sections = append(sections,
			renderPanel("Review crystal", renderReview(m, theme), theme),
			renderPanel("Controls", []string{renderControls(theme, "y/enter "+action, "esc back", "n/q cancel")}, theme),
		)
	}

	return strings.Join(sections, "\n\n") + "\n"
}

func (m *wizardModel) moveCursor(delta int) {
	length := len(m.currentItems())
	if length == 0 {
		m.cursor = 0
		return
	}
	m.cursor = (m.cursor + delta + length) % length
}

func (m *wizardModel) toggleCurrent() {
	if m.stage == stageReview {
		return
	}
	switch m.stage {
	case stageEcosystems:
		if len(m.ecosystemItems) > 0 {
			m.ecosystemItems[m.cursor].selected = !m.ecosystemItems[m.cursor].selected
		}
	case stageOptionalTools:
		if len(m.toolItems) > 0 {
			m.toolItems[m.cursor].selected = !m.toolItems[m.cursor].selected
		}
	}
	m.err = ""
}

func (m wizardModel) submit() (tea.Model, tea.Cmd) {
	switch m.stage {
	case stageEcosystems:
		if len(m.selectedEcosystems()) == 0 {
			m.err = "Select at least one ecosystem."
			return m, nil
		}
		m.stage = stageOptionalTools
		m.cursor = 0
		m.err = ""
		return m, nil
	case stageOptionalTools:
		plan, err := BuildPlan(m.targetRoot, m.platform, m.selectedEcosystems(), m.selectedOptionalTools())
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.plan = plan
		m.stage = stageReview
		m.cursor = 0
		m.err = ""
		return m, nil
	case stageReview:
		m.approved = true
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m *wizardModel) currentItems() []optionItem {
	switch m.stage {
	case stageEcosystems:
		return m.ecosystemItems
	case stageOptionalTools:
		return m.toolItems
	default:
		return nil
	}
}

func (m wizardModel) selectedEcosystems() []Ecosystem {
	result := make([]Ecosystem, 0, len(m.ecosystemItems))
	for i, item := range m.ecosystemItems {
		if item.selected {
			result = append(result, m.ecosystemValues[i])
		}
	}
	return normalizeEcosystems(result)
}

func (m wizardModel) selectedOptionalTools() []ToolName {
	result := make([]ToolName, 0, len(m.toolItems))
	for i, item := range m.toolItems {
		if item.selected {
			result = append(result, m.toolValues[i])
		}
	}
	return result
}

func (m wizardModel) stageTitle() string {
	switch m.stage {
	case stageEcosystems:
		return "Step 1/3 · Ecosystems"
	case stageOptionalTools:
		return "Step 2/3 · Optional tools"
	case stageReview:
		return "Step 3/3 · Review"
	default:
		return "Wizard"
	}
}

func newWizardTheme(ansi bool) wizardTheme {
	return wizardTheme{
		ansi:        ansi,
		panelBorder: colorRGB{88, 104, 163},
		title:       colorRGB{215, 230, 255},
		accent:      colorRGB{125, 175, 255},
		secondary:   colorRGB{154, 194, 255},
		gold:        colorRGB{236, 198, 92},
		error:       colorRGB{255, 128, 128},
		muted:       colorRGB{123, 132, 165},
	}
}

func (t wizardTheme) paint(text string, fg colorRGB, bold bool) string {
	if !t.ansi || text == "" {
		return text
	}
	prefix := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", fg.r, fg.g, fg.b)
	if bold {
		prefix += "\x1b[1m"
	}
	return prefix + text + "\x1b[0m"
}

func renderHeroPanel(m wizardModel, theme wizardTheme) string {
	meta := []string{
		theme.paint("FFXV-inspired install flow", theme.gold, true),
		fmt.Sprintf("Target   %s", m.targetRoot),
		fmt.Sprintf("Platform %s", m.platform),
	}
	if m.cover.TerminalPreview {
		meta = append(meta, theme.paint("Banner   rendered from skin/chocobo.png", theme.muted, false))
		body := append(append(strings.Split(m.cover.Preview, "\n"), ""), meta...)
		return renderPanel("FF15 INIT WIZARD", body, theme)
	}
	meta = append(meta, "Banner   skin/chocobo.png unavailable; wizard uses text styling only.")
	return renderPanel("FF15 INIT WIZARD", meta, theme)
}

func renderStagePanel(m wizardModel, theme wizardTheme) string {
	line := theme.paint(m.stageTitle(), theme.title, true)
	subtitle := "Select what joins the caravan."
	switch m.stage {
	case stageOptionalTools:
		subtitle = "Optional companions can travel with the core toolchain."
	case stageReview:
		subtitle = "Review the route before crystals are committed."
	}
	return renderPanel("Journey", []string{line, theme.paint(subtitle, theme.muted, false)}, theme)
}

func renderOptions(items []optionItem, cursor int, theme wizardTheme) []string {
	lines := make([]string, 0, len(items))
	for i, item := range items {
		pointer := "  "
		if i == cursor {
			pointer = theme.paint("❯ ", theme.gold, true)
		}
		check := theme.paint("○", theme.muted, false)
		if item.selected {
			check = theme.paint("◉", theme.accent, true)
		}
		label := item.label
		if item.selected {
			label = theme.paint(item.label, theme.title, true)
		}
		detail := theme.paint(item.detail, theme.muted, false)
		lines = append(lines, fmt.Sprintf("%s%s %s", pointer, check, label))
		lines = append(lines, fmt.Sprintf("   %s", detail))
	}
	return lines
}

func renderReview(m wizardModel, theme wizardTheme) []string {
	lines := []string{
		theme.paint(fmt.Sprintf("Ecosystems    %s", joinEcosystems(m.selectedEcosystems())), theme.title, true),
		theme.paint(fmt.Sprintf("Optional tools %s", joinTools(m.selectedOptionalTools())), theme.secondary, false),
		theme.paint(fmt.Sprintf("Plan          %d steps · %d installs · %d managed patches", len(m.plan.Steps), countSteps(m.plan, StepInstall), countSteps(m.plan, StepPatch)), theme.gold, false),
		"",
	}
	for _, line := range renderPlanSummaryLines(m.plan) {
		lines = append(lines, "• "+line)
	}
	return lines
}

func renderControls(theme wizardTheme, controls ...string) string {
	parts := make([]string, 0, len(controls))
	for _, control := range controls {
		parts = append(parts, theme.paint(control, theme.accent, false))
	}
	return strings.Join(parts, theme.paint("  •  ", theme.muted, false))
}

func renderPanel(title string, body []string, theme wizardTheme) string {
	width := utf8.RuneCountInString(stripANSI(title))
	for _, line := range body {
		if l := utf8.RuneCountInString(stripANSI(line)); l > width {
			width = l
		}
	}
	if width < 24 {
		width = 24
	}
	top := "╭─ " + padRight(title, width) + " ─╮"
	mid := make([]string, 0, len(body))
	for _, line := range body {
		mid = append(mid, fmt.Sprintf("│ %s │", padRightANSI(line, width)))
	}
	bottom := "╰─" + strings.Repeat("─", width+2) + "╯"
	lines := append([]string{theme.paint(top, theme.panelBorder, false)}, mid...)
	lines = append(lines, theme.paint(bottom, theme.panelBorder, false))
	return strings.Join(lines, "\n")
}

func padRightANSI(value string, width int) string {
	visible := utf8.RuneCountInString(stripANSI(value))
	if visible >= width {
		return value
	}
	return value + strings.Repeat(" ", width-visible)
}

func padRight(value string, width int) string {
	visible := utf8.RuneCountInString(value)
	if visible >= width {
		return value
	}
	return value + strings.Repeat(" ", width-visible)
}

func stripANSI(value string) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] == 0x1b {
			i++
			for i < len(value) && value[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteByte(value[i])
	}
	return b.String()
}

func joinEcosystems(values []Ecosystem) string {
	if len(values) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, string(value))
	}
	return strings.Join(parts, ", ")
}

func joinTools(values []ToolName) string {
	if len(values) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, string(value))
	}
	return strings.Join(parts, ", ")
}

func countSteps(plan Plan, kind StepKind) int {
	count := 0
	for _, step := range plan.Steps {
		if step.Kind == kind {
			count++
		}
	}
	return count
}
