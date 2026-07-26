## ADDED Requirements

### Requirement: FF15 init wizard

The system SHALL provide an `ff15` Go CLI with an interactive initialization flow that users can start with `ff15 --init`.

#### Scenario: Help output is available

- **WHEN** the user runs `ff15 --help`
- **THEN** the CLI shows available commands and options
- **AND** the output documents `--init`

#### Scenario: Init alias is accepted

- **WHEN** the user runs `ff15 init`
- **THEN** the CLI starts the same initialization flow as `ff15 --init`

### Requirement: Agent ecosystem selection

The init flow SHALL allow multi-select selection of agent ecosystems to configure for the current project.

#### Scenario: Multiple ecosystems selected

- **WHEN** the user selects Claude Code, Kiro, Pi Agents, and OpenCode
- **THEN** the generated action plan includes file creation or update steps for each selected ecosystem

#### Scenario: Single ecosystem selected

- **WHEN** the user selects only one ecosystem
- **THEN** the generated action plan only modifies files required for that ecosystem

### Requirement: Mandatory and optional tool selection

The init flow SHALL always include mandatory tools and SHALL allow users to opt into optional tools.

#### Scenario: Mandatory tools are always planned

- **WHEN** the init flow builds an installation plan
- **THEN** Engram, CodeGraph, and OpenSpec are included automatically

#### Scenario: Optional tools are selectable

- **WHEN** the user selects Headroom and RTK
- **THEN** the plan includes setup checkpoints for both tools
- **AND** omits unselected optional tools

### Requirement: Platform-aware local installation planning

The init flow SHALL detect the target operating system and choose installation commands or guidance compatible with that platform.

#### Scenario: Linux client detected

- **WHEN** the init flow runs on Linux
- **THEN** the plan uses Linux-compatible install commands or hints

#### Scenario: Windows client detected

- **WHEN** the init flow runs on Windows
- **THEN** the plan uses Windows-compatible install commands or hints

### Requirement: Checkpointed execution

The init flow SHALL report explicit checkpoints and surface errors without hiding failed steps.

#### Scenario: Step succeeds

- **WHEN** an installation verification or patching step completes
- **THEN** the UI marks that checkpoint as successful

#### Scenario: Step fails

- **WHEN** an installation verification or patching step fails
- **THEN** the UI shows the failing checkpoint
- **AND** includes actionable error detail for the user

### Requirement: Markdown/config patching for selected ecosystems

The init flow SHALL create missing project files and update existing ones with managed content blocks required by the selected ecosystems.

#### Scenario: Root guidance file missing

- **WHEN** a selected ecosystem requires `AGENTS.md` or `CLAUDE.md` and the file does not exist
- **THEN** the system creates the file with the required managed content

#### Scenario: Root guidance file exists

- **WHEN** a selected ecosystem requires `AGENTS.md` or `CLAUDE.md` and the file already exists
- **THEN** the system updates or appends only the managed content block
- **AND** preserves unrelated user content

### Requirement: Engram and CodeGraph guidance injection

The init flow SHALL inject reusable guidance for Engram memory, CodeGraph workflow, and RTK usage into the applicable markdown guidance files.

#### Scenario: Pi or root agent guidance updated

- **WHEN** the system patches an applicable `AGENTS.md` file
- **THEN** it includes Engram protocol guidance
- **AND** CodeGraph usage guidance
- **AND** RTK guidance content when RTK is selected

#### Scenario: Claude guidance updated

- **WHEN** the system patches `CLAUDE.md`
- **THEN** it includes the same operational guidance adapted for Claude-compatible project instructions
