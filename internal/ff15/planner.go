package ff15

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func BuildPlan(targetRoot string, platform Platform, ecosystems []Ecosystem, optionalTools []ToolName) (Plan, error) {
	normalizedEcosystems := normalizeEcosystems(ecosystems)
	if len(normalizedEcosystems) == 0 {
		return Plan{}, fmt.Errorf("at least one ecosystem must be selected")
	}
	selectedTools := normalizeTools(optionalTools)
	plan := Plan{
		Platform:      platform,
		TargetRoot:    targetRoot,
		Ecosystems:    normalizedEcosystems,
		SelectedTools: selectedTools,
	}
	for _, tool := range selectedTools {
		plan.Steps = append(plan.Steps, installStepForTool(tool, platform))
	}
	files := map[string]string{}
	for _, ecosystem := range normalizedEcosystems {
		for path, content := range filesForEcosystem(targetRoot, ecosystem, selectedTools) {
			files[path] = content
		}
	}
	for _, path := range sortedKeys(files) {
		plan.Steps = append(plan.Steps, Step{
			Kind:        StepPatch,
			Title:       fmt.Sprintf("Patch %s", filepath.ToSlash(path)),
			FilePath:    path,
			ManagedText: files[path],
		})
	}
	return plan, nil
}

func installStepForTool(tool ToolName, platform Platform) Step {
	title := fmt.Sprintf("Install %s", strings.ToUpper(string(tool[:1]))+string(tool[1:]))
	switch tool {
	case ToolOpenSpec:
		return Step{
			Kind:     StepInstall,
			Title:    title,
			Commands: []string{"npm install -g @fission-ai/openspec@latest"},
			Verify:   "openspec --version",
		}
	case ToolCodeGraph:
		return Step{
			Kind:     StepInstall,
			Title:    title,
			Commands: []string{"npm install -g @colbymchenry/codegraph@latest"},
			Verify:   "codegraph --version",
			ManualHint: platformHint(platform,
				"If npm global bin is not on PATH yet, open a new shell and run `codegraph install` after `ff15 init`.",
				"If the npm global bin is not on PATH yet, open a new PowerShell and run `codegraph install` after `ff15 init`.",
			),
		}
	case ToolEngram:
		return Step{
			Kind:     StepInstall,
			Title:    title,
			Commands: []string{"go install github.com/Gentleman-Programming/engram/cmd/engram@latest"},
			Verify:   "engram version",
			ManualHint: platformHint(platform,
				"Run `engram setup claude`, `engram setup pi`, `engram setup kiro`, or `engram setup opencode` after install as needed.",
				"Run `engram setup claude`, `engram setup pi`, `engram setup kiro`, or `engram setup opencode` in a new PowerShell after install as needed.",
			),
		}
	case ToolHeadroom:
		command := "python3 -m pip install --user \"headroom-ai[all]\""
		if platform == PlatformWindows {
			command = "py -m pip install --user \"headroom-ai[all]\""
		}
		verify := "headroom --version"
		if platform == PlatformWindows {
			verify = "headroom --version"
		}
		return Step{
			Kind:     StepInstall,
			Title:    title,
			Commands: []string{command},
			Verify:   verify,
			ManualHint: platformHint(platform,
				"Headroom CLI ships from the PyPI package. If Python or pip is missing, install Python 3.10+ first.",
				"Headroom CLI ships from the PyPI package. If `py` or pip is missing, install Python 3.10+ first.",
			),
		}
	case ToolRTK:
		return Step{
			Kind:     StepInstall,
			Title:    title,
			Commands: []string{"cargo install --git https://github.com/rtk-ai/rtk rtk"},
			Verify:   "rtk --help",
			ManualHint: platformHint(platform,
				"RTK installs via Cargo in this first slice. Install Rust/Cargo first if the command fails.",
				"RTK installs via Cargo in this first slice. Install Rust/Cargo first if the command fails.",
			),
		}
	default:
		return Step{Kind: StepInstall, Title: title}
	}
}

