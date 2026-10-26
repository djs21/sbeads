# sbeads

`sbeads` is an ultra-lightweight, high-performance issue tracker and task coordinator designed for autonomous AI agents and software engineers. Built in Go on top of SQLite with Write-Ahead Logging (WAL), `sbeads` delivers zero-friction local task tracking, explicit dependency graphs, and atomic agent claim/close operations without external services or heavy daemon overhead.

---

## Overview & Purpose

Modern autonomous agent swarms and multi-agent development workflows require structured, reliable task coordination:
- **Zero Overhead**: Single static binary with embedded SQLite storage. No server daemon, Docker container, or cloud service required.
- **Agent-First Coordination**: Explicit claim semantics (`sbeads claim`) prevent duplicate work across concurrent agent panes or subagents.
- **Dependency Graphs**: First-class parent/child issue dependencies (`sbeads dep add`) block dependent tasks until prerequisite tasks finish.
- **Fast Ready Queues**: Query unblocked, actionable issues instantly (`sbeads ready`) to drive autonomous loops.
- **Deterministic Storage**: Local SQLite database stored at `.sbeads/sbeads.db`. The containing `.sbeads/` directory is automatically created.
- **Smart Ancestor Discovery**: Like git or beads, `sbeads` walks upwards from nested subdirectories to discover the nearest `.sbeads/` directory, letting you run CLI commands anywhere inside your project tree.

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

### Storage Engine: Pure-Go SQLite in WAL Mode
- **Database Location (`.sbeads/sbeads.db`)**: By default, issues and dependencies are stored in `.sbeads/sbeads.db`. Both `NewStorage` and `CreateTables` automatically invoke `os.MkdirAll` to create `.sbeads/` (with permissions `0755`) if it does not exist.
- **Pure-Go Architecture (`CGO_ENABLED=0`)**: Backed by `modernc.org/sqlite`, `sbeads` requires zero C toolchains (`gcc`, `clang`, or `musl`) and compiles into 100% static, cross-platform standalone binaries for Linux, macOS, and Windows.
- **Ancestor Discovery**: When executed from any nested subdirectory, `sbeads` searches upward through parent directories to locate an existing `.sbeads/` directory, resolving to `<foundDir>/sbeads.db`.
- **WAL Mode (`_journal_mode=WAL&_busy_timeout=5000`)**: Connections open SQLite with Write-Ahead Logging and a 5,000ms busy timeout enabled. This allows concurrent readers, non-blocking reads while writes occur, and graceful lock acquisition, making it robust for concurrent agent swarms and subagents.
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
| `agent`      | `TEXT`             | Name or identifier of claiming agent (e.g. `@worker-1`) |
| `assignee`   | `TEXT`             | Optional assignee username or role |
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

## Status Badges & Observability

`sbeads` provides high-density observability badges designed for quick visual scanning by engineers and precise parsing by autonomous AI agents.

### Status Badge Indicators

| Badge | Status | Description |
|---|---|---|
| `[O]` | `OPEN` | Ready to be worked on; unblocked with no unclosed parent dependencies. |
| `[C]` | `CLAIMED` | Actively claimed and being executed by an agent or assignee (`[@agent]`). |
| `[B]` | `BLOCKED` | Blocked explicitly (`is_blocked = 1`) or waiting on unclosed parent dependencies. |
| `[X]` | `CLOSED` | Task has been successfully resolved or closed (`is_closed = 1`). |
| `[EPIC]` | *(type)* | Marker tag identifying parent epic issues containing subtasks. |

### Status Computation Precedence

The effective status of an issue is dynamically computed from its flags and dependency tree:
1. **`CLOSED` (`[X]`)**: `is_closed != 0`
2. **`BLOCKED` (`[B]`)**: `is_blocked != 0` OR any prerequisite parent dependency is unclosed.
3. **`CLAIMED` (`[C]`)**: `is_claimed != 0`
4. **`OPEN` (`[O]`)**: Default state when open, unclaimed, and unblocked.

---

## Installation & Build

### Prerequisites
- Go 1.27+ (or Go 1.21+)
- Zero external C toolchains or GCC required (100% pure-Go SQLite via `modernc.org/sqlite`, fully static binaries)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/djs21/sbeads.git
cd sbeads

# Build using build script
./build.sh --dev     # Development binary with git commit
./build.sh --prod    # Production stripped binary with git tag

# Or build directly with go
go build -o sbeads ./cmd/bd

