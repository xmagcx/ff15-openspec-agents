---
name: ff15-openspec-agents-sync
description: FF15-inspired GitHub Copilot agents for OpenSpec workflows. Team includes Noctis (orchestrator + OpenSpec author), Iris (issue management), Gladiolus (implementation), Prompto (code quality), Ignis (documentation + archival), and Lunafreya (PR creation).
---

# FF15 Copilot Agents - OpenSpec Edition

Skill for syncing FF15 team agent definitions to the project.

## Quick Start

```bash
python .claude/skills/ff15-openspec-agents-sync/scripts/sync_agents.py --target .
```

## Agents

- **Noctis** - Orchestrator + OpenSpec author
- **Iris** - Issue management
- **Gladiolus** - Implementation
- **Prompto** - Code quality
- **Ignis** - Documentation + archival
- **Lunafreya** - PR creation

## Details

- Usage: [USAGE.md](USAGE.md)
- Agent definitions: `agents/*.agent.md`
- Policies: `docs/*-policy.md`
