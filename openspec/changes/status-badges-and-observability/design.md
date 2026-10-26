## Context

The `sbeads` issue tracker serves as an embedded, CGO-free task management engine for both human developers and autonomous AI swarms. Following the CLI modernization to Cobra and 3-char Base36 IDs, developers and agents require high-density, low-latency visibility into task states, priority ordering, and assignment ownership.

Currently, once tasks are closed or blocked, they disappear from `ready` with no global viewing mechanism. Furthermore, `show` outputs disconnected binary flags (`Closed: 0`, `Claimed: 0`, `Blocked: 0`) without resolving the computed state, and agents cannot see dependency trees without executing multiple calls. Lastly, multi-agent coordination requires tracking which agent claimed what task and the ability to release claims upon task preemption or agent failure.

## Goals / Non-Goals

**Goals:**
- Implement durable SQLite migration adding `agent` and `assignee` columns to `issues` table with backward compatibility for existing databases.
- Implement `UnclaimIssue(id string)` in `backend/storage.go` and expose via domain layer.
- Build a dedicated status evaluation domain slice (`internal/domain/status.go`) with standard badges:
  - `[O]` (Open, Ready to work)
  - `[C]` (Claimed, In-progress)
  - `[B]` (Blocked by dependency or explicit flag)
  - `[X]` (Closed / Done)
  - `[EPIC]` (Prominent marker for epic issues)
  - Priority indicator `(P<priority>/<type>)`
- Implement `sbeads list` vertical slice (`cmd/bd/list.go`) supporting `--all`, `--closed`, `--blocked`, `--json`, and hierarchical tree indent.
- Implement `sbeads unclaim` vertical slice (`cmd/bd/unclaim.go`) supporting flags and positional arguments.
- Modernize `ready` output formatting to use status badges and priority tags.
- Enhance `show` output to display computed state, claiming identity, and full blocker/dependent relationships.
- Enforce Domain-Driven Vertical Slice Architecture (DD-VSA) with strict < 400 LOC ceiling per file across all code and test files.

**Non-Goals:**
- Networked multi-user authentication or RBAC (remains local SQLite).
- Arbitrary custom tagging engine beyond `issue_type` and priority.
- External webhook notifications.

## Decisions

### Decision 1: Idempotent Schema Migration in `CreateTables`
- **Choice**: Execute `ALTER TABLE issues ADD COLUMN agent TEXT DEFAULT '';` and `ALTER TABLE issues ADD COLUMN assignee TEXT DEFAULT '';` safely inside `backend.CreateTables`, ignoring `duplicate column name` error.
- **Rationale**: Preserves zero-configuration simplicity; existing `.sbeads/sbeads.db` databases automatically migrate on first launch without requiring external migration tooling.

### Decision 2: Status Evaluation in Domain Layer (`internal/domain/status.go`)
- **Choice**: Encapsulate status badge logic inside `internal/domain/status.go` rather than hardcoding formatting in CLI commands.
- **Rationale**: Keeps CLI slices strictly as presentation layers and enables unit testing of status resolution (checking open vs blocked vs claimed) against mock or in-memory models.

### Decision 3: Standard One-Liner Format for CLI Lists
- **Choice**: Format one-line issue representations as:
  `<id> [STATUS] [EPIC?] (P<priority>/<type>) <title> <[@agent]>`
  Example:
  `sbeads-3qt [B] [EPIC] (P1) Epic: Status Badges and Observability`
  `sbeads-55g [O] (P1/feat) Schema migration for agent persistence`
  `sbeads-9am [C] (P2/task) Implement tree renderer [@worker-1]`
- **Rationale**: Ultra-compact, token-efficient for LLM context, highly readable in terminal.

### Decision 4: Dependency Graph Embedding in `sbeads show`
- **Choice**: In `show.go`, query both incoming dependencies (blockers) and outgoing dependencies (dependents) and render them directly.
- **Rationale**: Eliminates the need to run `sbeads dep list` separately, giving an agent full situational awareness in a single command.

### Decision 5: DD-VSA Vertical Slices (< 400 LOC Ceiling)
- **Choice**: Keep `cmd/bd/list.go` (~120 LOC), `cmd/bd/unclaim.go` (~50 LOC), `internal/domain/status.go` (~90 LOC) well within the 400 LOC ceiling, accompanied by dedicated test files.
- **Rationale**: Prevents code bloat and guarantees high maintainability.

## Risks & Mitigations

| Risk | Mitigation |
| :--- | :--- |
| Schema migration failure on existing databases | Use idempotent `ALTER TABLE` execution ignoring duplicate column errors. |
| Circular dependency causing infinite loops in tree rendering | Track visited node IDs in depth-first traversal with a recursion limiter (max depth 10). |
| Performance degradation in `list` when querying dependencies | Batch fetch all dependencies or execute single join query in SQLite. |
