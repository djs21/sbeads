## ADDED Requirements

### Requirement: REQ-CLI-1: Enhanced `sbeads ready` Formatting
The `ready` command SHALL format all actionable issues using the status engine format `<id> [O] (P<priority>/<type>) <title>`.

#### Scenario: Listing ready issues with badges
- **GIVEN** an open issue with ID "sbeads-55g", priority 1, type "feature", title "Schema migration"
- **WHEN** `sbeads ready` is executed
- **THEN** output contains `sbeads-55g [O] (P1/feature) Schema migration`

### Requirement: REQ-CLI-2: Enhanced `sbeads show` Detailed Observability
The `show` command SHALL render a single computed `Status:` line, claiming agent info, blockers, and dependents.

#### Scenario: Showing issue with blockers and dependents
- **GIVEN** an issue with parent dependencies and child dependents
- **WHEN** `sbeads show <id>` is executed
- **THEN** output displays `Status:`, `Blocked by:`, and `Blocks:` sections

### Requirement: REQ-CLI-3: `sbeads list` Catalog Command
The `list` command SHALL support listing unclosed issues by default, all issues via `--all`, closed via `--closed`, blocked via `--blocked`, tree hierarchy via `--tree`, and machine-readable output via `--json`.

#### Scenario: Listing with --tree flag
- **GIVEN** an epic issue with child dependencies
- **WHEN** `sbeads list --tree` is executed
- **THEN** children are printed indented under the parent epic

### Requirement: REQ-CLI-4: `sbeads unclaim` Command
The `unclaim` command SHALL reset `is_claimed` to 0, clear `agent` and `assignee`, and release the task back to open.

#### Scenario: Releasing claimed issue
- **GIVEN** an issue claimed by "worker-1"
- **WHEN** `sbeads unclaim <id>` is executed
- **THEN** issue is unclaimed, agent is cleared, and output confirms `Unclaimed issue:`
