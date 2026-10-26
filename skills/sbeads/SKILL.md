---
name: sbeads
description: Lightweight SQLite-backed issue tracker and task coordinator for AI agents and developers. Use when managing tasks, breaking work into subtasks or bugs, coordinating multi-agent queues, tracking dependencies, or when the user mentions sbeads, bd, or issue tracking.
---

# sbeads Agent Skill

`sbeads` (binary `sbeads` or `bd`) is a local, fast, SQLite-backed task coordinator. Built with pure-Go SQLite (`modernc.org/sqlite`, `CGO_ENABLED=0`), it compiles into fully static binaries requiring zero C toolchains. It uses WAL mode (`_journal_mode=WAL&_busy_timeout=5000`) for concurrency, explicit claim semantics to prevent race conditions among agents, and parent/child dependency management.

---

## When to Use `sbeads`

Use `sbeads` when:
- Breaking a complex user request or milestone into ordered subtasks.
- Coordinating work across multiple subagents (`worker`, `pnr`, `reviewer`) where tasks must be claimed, tracked, and closed.
- Managing bug backlogs or feature roadmaps locally without network or cloud issue trackers.
- Explicit task dependency tracking is required (e.g., Task B cannot start until Task A completes).

---

## Binary & Environment Setup

Ensure the `sbeads` binary is built or located on your PATH:
```bash
# If sbeads is not in PATH, build it using build.sh
./build.sh --dev
# or directly:
go build -o sbeads ./cmd/bd
```

### Database Storage & Discovery
By default, `sbeads` stores issues in `.sbeads/sbeads.db` (creating the `.sbeads/` directory automatically with `0755` permissions).

It adopts the **beads discovery pattern**:
1. **`SBEADS_DIR`**: Explicit directory path containing or pointing to `sbeads.db`.
2. **`SBEADS_DB`**: Explicit database file path.
3. **Ancestor Search**: Automatically walks up parent directories from the current working directory searching for `.sbeads/`, finding `<ancestor>/.sbeads/sbeads.db`. This allows subagents running in nested source directories or subpackages to seamlessly interact with the project tracker without setting environment variables.
4. **Default**: Falls back to `.sbeads/sbeads.db` in the current working directory.

When running concurrent subagents across multiple git worktrees or shared panes, export `SBEADS_DIR` or `SBEADS_DB` to point to a centralized location:
```bash
export SBEADS_DIR="/home/djs/project/sbeads/.sbeads"
# or
export SBEADS_DB="/home/djs/project/sbeads/.sbeads/sbeads.db"
```

---

## Agent Protocol & State Machine

Agents must strictly follow the lifecycle state machine:

```
[Create Issue]
      │
      ▼
[Ready Queue] (sbeads ready)
      │
      ▼
[Claim Issue] (sbeads claim <id> --agent <name>)
      │  └─ If already claimed: Pick another ready issue
      ▼
[Execute Task]
      │
      ├── (if failed/aborted) ──> [Unclaim Issue] (sbeads unclaim <id>) ──> [Ready Queue]
      ▼
[Close Issue] (sbeads close <id>)
      │
      ▼
[Check Next Ready] (sbeads ready)
```

### Status Badge Lifecycle & Observability

`sbeads` outputs standardized badges for instant state evaluation:

| Badge | Status | Description | Action for LLM / Subagent |
|---|---|---|---|
| `[O]` | `OPEN` | Ready to be worked on; no unclosed parent dependencies. | Candidate to claim with `sbeads claim <id> --agent <name>`. |
| `[C]` | `CLAIMED` | Actively claimed by an agent or assignee (`[@agent]`). | Do not touch unless assigned to you. |
| `[B]` | `BLOCKED` | Blocked explicitly or waiting on unclosed parent tasks. | Do not attempt to claim; inspect blockers with `sbeads show <id>`. |
| `[X]` | `CLOSED` | Resolved / completed. | Finished work. |
| `[EPIC]` | *(type)* | Marker identifying a parent epic issue. | Inspect subtasks with `sbeads list --tree`. |

