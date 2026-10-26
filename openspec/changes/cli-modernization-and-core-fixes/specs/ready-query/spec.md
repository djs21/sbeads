## ADDED Requirements

### Requirement: REQ-RDY-1: Filter closed issues
The storage and domain ready query SHALL exclude any issue whose `is_closed` flag is set to 1.

#### Scenario: Closed issues excluded from ready output
- **WHEN** the database contains an open issue and a closed issue, and the user queries ready issues via `sbeads ready`
- **THEN** the output contains only the open issue and excludes the closed issue

### Requirement: REQ-RDY-2: Filter claimed/blocked issues
The storage and domain ready query SHALL exclude any issue whose `is_claimed` flag is set to 1 or whose `is_blocked` flag is set to 1.

#### Scenario: Claimed issue excluded from ready query
- **WHEN** an open issue has been claimed by an assignee (`is_claimed = 1`)
- **THEN** querying `sbeads ready` does not list the claimed issue

#### Scenario: Manually blocked issue excluded from ready query
- **WHEN** an open issue is marked as blocked (`is_blocked = 1`)
- **THEN** querying `sbeads ready` does not list the blocked issue

### Requirement: REQ-RDY-3: Dependency tree resolution - parent open blocks child
The storage ready query SHALL evaluate parent-child relationships such that an open child issue is NOT considered ready if any of its parent dependencies have `is_closed = 0`. Once all parent dependencies are closed (`is_closed = 1`), the child issue SHALL immediately become ready.

#### Scenario: Child issue with open parent dependency
- **WHEN** issue B depends on parent issue A, and issue A is not closed (`is_closed = 0`)
- **THEN** querying `sbeads ready` displays parent issue A but does not display child issue B

#### Scenario: Child issue becomes ready after parent is closed
- **WHEN** parent issue A is closed via `sbeads close <parent-id>`
- **THEN** querying `sbeads ready` subsequently includes child issue B in the ready list
