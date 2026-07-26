# Change: Add FF15 Go installer CLI

## Why

This repository ships FF15 OpenSpec assets, but it does not yet provide a unified local installer for Claude Code, Kiro, Pi Agents, and OpenCode. Users need a guided cross-platform flow that detects Linux vs Windows, plans local tool setup, and patches project instructions consistently.

## What Changes

- Add a Go-based `ff15` CLI entrypoint for local project setup.
- Add interactive multi-select flow for Claude Code, Kiro, Pi Agents, and OpenCode.
- Always plan mandatory tools: Engram, CodeGraph, and OpenSpec.
- Allow optional planning for Headroom and RTK.
- Detect Linux vs Windows and surface platform-appropriate install hints or commands.
- Add checkpointed dry-run and apply flow with explicit per-step reporting.
- Create or update project markdown/config files with managed blocks.
- Reuse the existing AGENTS sync behavior as managed markdown patching in Go.

## Impact

- Affected specs: `installer-tui`
- Affected code: new Go CLI, planner, patcher, runtime, and tests
- Affected project files: `AGENTS.md`, `CLAUDE.md`, `.kiro/steering/*.md`, `.pi/AGENTS.md`, `.opencode/agents/*.md`, README usage docs