### Core Invariants
1. **Always Claim Before Execution**: Run `sbeads claim <id> --agent <agent-id>` before modifying code or performing work. If claiming fails, do not proceed.
2. **Context Inspection via Trees**: Before planning, run `sbeads list --tree` to understand parent epics and child subtask relationships.
3. **Unclaim on Abort or Failure**: If you encounter a blocking uncertainty, failure, or need to yield work, run `sbeads unclaim <id>` so another agent or master can reassign it.
4. **Atomic Completion**: Only run `sbeads close <id>` after tests, linting, or verification pass.
5. **Dependencies Block Ready State**: When Task B depends on Task A (`dep add <TaskB> <TaskA>`), Task B will not appear in `sbeads ready` until Task A is closed.

---

## CLI Recipes & Invocation Cheatsheet

### 1. Creating Issues
```bash
# Basic task creation (positional or short flag)
sbeads create "Implement user session timeout"
sbeads create -t "Implement user session timeout"

# Detailed task creation with priority and classification
sbeads create \
  -t "Fix memory leak in buffer pool" \
  -y "bug" \
  -p 3 \
  -d "Buffer slice retains old references causing GC retention"

# Capturing the generated Issue ID in bash (3-character Base36 ID, e.g. sbeads-abc)
ISSUE_ID=$(sbeads create "Draft API specification" | grep -o '[a-z0-9_-]*-[0-9a-z]\{3\}')
echo "New issue: $ISSUE_ID"
```

### 2. Checking Ready Issues
```bash
# List ready issues (outputs one-liner format with status badges)
sbeads ready

# Example output:
# sbeads-k7a [OPEN] (P2/task) Draft API specification
# sbeads-3qt [OPEN] [EPIC] (P1/epic) Epic: Coordination Architecture
# sbeads-4z9 [OPEN] (P3/bug) Fix memory leak in buffer pool

# Scriptable: Get first ready issue ID (extracts the ID before space or badge)
FIRST_READY=$(sbeads ready | awk '{print $1}' | head -n 1)
```

### 3. Inspecting Project Hierarchy & Lists
```bash
# Inspect full hierarchical dependency tree (child tasks indented under parents/epics)
sbeads list --tree

# List all open (unclosed) issues (default)
sbeads list

# List all issues including closed
sbeads list -a

# Filter only blocked issues
sbeads list --blocked

# Filter only closed issues
sbeads list --closed

# Machine-readable JSON output for automated scripting
sbeads list --json
```

### 4. Claiming an Issue
```bash
# Positional claim with agent identifier (preferred)
sbeads claim "$ISSUE_ID" --agent "worker-agent"

# Claim with explicit flag
sbeads claim --issue-id "$ISSUE_ID" --agent "worker-agent"

# Claim with assignee
sbeads claim "$ISSUE_ID" --agent "worker-agent" --assignee "djs"
```

### 5. Unclaiming an Issue (Aborting / Releasing)
```bash
# Release claim so another agent or developer can pick it up
sbeads unclaim "$ISSUE_ID"

# Flag-based
sbeads unclaim --issue-id "$ISSUE_ID"
```

### 6. Inspecting Issue Details
```bash
sbeads show "$ISSUE_ID"

# Output format:
# Issue: sbeads-k7a
# Title: Draft API specification
# Type: task
# Priority: 2
# Status: OPEN [O]
# Agent: @worker-agent
# Description: ...
# Blocked by:
#   sbeads-3qt: Epic: Coordination Architecture
# Blocks:
#   sbeads-b8x: Deploy service
```

### 7. Managing Dependencies
```bash
# Add dependency: child depends on parent
# child cannot start until parent finishes
sbeads dep add --child "$CHILD_ISSUE_ID" --parent "$PARENT_ISSUE_ID"
# Or positional:
sbeads dep add "$CHILD_ISSUE_ID" "$PARENT_ISSUE_ID"

# Inspect dependencies of an issue
sbeads dep list "$CHILD_ISSUE_ID"

# Remove dependency when no longer blocking
sbeads dep remove "$CHILD_ISSUE_ID" "$PARENT_ISSUE_ID"
```

### 8. Closing Completed Issues
```bash
sbeads close "$ISSUE_ID"
```

---

## Output Parsing Patterns

