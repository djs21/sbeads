# sbeads

`sbeads` is an ultra-lightweight, high-performance issue tracker and task coordinator designed for autonomous AI agents and software engineers. Built in Go on top of SQLite with Write-Ahead Logging (WAL), `sbeads` delivers zero-friction local task tracking, explicit dependency graphs, and atomic agent claim/close operations without external services or heavy daemon overhead.

---

## Overview & Purpose

Modern autonomous agent swarms and multi-agent development workflows require structured, reliable task coordination:
- **Zero Overhead**: Single static binary with embedded SQLite storage. No server daemon, Docker container, or cloud service required.
- **Agent-First Coordination**: Explicit claim semantics (`sbeads claim`) prevent duplicate work across concurrent agent panes or subagents.
- **Dependency Graphs**: First-class parent/child issue dependencies (`sbeads dep add`) block dependent tasks until prerequisite tasks finish.
- **Fast Ready Queues**: Query unblocked, actionable issues instantly (`sbeads ready`) to drive autonomous loops.
- **Deterministic Storage**: Local SQLite database file (`.sbeads`) that can be committed, scoped per git worktree, or shared across processes via environment variables.

---

## Architecture & Storage

`sbeads` is architected into clean, modular layers:

```
sbeads/
├── cmd/bd/           # CLI entrypoint, flag parsing, positional argument handling
├── internal/
│   ├── types/        # Core data models (Issue, Dependency) and ID generation
│   └── domain/       # Business logic, state transitions, validation, and domain errors
└── backend/          # SQLite WAL storage engine, auto-schema creation, query execution
```

### Storage Engine: SQLite in WAL Mode
- **Database File**: Default path is `.sbeads` in the active working directory. Override using the `SBEADS_DB` environment variable (e.g., `export SBEADS_DB=/shared/path/tasks.db`).
- **WAL Mode (`_journal_mode=WAL`)**: Connections open SQLite with Write-Ahead Logging enabled. This allows concurrent readers and non-blocking reads while writes occur, making it robust for concurrent agent swarms and subagents.
- **Auto-Initialization**: Database tables and indices are created automatically via `CREATE TABLE IF NOT EXISTS` on first invocation.

### Data Schema

#### `issues`
| Field | SQLite Type | Description |
|---|---|---|
| `id` | `TEXT PRIMARY KEY` | Unique ID in the format `issue-<32-hex-characters>` (cryptographically random) |
| `title` | `TEXT NOT NULL` | Short title or summary of the task |
| `issue_type` | `TEXT NOT NULL` | Type classification (e.g. `bug`, `feature`, `task`) |
| `priority` | `INTEGER NOT NULL` | Integer priority value (0–10, non-negative; higher = more urgent) |
| `description`| `TEXT` | Detailed problem description, criteria, or agent brief |
| `is_blocked` | `INTEGER DEFAULT 0`| Blocked state: `0` = unblocked / ready, `1` = blocked or in-progress |
| `is_closed`  | `INTEGER DEFAULT 0`| Lifecycle status: `0` = open, `1` = closed |
| `is_claimed` | `INTEGER DEFAULT 0`| Claim status: `0` = unclaimed, `1` = claimed by an agent |
| `created_at` | `TEXT NOT NULL` | RFC3339 timestamp of creation |
| `updated_at` | `TEXT NOT NULL` | RFC3339 timestamp of last update |

#### `dependencies`
| Field | SQLite Type | Description |
|---|---|---|
| `child_id` | `TEXT NOT NULL` | ID of the dependent child issue |
| `parent_id` | `TEXT NOT NULL` | ID of the prerequisite parent issue |

*Primary Key:* `(child_id, parent_id)`  
*Semantics:* Child issue depends on parent issue (`child -> parent`). The parent issue must be completed before the child can proceed.

---

## Installation & Build

### Prerequisites
- Go 1.27+ (or Go 1.21+ with CGO enabled)
- GCC / C toolchain (required for `go-sqlite3` CGO compilation)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/djs21/sbeads.git
cd sbeads

# Build the binary
go build -o sbeads ./cmd/bd

