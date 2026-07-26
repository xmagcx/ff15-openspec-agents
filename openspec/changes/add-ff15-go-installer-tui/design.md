## Context

This repo currently provides OpenSpec-driven agent assets plus a Python sync script focused on GitHub Copilot folders. The new feature is broader: a Go CLI named `ff15` that configures multiple agent ecosystems and plans local tool setup.

## Goals / Non-Goals

- Goals:
  - Provide a first working cross-platform init flow for Linux and Windows.
  - Keep first-run UX interactive and selection-driven.
  - Support dry-run planning and checkpointed apply behavior.
  - Patch project files deterministically with managed blocks.
  - Reuse the useful AGENTS.md sync behavior from the existing Python script.
- Non-Goals:
  - Full-screen rich TUI in the first slice.
  - macOS support in the first slice.
  - Deep upstream installer automation for every third-party tool.

## Decisions

- Decision: Build the first version in Go.
  - Why: the user explicitly changed direction away from Python, and Go is easier to distribute as a single local binary on Linux and Windows.
- Decision: Support both `ff15 --init` and `ff15 init`.
  - Why: both invocation styles were requested.
- Decision: Keep the prompt flow stdlib-first.
  - Why: minimizes dependencies while still delivering interactive multi-select behavior.
- Decision: Model tool setup as a checkpointed plan with optional execution.
  - Why: dry-run output stays reviewable and safer for first adoption.
- Decision: Use managed markdown blocks for generated guidance.
  - Why: reruns stay idempotent and preserve unrelated user content.

## Alternatives considered

- Extend the Python script directly.
  - Rejected: it no longer matches the preferred implementation language or broader scope.
- Build a dependency-heavy full-screen TUI first.
  - Rejected: adds complexity before the planning model stabilizes.
- Only patch root files and skip ecosystem folders.
  - Rejected: selected ecosystems need dedicated, reviewable generated entrypoints.

## Architecture

- `cmd/ff15/main.go`: executable entrypoint
- `internal/ff15/cli.go`: argument parsing and init flow
- `internal/ff15/model.go`: platform, ecosystem, tool, and plan types
- `internal/ff15/planner.go`: selection-to-plan conversion
- `internal/ff15/runtime.go`: checkpoint execution for install commands
- `internal/ff15/patcher.go`: managed markdown create/update logic
- `internal/ff15/ui.go`: interactive prompt helpers
- `internal/ff15/*_test.go`: focused unit coverage

## Risks / Trade-offs

- Automatic installation metadata for third-party tools may evolve.
  - Mitigation: keep command and hint selection centralized in the planner.
- Windows execution depends on PowerShell availability.
  - Mitigation: branch shell execution by detected OS and report clear checkpoint failures.
- The first slice is an interactive CLI rather than a rich terminal UI.
  - Mitigation: keep planner/runtime boundaries clean so a richer UI can be added later.

## Migration Plan

1. Replace the Python-first proposal language with Go.
2. Add the Go CLI and focused tests.
3. Document local build/install steps in README.
4. Iterate on installer metadata as tool distribution details stabilize.
