package ff15

import "path/filepath"

func allSupportedEcosystems() []Ecosystem {
	return []Ecosystem{EcosystemClaude, EcosystemKiro, EcosystemPi, EcosystemOpenCode}
}

func defaultSyncEcosystems(selected []Ecosystem) []Ecosystem {
	normalized := normalizeEcosystems(selected)
	if len(normalized) != 0 {
		return normalized
	}
	return allSupportedEcosystems()
}

func ecosystemManagedPaths(targetRoot string, ecosystem Ecosystem) []string {
	switch ecosystem {
	case EcosystemClaude:
		return []string{filepath.Join(targetRoot, "CLAUDE.md")}
	case EcosystemKiro:
		return []string{filepath.Join(targetRoot, ".kiro", "steering", "ff15-openspec-agents.md")}
	case EcosystemPi:
		return []string{filepath.Join(targetRoot, "AGENTS.md"), filepath.Join(targetRoot, ".pi", "AGENTS.md")}
	case EcosystemOpenCode:
		return []string{filepath.Join(targetRoot, "AGENTS.md"), filepath.Join(targetRoot, ".opencode", "agents", "ff15-openspec-agents.md")}
	default:
		return nil
	}
}

func ecosystemAssetDir(targetRoot string, ecosystem Ecosystem) string {
	switch ecosystem {
	case EcosystemClaude:
		return filepath.Join(targetRoot, ".claude", "agents")
	case EcosystemKiro:
		return filepath.Join(targetRoot, ".kiro", "steering")
	case EcosystemOpenCode:
		return filepath.Join(targetRoot, ".opencode", "agents")
	default:
		return ""
	}
}

func ecosystemDisplayName(ecosystem Ecosystem) string {
	switch ecosystem {
	case EcosystemClaude:
		return "Claude"
	case EcosystemKiro:
		return "Kiro"
	case EcosystemPi:
		return "Pi"
	case EcosystemOpenCode:
		return "OpenCode"
	default:
		return string(ecosystem)
	}
}