# Verify installation
./sbeads
```

To install system-wide or into `$GOPATH/bin`:
```bash
go install ./cmd/bd
# Binary will be named 'bd' in $GOPATH/bin, or move 'sbeads':
sudo mv ./sbeads /usr/local/bin/sbeads
```

---

## Configuration

Set the `SBEADS_DB` environment variable to control the location of the SQLite database:

```bash
# Default behavior (uses ./.sbeads)
sbeads ready

# Custom database file (e.g., shared across subagent worktrees)
export SBEADS_DB="/var/tmp/swarm-tasks.db"
sbeads ready
```

---

## CLI Command Reference

All subcommands support both explicit flags and positional shorthand where intuitive.

### 1. `create`
Creates a new issue in the tracker.

```bash
sbeads create [flags] [title]
```

**Options & Flags:**
- `--title <string>`: Issue title (required if positional title omitted).
- `--issue-type <string>`: Type category (e.g. `bug`, `feature`, `task`).
- `--priority <int>`: Priority rating (0–10, default: `0`). Must be non-negative.
- `--description <string>`: Detailed issue description or specification.
- `--blocked <0|1>`: Explicit blocked flag (`0` = unblocked, `1` = blocked). Default: `0`.

**Positional Shorthand:**
If `--title` is omitted, the first positional argument is treated as the title.

**Examples:**
```bash
# Positional title
sbeads create "Fix database reconnection timeout"

# Full flag specification
sbeads create \
  --title "Implement user auth token revocation" \
  --issue-type "feature" \
  --priority 2 \
  --description "Tokens must be immediately invalidated in SQLite cache upon logout"

# Creating a pre-blocked task
sbeads create "Deploy to staging" --blocked 1
```

**Output:**
```
Created issue: issue-4df45e43f1145ee2dfef2ec6c367adbf (Fix database reconnection timeout)
```

---

### 2. `ready`
Lists all actionable issues that are ready to be claimed or worked on.

```bash
sbeads ready
```

**Options & Flags:**
None.

**Output:**
```
issue-4df45e43f1145ee2dfef2ec6c367adbf: Fix database reconnection timeout
issue-a75d5e49392e92c4bbf332c02919aa78: Implement user auth token revocation
```
If no issues are ready:
```
No ready issues
```

---

### 3. `claim`
Claims an issue for an agent or developer. Sets `is_claimed = 1` and `is_blocked = 1` to prevent race conditions and concurrent double-claims.

```bash
sbeads claim [flags] [issue-id]
```

**Options & Flags:**
- `--issue-id <string>`: Issue ID to claim (required if positional ID omitted).
- `--agent <string>`: Name or identifier of claiming agent (default: `"me"`).
- `--assignee <string>`: Assignee username or role.

**Validation & Error Conditions:**
- Fails if the issue is already claimed (`issue already claimed`).
- Fails if the issue does not exist (`issue not found`).

**Examples:**
```bash
# Claim using positional ID
sbeads claim issue-4df45e43f1145ee2dfef2ec6c367adbf

# Claim with explicit agent name
sbeads claim --issue-id issue-4df45e43f1145ee2dfef2ec6c367adbf --agent worker-1 --assignee djs
```

**Output:**
```
Claimed issue: issue-4df45e43f1145ee2dfef2ec6c367adbf (Fix database reconnection timeout) by worker-1
```

---

### 4. `close`
Marks an open issue as resolved/closed (`is_closed = 1`).

```bash
sbeads close [flags] [issue-id]
```

**Options & Flags:**
- `--issue-id <string>`: Issue ID to close (required if positional ID omitted).

**Validation & Error Conditions:**
- Fails if the issue is already closed (`issue already closed`).
- Fails if the issue does not exist (`issue not found`).

**Examples:**
```bash
# Positional issue ID
sbeads close issue-4df45e43f1145ee2dfef2ec6c367adbf