func filesForEcosystem(targetRoot string, ecosystem Ecosystem, tools []ToolName) map[string]string {
	shared := sharedGuidance(tools)
	result := map[string]string{}
	for _, path := range ecosystemManagedPaths(targetRoot, ecosystem) {
		switch path {
		case filepath.Join(targetRoot, "CLAUDE.md"):
			result[path] = sharedClaudeGuidance(tools)
		case filepath.Join(targetRoot, "AGENTS.md"):
			result[path] = shared
		case filepath.Join(targetRoot, ".pi", "AGENTS.md"):
			result[path] = ecosystemNote("Pi Agents", []string{"Reference the root AGENTS.md managed block first.", "Add Pi-specific behavior only outside the managed block."})
		case filepath.Join(targetRoot, ".opencode", "agents", "ff15-openspec-agents.md"):
			result[path] = ecosystemNote("OpenCode", []string{"Point OpenCode agents back to the root AGENTS.md managed guidance.", "Keep generated instructions idempotent so reruns stay safe."})
		case filepath.Join(targetRoot, ".kiro", "steering", "ff15-openspec-agents.md"):
			result[path] = shared
		}
	}
	if ecosystem == EcosystemClaude {
		result[filepath.Join(targetRoot, ".claude", "agents", "ff15-openspec-agents.md")] = ecosystemNote("Claude Code", []string{"Use CLAUDE.md as the shared project contract.", "Keep generated agent notes short and repo-specific."})
	}
	return result
}

func sharedGuidance(tools []ToolName) string {
	sections := []string{
		"# Engram Documentation & Protocols",
		"",
		"## 1. Artifact Storage Policy (Orchestrator)",
		"",
		"- `engram`: Default when available; persistent memory across sessions.",
		"- `openspec`: File-based artifacts; use when the project needs reviewable spec files.",
		"",
		"## 2. Sub-Agent Context Protocol (SDD Orchestrator)",
		"",
		"### Non-SDD Tasks (General Delegation)",
		"",
		"- Read context: Orchestrator searches Engram via `mem_search` and passes relevant context into the sub-agent prompt.",
		"- Write context: Sub-agent MUST save significant discoveries, decisions, or bug fixes via `mem_save` before returning.",
		"- Prompt rule: Always add this line to delegated prompts: `If you make important discoveries, decisions, or fix bugs, save them to engram via mem_save with project: {project}.`",
		"",
		"### Content Retrieval",
		"",
		"1. `mem_search(query: topic_key, project: project)` -> get observation id",
		"2. `mem_get_observation(id)` -> full content. Required because search results are truncated.",
		"",
		"## 3. Recovery Rule (Orchestrator)",
		"",
		"`engram -> mem_search(...) -> mem_get_observation(...)`",
		"",
		"## 4. Engram Persistent Memory Protocol",
		"",
		"Use Engram proactively. Do not wait for the user to ask.",
		"",
		"### Proactive save triggers",
		"",
		"Call `mem_save` immediately after any of these:",
		"- Architecture or design decision",
		"- Team convention or workflow change",
		"- Tool or library choice with tradeoffs",
		"- Bug fix with root cause",
		"- Feature implemented with a non-obvious approach",
		"- Significant issue, GitHub, Jira, or Notion artifact update",
		"- Configuration or environment change",
		"- Non-obvious discovery, gotcha, or edge case",
		"- Pattern established",
		"- User preference or constraint learned",
		"",
		"Self-check after every task: Did I make a decision, fix a bug, learn something non-obvious, or establish a convention? If yes, call `mem_save` now.",
		"",
		"### `mem_save` format",
		"",
		"- `title`: Verb + what; short and searchable",
		"- `type`: `bugfix` | `decision` | `architecture` | `discovery` | `pattern` | `config` | `preference`",
		"- `scope`: `project` by default, or `personal`",
		"- `topic_key`: stable key for evolving topics",
		"- `content`: include What, Why, Where, Learned",
		"",
		"### Topic update rules",
		"",
		"- Different topics must not overwrite each other.",
		"- Same topic evolving -> reuse the same `topic_key`.",
		"- Unsure about the key -> call `mem_suggest_topic_key` first.",
		"- Know the exact observation id -> use `mem_update`.",
		"",
		"### When to search memory",
		"",
		"On any variation of remember/recall/past work questions:",
		"1. `mem_context` first",
		"2. then `mem_search` if needed",
		"3. then `mem_get_observation` for full content",
		"",
		"Search proactively when starting work on something that may have been done before, or when the user's first message references an existing project feature/problem.",
		"",
		"### Session close protocol",
		"",
		"Before ending a session or saying done/listo, call `mem_session_summary` with Goal, Instructions, Discoveries, Accomplished, Next Steps, and Relevant Files.",
		"",
		"## 5. CodeGraph",
		"",
		"When answering structural or codebase questions, use CodeGraph before broad filesystem searches.",
		"",
		"CodeGraph-aware worktree placement:",
		"- Create CodeGraph-dependent worktrees under the user's home directory, not under `/tmp` or other temporary folders.",
		"- Every worktree needs its own `.codegraph/` index. Never copy or reuse another checkout's index.",
		"",
		"CodeGraph intelligence surface:",
		"- Prefer the `codegraph_explore` MCP tool when available.",
		"- Otherwise use upstream CLI commands such as `codegraph status`, `codegraph query`, `codegraph explore`, `codegraph node`, `codegraph files`, `codegraph callers`, `codegraph callees`, `codegraph impact`, and `codegraph affected`.",
		"- Do not use `gentle-ai codegraph` as a general proxy.",
		"- Do not recommend destructive lifecycle commands such as `codegraph uninit`, `codegraph install`, `codegraph uninstall`, or `codegraph upgrade` unless the user explicitly asks.",
		"",
		"Required order for structural/codebase questions:",
		"1. Resolve the project root with `git rev-parse --show-toplevel || pwd`.",
		"2. Confirm the root is a real project/workspace.",
		"3. Check for `<project-root>/.codegraph/` before broad Read/Glob/Grep exploration.",
		"4. Missing `.codegraph/` means initialize CodeGraph; it is not a reason to skip it.",
		"5. Use `codegraph_explore` or read-only upstream CLI commands after initialization.",
		"6. After edits, rely on watcher auto-sync by default and run `codegraph sync` only when files stay stale.",
		"7. Fall back to normal filesystem tools only after CodeGraph initialization or use fails, and explain the fallback briefly.",
	}
	if hasTool(tools, ToolHeadroom) {
		sections = append(sections,
			"",
			"## 6. Headroom",
			"",
			"- Headroom is optional. Use it only when the project explicitly opts in.",
			"- Preferred local install path in this first slice is the PyPI CLI: `python -m pip install --user \"headroom-ai[all]\"` or `py -m pip install --user \"headroom-ai[all]\"` on Windows.",
			"- Use `headroom doctor` after installation to verify the local proxy/tooling path.",
		)
	}
	if hasTool(tools, ToolRTK) {
		sections = append(sections, "", rtkGuidance())
	}
	return strings.Join(sections, "\n")
}

