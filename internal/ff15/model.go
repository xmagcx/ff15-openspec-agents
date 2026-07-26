package ff15

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
)

type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
)

func DetectPlatform(goos string) (Platform, error) {
	value := strings.TrimSpace(strings.ToLower(goos))
	if value == "" {
		value = runtime.GOOS
	}
	switch value {
	case "linux":
		return PlatformLinux, nil
	case "windows":
		return PlatformWindows, nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", value)
	}
}

type Ecosystem string

const (
	EcosystemClaude   Ecosystem = "claude"
	EcosystemKiro     Ecosystem = "kiro"
	EcosystemPi       Ecosystem = "pi"
	EcosystemOpenCode Ecosystem = "opencode"
)

func ParseEcosystem(raw string) (Ecosystem, bool) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "claude", "claude-code":
		return EcosystemClaude, true
	case "kiro":
		return EcosystemKiro, true
	case "pi", "pi-agents":
		return EcosystemPi, true
	case "opencode", "open-code":
		return EcosystemOpenCode, true
	default:
		return "", false
	}
}

type ToolName string

const (
	ToolEngram    ToolName = "engram"
	ToolCodeGraph ToolName = "codegraph"
	ToolOpenSpec  ToolName = "openspec"
	ToolHeadroom  ToolName = "headroom"
	ToolRTK       ToolName = "rtk"
	ToolSpek      ToolName = "spek"
	ToolDossier   ToolName = "dossier"
)

func ParseToolName(raw string) (ToolName, bool) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "engram":
		return ToolEngram, true
	case "codegraph":
		return ToolCodeGraph, true
	case "openspec":
		return ToolOpenSpec, true
	case "headroom":
		return ToolHeadroom, true
	case "rtk":
		return ToolRTK, true
	case "spek":
		return ToolSpek, true
	case "dossier":
		return ToolDossier, true
	default:
		return "", false
	}
}

func isOptionalTool(tool ToolName) bool {
	return tool == ToolHeadroom || tool == ToolRTK || tool == ToolSpek || tool == ToolDossier
}

type StepKind string

const (
	StepInstall StepKind = "install"
	StepSync    StepKind = "sync"
	StepPatch   StepKind = "patch"
)

type Step struct {
	Kind        StepKind
	Title       string
	Commands    []string
	Verify      string
	ManualHint  string
	FilePath    string
	ManagedText string
}

type Plan struct {
	Platform      Platform
	TargetRoot    string
	Ecosystems    []Ecosystem
	SelectedTools []ToolName
	Steps         []Step
}

func normalizeEcosystems(values []Ecosystem) []Ecosystem {
	seen := map[Ecosystem]bool{}
	result := make([]Ecosystem, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func normalizeTools(optional []ToolName) []ToolName {
	seen := map[ToolName]bool{}
	result := []ToolName{ToolEngram, ToolCodeGraph, ToolOpenSpec}
	for _, tool := range result {
		seen[tool] = true
	}
	for _, tool := range optional {
		if tool == "" || seen[tool] {
			continue
		}
		seen[tool] = true
		result = append(result, tool)
	}
	return result
}