# Flag-based
sbeads close --issue-id issue-4df45e43f1145ee2dfef2ec6c367adbf
```

**Output:**
```
Closed issue: issue-4df45e43f1145ee2dfef2ec6c367adbf (Fix database reconnection timeout)
```

---

### 5. `show`
Displays full details and metadata for a specific issue.

```bash
sbeads show [flags] [issue-id]
```

**Options & Flags:**
- `--issue-id <string>`: Issue ID to display (required if positional ID omitted).

**Examples:**
```bash
sbeads show issue-4df45e43f1145ee2dfef2ec6c367adbf
```

**Output:**
```
Issue: issue-4df45e43f1145ee2dfef2ec6c367adbf
Title: Fix database reconnection timeout
Type: bug
Priority: 2
Description: Tokens must be immediately invalidated in SQLite cache upon logout
Blocked: 1
Closed: 0
Claimed: 1
```

---

### 6. `dep` (Dependencies)
Manage dependency edges between issues.

#### `dep add`
Defines that a child issue depends on a parent issue (`child -> parent`). The child task cannot be started until the parent task is finished.

```bash
sbeads dep add [flags] [child-id] [parent-id]
```

**Options & Flags:**
- `--child <string>`: Child issue ID.
- `--parent <string>`: Parent issue ID.

**Validation & Error Conditions:**
- Self-dependency is disallowed: `child == parent` returns error `cannot depend on self`.
- Duplicate dependencies return error `dependency already exists`.

**Examples:**
```bash
# Using positional arguments (child parent)
sbeads dep add issue-child-id issue-parent-id

# Using flags
sbeads dep add --child issue-child-id --parent issue-parent-id
```

**Output:**
```
Added dependency: issue-child-id -> issue-parent-id
```

#### `dep list`
Lists all parent issues that a given issue depends on.

```bash
sbeads dep list [flags] [issue-id]
```

**Options & Flags:**
- `--issue-id <string>`: Issue ID to inspect (required if positional ID omitted).

**Examples:**
```bash
sbeads dep list issue-child-id
```

**Output:**
```
  issue-parent-id-1
  issue-parent-id-2
```
If no dependencies exist:
```
No dependencies
```

#### `dep remove`
Removes a dependency relationship between a child issue and a parent issue.

```bash
sbeads dep remove [flags] [child-id] [parent-id]
```

**Options & Flags:**
- `--child <string>`: Child issue ID.
- `--parent <string>`: Parent issue ID.

**Validation & Error Conditions:**
- Returns error `dependency not found` if the relation does not exist.

---

## Concrete Workflow Examples

### Workflow 1: Feature Decomposition & Dependency Tracking
Break down an architectural milestone into dependent subtasks:

```bash
# 1. Create parent prerequisite task (Database Schema)
PARENT_ID=$(sbeads create "Design SQLite WAL migration schema" --issue-type feature --priority 3 | awk '{print $3}')

# 2. Create child dependent task (Service Logic)
CHILD_ID=$(sbeads create "Implement storage service layer" --issue-type feature --priority 2 | awk '{print $3}')

# 3. Create integration testing task
TEST_ID=$(sbeads create "Write end-to-end integration tests" --issue-type task --priority 1 | awk '{print $3}')

# 4. Link dependencies:
# - Service logic depends on Schema
sbeads dep add "$CHILD_ID" "$PARENT_ID"

# - Tests depend on Service logic
sbeads dep add "$TEST_ID" "$CHILD_ID"

# 5. Check dependencies for integration tests
sbeads dep list "$TEST_ID"
```

### Workflow 2: Autonomous Agent Task Triage & Claim Cycle
An autonomous agent loop picking and finishing tasks:

```bash
# Step 1: Scan for ready tasks
sbeads ready

# Step 2: Agent selects the top ready issue and claims it
ISSUE_ID="issue-4df45e43f1145ee2dfef2ec6c367adbf"
sbeads claim --issue-id "$ISSUE_ID" --agent "worker-pnr-1"

# Step 3: Inspect detailed requirements
sbeads show "$ISSUE_ID"

# Step 4: [Agent executes code implementation and verifies tests]

# Step 5: Close task upon completion
sbeads close "$ISSUE_ID"

# Step 6: Query next ready work item
sbeads ready
```

---

## Running Tests

Run the full Go test suite across all layers:

```bash
# Run all unit and integration tests
go test -v ./...

# Run tests with race detection enabled
go test -race -v ./...

# Run targeted CLI routing tests
go test -v ./cmd/bd/...
```

All tests pass using SQLite in-memory and temporary directory files with clean tear-down.

---

## License

MIT License. See LICENSE for details.
