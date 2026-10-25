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
[Claim Issue] (sbeads claim --agent <name>)
      │  └─ If already claimed: Pick another ready issue
      ▼
[Execute Task]
      │
      ▼
[Close Issue] (sbeads close <id>)
      │
      ▼
[Check Next Ready] (sbeads ready)
```

### Core Invariants
1. **Always Claim Before Execution**: Run `sbeads claim <id> --agent <agent-id>` before modifying code or performing work. If claiming fails, do not proceed.
2. **Atomic Completion**: Only run `sbeads close <id>` after tests, linting, or verification pass.
3. **Dependencies Block Ready State**: When Task B depends on Task A (`dep add <TaskB> <TaskA>`), Task B will not be worked on until Task A is resolved.

---

## CLI Recipes & Invocation Cheatsheet

### 1. Creating Issues
```bash
# Basic task creation
sbeads create "Implement user session timeout"

# Detailed task creation with priority and classification
sbeads create \
  --title "Fix memory leak in buffer pool" \
  --issue-type "bug" \
  --priority 3 \
  --description "Buffer slice retains old references causing GC retention"

# Capturing the generated Issue ID in bash
ISSUE_ID=$(sbeads create "Draft API specification" | grep -o 'issue-[a-f0-9]\{32\}')
echo "New issue: $ISSUE_ID"
```

### 2. Checking Ready Issues
```bash
# List ready issues
sbeads ready

# Example output:
# issue-8b209d80dcf2f205a2e7c413b860b299: Draft API specification
# issue-e970a2f1ca0b8a248fca8b99c855a822: Implement user session timeout

# Scriptable: Get first ready issue ID
FIRST_READY=$(sbeads ready | grep -E '^issue-[a-f0-9]{32}' | head -n 1 | cut -d: -f1)
```

### 3. Claiming an Issue
```bash
# Claim by issue ID with agent identity
sbeads claim --issue-id "$ISSUE_ID" --agent "worker-agent"

# Positional claim
sbeads claim "$ISSUE_ID" --agent "worker-agent"
```

### 4. Inspecting Issue Details
```bash
sbeads show "$ISSUE_ID"

# Parse fields from output:
# Issue: issue-...
# Title: ...
# Type: ...
# Priority: ...
# Description: ...
# Blocked: 0|1
# Closed: 0|1
# Claimed: 0|1
```

### 5. Managing Dependencies
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

### 6. Closing Completed Issues
```bash
sbeads close "$ISSUE_ID"
```

---

## Output Parsing Patterns

### Extracting Issue IDs
All `sbeads` issue IDs follow the regular expression:
```regex
issue-[0-9a-f]{32}
```

- From `sbeads create`:
  ```bash
  # Output: Created issue: issue-4df45e43f1145ee2dfef2ec6c367adbf (Title)
  ID=$(echo "$OUTPUT" | sed -n 's/.*Created issue: \([a-z0-9-]*\).*/\1/p')
  ```

- From `sbeads ready`:
  ```bash
  # Output: issue-4df45e43f1145ee2dfef2ec6c367adbf: Title
  ID=$(echo "$LINE" | cut -d: -f1)
  TITLE=$(echo "$LINE" | cut -d: -f2- | sed 's/^ //')
  ```

- From `sbeads show`:
  ```bash
  # Key-Value format:
  TITLE=$(echo "$SHOW_OUTPUT" | awk -F': ' '$1=="Title" {print $2}')
  BLOCKED=$(echo "$SHOW_OUTPUT" | awk -F': ' '$1=="Blocked" {print $2}')
  CLAIMED=$(echo "$SHOW_OUTPUT" | awk -F': ' '$1=="Claimed" {print $2}')
  CLOSED=$(echo "$SHOW_OUTPUT" | awk -F': ' '$1=="Closed" {print $2}')
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
3. **Tear-Down Hygiene**: If an agent aborts or fails a task, keep state clean. If work was not finished, unclaim or document progress in issue description or logs before terminating.
