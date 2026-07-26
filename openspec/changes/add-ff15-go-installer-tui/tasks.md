## 1. CLI and packaging

- [x] 1.1 Create Go project entrypoint for `ff15`
- [x] 1.2 Add `--help`, `--init`, and `init` command flow
- [x] 1.3 Add dry-run and checkpoint reporting primitives

## 2. Interactive installer flow

- [x] 2.1 Add multi-select for Claude Code, Kiro, Pi Agents, and OpenCode
- [x] 2.2 Add mandatory tool handling for Engram, CodeGraph, and OpenSpec
- [x] 2.3 Add optional tool handling for Headroom and RTK
- [x] 2.4 Add Linux/Windows OS detection and platform-aware planning

## 3. Project file generation and patching

- [x] 3.1 Create/update target markdown/config files per selected ecosystem
- [x] 3.2 Inject managed guidance into `AGENTS.md` and `CLAUDE.md`
- [x] 3.3 Inject CodeGraph and RTK guidance where selected

## 4. Validation

- [x] 4.1 Add tests for OS detection, planning, and markdown patching
- [x] 4.2 Run validation for CLI help and dry-run init flow
- [x] 4.3 Update README with Go installation and local usage