# Verify installation
./sbeads version
```

To install system-wide or into `$GOPATH/bin`:
```bash
go install ./cmd/bd
# Binary will be named 'bd' in $GOPATH/bin, or move 'sbeads':
sudo mv ./sbeads /usr/local/bin/sbeads
```

---

## Configuration & Database Discovery

`sbeads` resolves its database path using the standard beads search order:

1. **`SBEADS_DIR`**: If set, `sbeads` uses this directory (appending `sbeads.db`) or uses the direct path if pointing directly to a `.db` file.
2. **`SBEADS_DB`**: If set, `sbeads` uses this database file path directly (or appends `sbeads.db` if pointing to a directory).
3. **Ancestor Search**: Walks upwards from the current working directory looking for a directory named `.sbeads`. If found, uses `<foundDir>/sbeads.db`.
4. **Default**: Returns `.sbeads/sbeads.db` in the current working directory.

### Examples

```bash
# 1. Default / Ancestor discovery (uses .sbeads/sbeads.db in project root or current dir)
cd /my/project/nested/subpkg
sbeads ready  # automatically finds /my/project/.sbeads/sbeads.db!

# 2. Directory override using SBEADS_DIR
export SBEADS_DIR="/var/tmp/sbeads-storage"
sbeads ready  # uses /var/tmp/sbeads-storage/sbeads.db

# 3. Direct DB file override using SBEADS_DB
export SBEADS_DB="/shared/path/swarm.db"
sbeads ready  # uses /shared/path/swarm.db directly
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
- `-t`, `--title <string>`: Issue title (required if positional title omitted).
- `-y`, `--type`, `--issue-type <string>`: Type category (e.g. `bug`, `feature`, `task`). Default: `task`.
- `-p`, `--priority <int>`: Priority rating (0–10, default: `0`). Must be non-negative.
- `-d`, `--description <string>`: Detailed issue description or specification.
- `-b`, `--blocked <0|1>`: Explicit blocked flag (`0` = unblocked, `1` = blocked). Default: `0`.
- `-h`, `--help`: Help for create command.

**Positional Shorthand:**
If `--title` is omitted, the first positional argument is treated as the title.

**Examples:**
```bash
# Positional title
sbeads create "Fix database reconnection timeout"

# Short flags specification
sbeads create -t "Implement user auth token revocation" -y "feature" -p 2 -d "Tokens must be immediately invalidated"

# Full flag specification
sbeads create \
  --title "Implement user auth token revocation" \
  --issue-type "feature" \
  --priority 2 \
  --description "Tokens must be immediately invalidated in SQLite cache upon logout"

# Creating a pre-blocked task
sbeads create "Deploy to staging" -b 1
```

**Output:**
```
Created issue: sbeads-k7a (Fix database reconnection timeout)
```

---

### 2. `ready`
Lists all actionable issues that are ready to be claimed or worked on (unblocked, unclaimed, and unclosed). Outputs a standardized one-liner format with status badges, priority tag `(P<pri>/<type>)`, and optional `[EPIC]` tag.

```bash
sbeads ready
```

**Options & Flags:**
None.

**Output Format:**
```
<id> [<STATUS>] <[EPIC]> (P<priority>/<type>) <title> <[@agent]>
```

**Examples:**
```bash
sbeads ready
```

**Output:**
```
sbeads-k7a [OPEN] (P2/bug) Fix database reconnection timeout
sbeads-3qt [OPEN] [EPIC] (P1/epic) Epic: Status Badges, Observability, and Coordination Lifecycle
```
If no issues are ready:
```
No ready issues
```

---

### 3. `list`
Lists issues with comprehensive filtering and hierarchical tree visualization.

```bash
sbeads list [flags]
```

**Options & Flags:**
- `-a`, `--all`: Show all issues (including closed).
- `-b`, `--blocked`: Show only blocked issues (explicitly blocked or waiting on unclosed parent dependencies).
- `-c`, `--closed`: Show only closed issues.
- `-t`, `--tree`: Show hierarchical tree view (child tasks are indented 2 spaces under their parents/epics).
- `--json`: Output in machine-readable JSON format.
- `-h`, `--help`: Help for list command.

**Default Behavior:**
When run without filter flags (`-a`, `-b`, `-c`), `sbeads list` displays all unclosed issues (`is_closed == 0`).

**Examples:**
```bash
# Default: list unclosed issues
sbeads list

# Hierarchical dependency tree view
sbeads list --tree

# List all issues including completed/closed
sbeads list -a

# Filter only blocked issues
sbeads list --blocked

# Filter only closed issues
sbeads list --closed

# Machine-readable JSON output
sbeads list --json
```

**Output (Flat):**
```
sbeads-k7a [OPEN] (P2/bug) Fix database reconnection timeout
sbeads-4z9 [CLAIMED] (P2/feature) Implement user auth token revocation [@worker-1]
sbeads-b8x [BLOCKED] (P1/task) Deploy authentication service
```

**Output (Tree `--tree`):**
```
sbeads-3qt [OPEN] [EPIC] (P1/epic) Epic: Status Badges, Observability, and Coordination Lifecycle
  sbeads-k7a [OPEN] (P2/bug) Fix database reconnection timeout
  sbeads-b8x [BLOCKED] (P1/task) Deploy authentication service
```