### Extracting Issue IDs
All `sbeads` issue IDs follow the regular expression (3-character Base36 hash):
```regex
[a-z0-9_-]+-[0-9a-z]{3}
```
Prefix lookup is supported: you can query using exact ID `sbeads-abc` or short hash `abc`!

- From `sbeads create`:
  ```bash
  # Output: Created issue: sbeads-k7a (Title)
  ID=$(echo "$OUTPUT" | sed -n 's/.*Created issue: \([a-z0-9-]*\).*/\1/p')
  ```

- From `sbeads ready`:
  ```bash
  # Output: sbeads-k7a [OPEN] (P2/task) Draft API specification
  ID=$(echo "$LINE" | awk '{print $1}')
  STATUS=$(echo "$LINE" | awk '{print $2}')
  ```

- From `sbeads show`:
  ```bash
  # Key-Value format:
  TITLE=$(echo "$SHOW_OUTPUT" | awk -F': ' '$1=="Title" {print $2}')
  STATUS=$(echo "$SHOW_OUTPUT" | awk -F': ' '$1=="Status" {print $2}')
  AGENT=$(echo "$SHOW_OUTPUT" | awk -F': ' '$1=="Agent" {print $2}')
  ```

---

## Dependency Coordination Patterns

### Pattern A: Linear Chain (Sequential Tasks)
```
Step 1 ──> Step 2 ──> Step 3
```
```bash
ID1=$(sbeads create "Step 1: Database Migration" | awk '{print $3}')
ID2=$(sbeads create "Step 2: API Endpoints" | awk '{print $3}')
ID3=$(sbeads create "Step 3: Frontend Client" | awk '{print $3}')

# ID2 depends on ID1; ID3 depends on ID2
sbeads dep add "$ID2" "$ID1"
sbeads dep add "$ID3" "$ID2"
```

### Pattern B: Fan-Out / Fan-In (Diamond Dependency)
```
         ┌──> Subtask A ──┐
Setup ───┤                ├──> Integration Verification
         └──> Subtask B ──┘
```
```bash
SETUP=$(sbeads create "Setup baseline configuration" | awk '{print $3}')
TASK_A=$(sbeads create "Implement Backend Service" | awk '{print $3}')
TASK_B=$(sbeads create "Implement Worker Queue" | awk '{print $3}')
VERIFY=$(sbeads create "Run Integration Suite" | awk '{print $3}')

sbeads dep add "$TASK_A" "$SETUP"
sbeads dep add "$TASK_B" "$SETUP"
sbeads dep add "$VERIFY" "$TASK_A"
sbeads dep add "$VERIFY" "$TASK_B"
```

---

## Error Recovery & Conflict Handling

| Error Output | Root Cause | Recovery Strategy |
|---|---|---|
| `issue already claimed` | Another subagent claimed this issue first | Do NOT proceed with this issue. Re-run `sbeads ready` and select the next available unclaimed issue. |
| `issue already closed` | Task was already completed | Verify completion status via `sbeads show <id>`. Move to next task. |
| `issue not found: <id>` | Invalid or mistyped Issue ID | Verify ID using `sbeads ready` or git session logs. Do not fabricate IDs. |
| `cannot depend on self` | Passing the same ID as child and parent | Check script variables; child and parent must be distinct. |
| `dependency already exists`| Edge already exists in graph | Safe to ignore or continue; the dependency relationship is already established. |
| `--title is required` | Missing title flag or argument | Provide title either via `--title "<text>"` or as first positional argument. |
| `priority must be non-negative` | Negative priority value | Ensure priority is an integer >= 0 (standard range 0–10). |

---

## Concurrency Best Practices for Multi-Agent Swarms

1. **Leverage WAL Mode**: SQLite runs in WAL mode with low contention. Short CLI calls ensure lock hold times are milliseconds.
2. **Check-Then-Claim Is Atomic in Storage**: `sbeads claim` updates both `is_claimed = 1` and `is_blocked = 1` within an atomic transaction. If the command succeeds (exit code 0), your agent exclusively owns the task.
3. **Tear-Down & Abort Hygiene**: If an agent aborts, encounters an unresolvable blocker, or fails a task, release it immediately using `sbeads unclaim <id>` so the issue reverts to open status and other swarm workers can claim it.
