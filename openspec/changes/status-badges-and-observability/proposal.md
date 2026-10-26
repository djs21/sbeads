## Why

While `sbeads` now benefits from a Cobra DD-VSA CLI structure, dependency-aware ready filtering, and 3-character Base36 IDs, workflow observability and multi-agent coordination remain severely bottlenecked by four core visibility gaps:

1. **Absence of Status Indicators**: `sbeads ready` outputs only `<id>: <title>`. Neither autonomous AI agents nor human developers can discern issue priority, classification type, or state without executing secondary `show` commands for each item.
2. **Missing Catalog Visibility (`sbeads list`)**: When tasks are closed or blocked, they completely disappear from `sbeads ready` with no high-level command to view the overarching backlog, inspect completed work, or view hierarchical parent-child relationships.
3. **Crude Binary Inspection in `sbeads show`**: `sbeads show` prints raw database bits (`Blocked: 0`, `Closed: 1`, `Claimed: 0`) rather than a single definitive computed lifecycle state, and completely omits blocker/dependent relations and agent assignment.
4. **Transient Agent Coordination & No Release Mechanism**: While `ClaimIssue` accepts `agent` and `assignee`, the SQLite `issues` table lacks persistent columns for them. Additionally, there is no `unclaim` command to release claimed tasks when an agent terminates, crashes, or yields.

Implementing status badges (`[O]`, `[C]`, `[B]`, `[X]`, `[EPIC]`), persistent agent/assignee storage, `sbeads list` (with filtering and hierarchical tree grouping), `sbeads unclaim`, and context-rich `show` rendering provides complete end-to-end task observability for both human operators and LLM swarm agents.

## What Changes

- **Schema Migration & Persistence**: Add `agent TEXT DEFAULT ''` and `assignee TEXT DEFAULT ''` columns to SQLite `issues` table using idempotent schema migration. Update `types.Issue`, `ClaimIssue`, and introduce `UnclaimIssue` in `backend/storage.go`.
- **Status Badge Engine (`internal/domain/status.go`)**: Implement badge evaluation:
  - `[O]` : Open (unclaimed, unclosed, unblocked).
  - `[C]` : Claimed / In-progress (`is_claimed = 1`).
  - `[B]` : Blocked (explicitly blocked or waiting on unclosed parent dependencies).
  - `[X]` : Closed / Done (`is_closed = 1`).
  - `[EPIC]` : Prominent badge when `issue_type = "epic"`.
  - Priority badge `(P0..P10)` and type tag.
- **Enhanced `ready` Command Slice (`cmd/bd/ready.go`)**: Render rich one-liner format: `<id> [O] (P1/task) <title>` with `--all` flag to optionally view claimed items.
- **Detailed `show` Command Slice (`cmd/bd/show.go`)**: Replace raw integer flags with a single computed status (`Status: CLAIMED (by @worker-1)`), claiming agent/assignee info, and direct blocker (`Blocked by:`) and dependent (`Blocks:`) tree edges.
- **New Subcommand `sbeads list` (`cmd/bd/list.go`)**: Provide catalog overview supporting `--all`, `--closed`, `--blocked`, `--json`, and hierarchical tree indent under epics.
- **New Subcommand `sbeads unclaim` (`cmd/bd/unclaim.go`)**: Enable releasing claimed tasks back to `[O]` state.

## Capabilities

### New Capabilities
- `status-engine`: Unified badge evaluator (`[O]`, `[C]`, `[B]`, `[X]`, `[EPIC]`) and one-liner string formatter for CLI commands.
- `cli-list`: Subcommand `sbeads list` displaying the full issue catalog with status filtering and hierarchical tree visualization.
- `cli-unclaim`: Subcommand `sbeads unclaim <id>` allowing agents and users to release claimed issues back to open queue.

### Modified Capabilities
- `cli-ready`: Outputs rich status and priority badges (`<id> [O] (P1/task) <title>`).
- `cli-show`: Outputs computed status, agent assignment, and blocker/dependent dependency edges.
- `storage-schema`: SQLite `issues` schema adds persistent `agent` and `assignee` fields, with storage methods `UnclaimIssue` and updated `ClaimIssue`.

## Impact

- **Database**: Adds `agent` and `assignee` columns to `issues` table via `ALTER TABLE` backward-compatible migration.
- **CLI Commands**: Modifies `ready` and `show`, introduces `list` and `unclaim`.
- **Architecture**: Enforces Domain-Driven Vertical Slice Architecture (DD-VSA) under `cmd/bd/` strictly under 400 LOC per file.
- **AI Agent Workflow**: Enables autonomous LLM agents to pick highest-priority ready tasks with zero extra roundtrips and release tasks seamlessly.