---

### 4. `claim`
Claims an issue for an agent or developer. Sets `is_claimed = 1` and `is_blocked = 1` to prevent race conditions and concurrent double-claims across agent panes.

```bash
sbeads claim [flags] [issue-id]
```

**Options & Flags:**
- `-i`, `--issue-id <string>`: Issue ID to claim (required if positional ID omitted).
- `-m`, `--agent <string>`: Name or identifier of claiming agent (default: `"me"`).
- `-a`, `--assignee <string>`: Assignee username or role.
- `-h`, `--help`: Help for claim command.

**Validation & Error Conditions:**
- Fails if the issue is already claimed (`issue already claimed`).
- Fails if the issue does not exist (`issue not found`).

**Examples:**
```bash
# Claim using positional ID (or short hash) with agent identifier
sbeads claim sbeads-k7a --agent worker-1

# Positional claim with default agent ("me")
sbeads claim sbeads-k7a

# Claim with explicit flags
sbeads claim --issue-id sbeads-k7a --agent worker-1 --assignee djs
```

**Output:**
```
Claimed issue: sbeads-k7a (Fix database reconnection timeout) by worker-1
```

---

### 5. `unclaim`
Releases a previous claim on an issue. Clears the assigned agent and assignee, resets `is_claimed = 0`, and unblocks the issue (`is_blocked = 0`) unless it remains blocked by unclosed parent dependencies.

```bash
sbeads unclaim [flags] [issue-id]
```

**Options & Flags:**
- `-i`, `--issue-id <string>`: Issue ID to unclaim (required if positional ID omitted).
- `-h`, `--help`: Help for unclaim command.

**Validation & Error Conditions:**
- Fails if the issue does not exist (`issue not found`).

**Examples:**
```bash
# Positional issue ID
sbeads unclaim sbeads-k7a

# Flag-based
sbeads unclaim --issue-id sbeads-k7a
```

**Output:**
```
Unclaimed issue: sbeads-k7a (Fix database reconnection timeout)
```

---

### 6. `close`
Marks an open issue as resolved/closed (`is_closed = 1`).

```bash
sbeads close [flags] [issue-id]
```

**Options & Flags:**
- `-i`, `--issue-id <string>`: Issue ID to close (required if positional ID omitted).

**Validation & Error Conditions:**
- Fails if the issue is already closed (`issue already closed`).
- Fails if the issue does not exist (`issue not found`).

**Examples:**
```bash
# Positional issue ID
sbeads close sbeads-k7a

# Flag-based
sbeads close --issue-id sbeads-k7a
```

**Output:**
```
Closed issue: sbeads-k7a (Fix database reconnection timeout)
```

---

### 7. `show`
Displays full details, computed lifecycle status, agent metadata, and bidirectional dependency relationships (`Blocked by:` and `Blocks:`) for a specific issue.

```bash
sbeads show [flags] [issue-id]
```

**Options & Flags:**
- `-i`, `--issue-id <string>`: Issue ID to display (required if positional ID omitted).

**Examples:**
```bash
sbeads show sbeads-k7a
```

**Output:**
```
Issue: sbeads-k7a
Title: Fix database reconnection timeout
Type: bug
Priority: 2
Status: OPEN [O]
Agent: @worker-1
Description: Tokens must be immediately invalidated in SQLite cache upon logout
Blocked by:
  sbeads-3qt: Epic: Status Badges, Observability, and Coordination Lifecycle
Blocks:
  sbeads-b8x: Deploy authentication service
```

---

### 8. `dep` (Dependencies)
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

### 9. `version`
Prints the application version string.

```bash
sbeads version
```

**Output:**
```
sbeads version dev-8fe518a
```

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

# 6. Inspect overall dependency tree hierarchy
sbeads list --tree
```

### Workflow 2: Autonomous Agent Task Triage & Claim Cycle
An autonomous agent loop picking, executing, or releasing tasks:

```bash
# Step 1: Scan for ready tasks
sbeads ready
# Output:
# sbeads-k7a [OPEN] (P2/task) Fix database reconnection timeout

# Step 2: Agent selects the top ready issue and claims it
ISSUE_ID="sbeads-k7a"
sbeads claim "$ISSUE_ID" --agent "worker-pnr-1"

# Step 3: Inspect detailed requirements and blockers
sbeads show "$ISSUE_ID"

# Step 4a: [If task execution succeeds]
# Run tests and verification, then close:
sbeads close "$ISSUE_ID"

# Step 4b: [If task must be aborted or handed off]
# Release claim so other agents can pick it up:
sbeads unclaim "$ISSUE_ID"

# Step 5: Query next ready work item
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
