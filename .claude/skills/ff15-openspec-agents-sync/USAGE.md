# FF15 Copilot Agents - OpenSpec Edition - Usage Guide

This guide explains how to use the FF15 OpenSpec agents effectively with GitHub Copilot.

## Overview

FF15 OpenSpec agents provide a spec-driven development workflow built on the [OpenSpec](https://github.com/Fission-AI/OpenSpec) framework. OpenSpec is a lightweight spec management tool for AI-assisted coding that enables predictable and auditable development by having humans and AI agree on specs before implementation.

This agent team creates OpenSpec documents (proposal, tasks, design) and then autonomously implements, improves quality, documents, and creates PRs based on those specs.

---

## Prerequisites

Before using the FF15 OpenSpec agents, you need the following environment:

- **Node.js** >= 20.19.0 (check with `node --version`)
- **GitHub Copilot** (VS Code or compatible editor)
- **OpenSpec CLI** (setup in the following section)
- **Git** (for version control and PR creation)

---

## OpenSpec Setup

### Step 1: Install OpenSpec CLI

**Option A: Install via npm**

```bash
npm install -g @fission-ai/openspec@latest
```

Verify installation:

```bash
openspec --version
```

**Option B: Install via Nix (for NixOS or Nix package manager users)**

Run directly (no installation needed):

```bash
nix run github:Fission-AI/OpenSpec -- init
```

Install to profile:

```bash
nix profile install github:Fission-AI/OpenSpec
```

### Step 2: Initialize the Project

Navigate to your project directory:

```bash
cd your-project
```

Initialize OpenSpec:

```bash
openspec init
```

During the initialization process:
- You will be prompted to select your AI tool (Claude Code, Cursor, Qoder, etc.)
- Slash commands for the selected tool are automatically configured
- `AGENTS.md` is created at the project root
- The `openspec/` directory structure is created

**Important**: After initialization, restart your AI assistant to enable the slash commands.

### Step 3: Set Up Project Context (Optional)

After initialization, configure project-specific information:

```bash
# Ask your AI assistant:
"Please read openspec/project.md and help me fill it out with details about my project, tech stack, and conventions"
```

In `openspec/project.md`, document the conventions, architecture patterns, coding standards, etc. to be followed across the project.

### Step 4: Verify Setup

Verify that everything is set up correctly:

```bash
openspec list
```

This command shows active change folders (empty in the initial state).

### Further Information

- **OpenSpec Official Repository**: https://github.com/Fission-AI/OpenSpec
- **OpenSpec Official Site**: https://openspec.dev/
- **Documentation**: https://github.com/Fission-AI/OpenSpec/tree/main/docs

---

## Philosophy First

FF15 OpenSpec agents focus on **autonomous execution with minimal disruption**. They operate through a clear spec-driven process:
- **Specs come first** via OpenSpec documents
- **Minimal user intervention** (approval + verification only)
- **Quality is built in** through autonomous review and improvement
- **Documentation is continuous** throughout the workflow

---

## Quick Start

### Start with Noctis

Most tasks start with **Noctis**:

```
/noctis Add a shopping cart feature
/noctis Refactor the authentication module
/noctis Add real-time notifications
```

Noctis will:
1. Create OpenSpec documents (proposal, tasks, design) with you
2. Request approval of the spec
3. Optionally delegate Issue creation to Iris
4. Delegate implementation to Gladiolus
5. Delegate code quality improvement to Prompto
6. Delegate documentation and archival to Ignis
7. Delegate PR creation to Lunafreya
8. Notify you with the PR link for final verification

---

## Team Structure

```
                         @Noctis
                  (Orchestrator + Spec Author)
                 Creates OpenSpec and leads the workflow
                            |
        ┌───────────────────┼───────────────────┬───────────────────┬───────────────────┐
        |                   |                   |                   |                   |
     @Iris            @Gladiolus           @Prompto             @Ignis           @Lunafreya
   (Issues)           (Implementation)    (Quality)         (Documentation)         (PR)
   Manages           Builds based        Reviews and       Archives specs       Creates
   GitHub            on OpenSpec         refines for       and updates          pull
   Issues            specs              quality           documentation        requests
```

**Key**: Spec-driven development with autonomous execution.

---

## When to Use Each Agent

### 👑 Noctis (Orchestrator + Spec Author)

**Best for:** Complex tasks that need OpenSpec creation and coordinated workflow

**Examples:**
```
@Noctis Implement OAuth2 authentication
@Noctis Migrate the API from REST to GraphQL
@Noctis Build an admin dashboard with user management
```

**What Noctis does:**
- Creates OpenSpec documents through dialogue (proposal.md, tasks.md, design.md)
- Requests user approval of the spec
- Orchestrates the implementation workflow
- Delegates to the appropriate specialists
- Tracks progress throughout the workflow
- Notifies the user at key milestones
- Delivers a complete, integrated result

**OpenSpec Creation Process:**
1. Discuss requirements with the user
2. Create proposal.md (overview, motivation, approach)
3. Create tasks.md (implementation checklist)
4. Create design.md (detailed technical design)
5. Request user approval before proceeding

**When Noctis calls the team:**
- Issue management needed → `@Iris`
- Implementation needed → `@Gladiolus`
- Code quality improvement needed → `@Prompto`
- Documentation and archival needed → `@Ignis`
- PR creation needed → `@Lunafreya`

---

### 📋 Iris (Issue Management Specialist)

**Best for:** Creating and managing GitHub Issues based on specs

**Examples:**
```
@Iris Create Issues for the shopping cart feature
@Iris Update Issue #42 with implementation status
@Iris Create Issues based on OpenSpec tasks
```

**What Iris does:**
- Creates GitHub Issues with clear descriptions
- References OpenSpec documents within Issues
- Manages Issue lifecycle (create, update, close)
- Links Issues to pull requests
- Organizes work items

**When to consult Iris:**
- After OpenSpec approval, before implementation
- Need GitHub Issues for tracking
- Want to link implementation to Issues
- Managing multiple related Issues

---

### 🧠 Ignis (Documentation and Archive Specialist)

**Best for:** Documentation updates and OpenSpec archival

**Examples:**
```
@Ignis Update the authentication feature documentation
@Ignis Archive OpenSpec and update CHANGELOG
@Ignis Create comprehensive README for the new module
```

**What Ignis does:**
- Updates README with new features and changes
- Maintains CHANGELOG with version history
- Archives completed OpenSpec documents
- Creates technical documentation
- Documents API interfaces and usage
- Keeps documentation in sync with code

**OpenSpec Archival:**
1. Move proposal.md to openspec/changes/archive/
2. Rename with timestamp (YYYY-MM-DD__description.md)
3. Update references in documentation
4. Ensure traceability of changes

**Documentation Types:**
- README: Project overview and setup
- CHANGELOG: Version history and changes
- Technical documentation: Architecture and design
- API documentation: Interface specifications
- OpenSpec archives: Historical change records

**When to consult Ignis:**
- After implementation is complete
- Need documentation updates
- Need OpenSpec archival
- Creating new project documentation

---

### 💪 Gladiolus (Implementation Specialist)

**Best for:** Writing code and building features based on OpenSpec

**Examples:**
```
@Gladiolus Implement based on OpenSpec change-042
@Gladiolus Build the shopping cart feature
@Gladiolus Add input validation per the spec
```

**What Gladiolus does:**
- Implements features based on OpenSpec design
- Follows the spec precisely
- Writes clean, working code
- Tests features thoroughly
- Drives to completion
- Reports progress and blockers

**Implementation Philosophy:**
- Follow OpenSpec design.md as the blueprint
- Maintain code quality standards
- Include tests as part of implementation
- No scope creep beyond the spec
- Speak up about blocking issues

**When to consult Gladiolus:**
- After OpenSpec approval
- Need direct implementation
- Building features from a clear spec
- Bug fixes with clear requirements

---

### ✨ Prompto (Code Quality Specialist)

**Best for:** Code quality improvement, OpenSpec compliance, refactoring

**Examples:**
```
@Prompto Review and improve the authentication code
@Prompto Ensure OpenSpec compliance in recent changes
@Prompto Refactor for better maintainability
```

**What Prompto does:**
- Verifies OpenSpec compliance (cross-references design.md)
- Enforces review-policy guidelines
- Performs code quality reviews
- Refactors for clarity and maintainability
- Identifies improvement opportunities
- Ensures consistent code patterns
- Performs safe refactoring without breaking features

**Quality Improvement Focus:**
- **OpenSpec Compliance**: Implementation matches design
- **Review Policy**: Follows project review guidelines
- **Code Clarity**: Readable and maintainable
- **Pattern Consistency**: Follows established patterns
- **Best Practices**: Adheres to language standards

**Refactoring Approach:**
- Clarity over cleverness
- Consistent naming and patterns
- Maintain functionality without behavior change
- Safe transformations with tests

**When to consult Prompto:**
- After implementation, before documentation
- Need quality improvement
- Ensuring OpenSpec compliance
- Code readability concerns
- Refactoring for better maintainability

---

### 🌙 Lunafreya (PR Creation Specialist)

**Best for:** Creating and finalizing pull requests

**Examples:**
```
@Lunafreya Create a PR for the authentication implementation
@Lunafreya Finalize the PR with proper description
@Lunafreya Create a PR linking to Issue #42
```

**What Lunafreya does:**
- Creates pull requests with clear descriptions
- Links PRs to related Issues
- References OpenSpec documents
- Verifies CI passes
- Ensures all changes are committed
- Prepares for merge and deployment

**PR Creation Checklist:**
1. All code changes committed
2. Tests pass (CI is green)
3. Documentation updated
4. CHANGELOG updated
5. OpenSpec archived
6. Issue references included

**When to consult Lunafreya:**
- After documentation and archival are complete
- Ready to create a pull request
- Need a PR for completed work
- Finalizing implementation delivery

---

## Common Workflows

### Feature Implementation (OpenSpec-Driven)

```
User: @Noctis Add a shopping cart feature

Workflow:
1. Noctis: Creates OpenSpec (proposal, tasks, design) with user
2. User: Approves the spec
3. Iris: Creates GitHub Issues (if requested)
4. Gladiolus: Implements based on OpenSpec design
5. Prompto: Reviews OpenSpec compliance and quality
6. Ignis: Updates documentation and archives OpenSpec
7. Lunafreya: Creates PR with proper description
8. Noctis: Notifies user, requests verification
9. User: Verifies and approves merge
```

### Bug Fix with OpenSpec

```
User: @Noctis Fix the session timeout issue

Workflow:
1. Noctis: Creates minimal OpenSpec (proposal + tasks)
2. User: Approves fix approach
3. Gladiolus: Implements the fix
4. Prompto: Reviews quality
5. Ignis: Updates CHANGELOG
6. Lunafreya: Creates PR
7. Noctis: Notifies completion
```

### Direct Agent Usage

```
User: @Iris Create an Issue for CSV export feature
→ Direct Issue creation

User: @Gladiolus Implement per OpenSpec change-042
→ Direct implementation

User: @Prompto Improve code quality of the auth module
→ Direct quality improvement

User: @Lunafreya Create a PR for completed work
→ Direct PR creation
```

---

## Best Practices

### Do's

✅ **Start with Noctis for new features** - Let it create the OpenSpec
✅ **Approve specs before implementation** - Ensure clarity upfront
✅ **Trust the autonomous workflow** - Minimal intervention is enough
✅ **Verify the final result** - Review the PR before merging
✅ **Use direct agents for simple tasks** - Skip orchestration when appropriate

### Don'ts

❌ **Don't skip OpenSpec for complex features** - Prevents scope creep
❌ **Don't micromanage the workflow** - Trust the process
❌ **Don't skip quality reviews** - Prompto catches important issues
❌ **Don't skip documentation** - Ignis keeps everything current

---

## Task Type Examples

### New Feature
```
@Noctis Add two-factor authentication
```
→ Full OpenSpec workflow with team coordination

### Bug Fix
```
@Noctis Fix payment processing timeout
```
→ Minimal OpenSpec with direct fix

### Issue Creation
```
@Iris Create an Issue for search optimization
```
→ Direct to Iris for Issue management

### Implementation
```
@Gladiolus Implement per OpenSpec change-042
```
→ Direct to Gladiolus with clear spec

### Code Quality
```
@Prompto Review and improve the auth module
```
→ Direct to Prompto for quality improvement

### Documentation
```
@Ignis Update docs and archive OpenSpec
```
→ Direct to Ignis for documentation and archival

### PR Creation
```
@Lunafreya Create PR for auth implementation
```
→ Direct to Lunafreya when ready

---

## Agent Selection Quick Reference

| Need | Call |
|------|------|
| Complex feature with OpenSpec | `@Noctis` |
| GitHub Issue creation/management | `@Iris` |
| Implementation from OpenSpec | `@Gladiolus` |
| Code quality improvement | `@Prompto` |
| Documentation and archival | `@Ignis` |
| Pull request creation | `@Lunafreya` |
| Workflow orchestration | `@Noctis` |
| Quick bug fix | `@Noctis` (minimal spec) |

---

## OpenSpec Documents

The workflow revolves around three main documents:

### proposal.md
- **Overview**: What are we building?
- **Motivation**: Why are we building it?
- **Approach**: How will we build it?
- **Scope**: What's included and excluded?

### tasks.md
- **Checklist**: Step-by-step implementation tasks
- **Progress tracking**: Mark completed items
- **Dependencies**: Task ordering

### design.md
- **Technical design**: Detailed implementation plan
- **Architecture**: System structure
- **API contracts**: Interfaces and data models
- **Edge cases**: Error handling and validation

---

**Remember**: Spec-driven development with autonomous execution and minimal user disruption.