func sharedClaudeGuidance(tools []ToolName) string {
	return strings.Join([]string{
		"# FF15 Claude guidance",
		"",
		"This managed block keeps Claude Code aligned with the project-level FF15 workflow.",
		"",
		sharedGuidance(tools),
	}, "\n")
}

func ecosystemNote(name string, bullets []string) string {
	lines := []string{"# " + name + " FF15 note", "", "This file was generated by `ff15 init`."}
	for _, bullet := range bullets {
		lines = append(lines, "- "+bullet)
	}
	return strings.Join(lines, "\n")
}

func hasTool(tools []ToolName, want ToolName) bool {
	for _, tool := range tools {
		if tool == want {
			return true
		}
	}
	return false
}

func platformHint(platform Platform, linuxHint, windowsHint string) string {
	if platform == PlatformWindows {
		return windowsHint
	}
	return linuxHint
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return filepath.ToSlash(keys[i]) < filepath.ToSlash(keys[j])
	})
	return keys
}

func rtkGuidance() string {
	return strings.Join([]string{
		"## 7. RTK (Rust Token Killer)",
		"",
		"Golden rule: always prefix commands with `rtk`.",
		"",
		"Examples:",
		"- `rtk cargo build`",
		"- `rtk cargo test`",
		"- `rtk go test`",
		"- `rtk git status`",
		"- `rtk git diff`",
		"- `rtk npm run <script>`",
		"- `rtk grep <pattern>`",
		"- `rtk find <pattern>`",
		"",
		"Even in chained commands, prefix each command with `rtk`.",
		"",
		"Wrong:",
		"- `git add . && git commit -m \"msg\" && git push`",
		"",
		"Correct:",
		"- `rtk git add . && rtk git commit -m \"msg\" && rtk git push`",
	}, "\n")
}
