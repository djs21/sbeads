## ADDED Requirements

### Requirement: REQ-CLI-1: Cobra CLI structure
The system SHALL organize the command-line interface using `spf13/cobra` vertical slices, where each command is implemented in its own dedicated source file under `cmd/bd/` adhering to a maximum 400 LOC ceiling, and shares execution context through an initialized `AppContext`.

#### Scenario: Subcommand execution through vertical slices
- **WHEN** user executes any supported subcommand such as `sbeads create`, `sbeads ready`, `sbeads claim`, `sbeads close`, `sbeads show`, `sbeads dep`, or `sbeads version`
- **THEN** Cobra routes execution to the respective vertical slice handler and injects the initialized `AppContext`

#### Scenario: Database initialization in PersistentPreRunE
- **WHEN** user runs any database-requiring command without an explicit database flag
- **THEN** `rootCmd.PersistentPreRunE` automatically resolves the target database path via environment variables or ancestor directory traversal and initializes storage and domain services before invoking the command

### Requirement: REQ-CLI-2: POSIX flag support
The system SHALL support standard POSIX flag conventions across all subcommands, including single-hyphen short flags (e.g., `-t`, `-p`, `-d`, `-a`, `-m`, `-h`) and double-hyphen long flags (e.g., `--title`, `--priority`, `--description`, `--assignee`, `--agent`, `--help`).

#### Scenario: Create issue with short flags
- **WHEN** user executes `sbeads create -t "Fix crash" -p 1 -d "Memory leak in query"`
- **THEN** the command successfully parses the title as "Fix crash", priority as 1, and description as "Memory leak in query" and creates the issue

#### Scenario: Create issue with long flags
- **WHEN** user executes `sbeads create --title "Fix crash" --priority 2 --type bug`
- **THEN** the command successfully parses the arguments identically to short flag invocations

### Requirement: REQ-CLI-3: Per-command help with `-h`/`--help` & `sbeads help <cmd>`
The system SHALL provide contextual help information for every subcommand when invoked with `-h`, `--help`, or via the standard Cobra syntax `sbeads help <cmd>`.

#### Scenario: Show help via short flag
- **WHEN** user executes `sbeads claim -h`
- **THEN** the CLI outputs usage, description, and available flags specific to the `claim` subcommand and exits with code 0

#### Scenario: Show help via long flag
- **WHEN** user executes `sbeads ready --help`
- **THEN** the CLI outputs usage, description, and available flags specific to the `ready` subcommand and exits with code 0

#### Scenario: Show help via subcommand syntax
- **WHEN** user executes `sbeads help create`
- **THEN** the CLI outputs the identical help documentation as `sbeads create --help` and exits with code 0
