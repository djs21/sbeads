## Context

The `sbeads` issue tracker was initially implemented with a monolithic CLI in `cmd/bd/main.go` using the standard library `flag` package. As functionality expanded to issue creation, claiming, closing, viewing, and dependency tracking, several architectural and behavioral deficits emerged:

1. **CLI Architectural Coupling**: All commands, database path discovery, argument parsing, error handling, and formatting live within a single ~400-line file without dependency injection, making automated testing difficult and violating separation of concerns.
2. **Flag & Help Limitations**: Go's stdlib `flag` package parses `-flag` rather than standard POSIX `--flag` / `-f` syntax, and does not provide built-in subcommand dispatch or standardized help (`sbeads help <cmd>` and `<cmd> --help`).
3. **Flawed Ready Query**: The domain method `ReadyIssues()` blindly calls `storage.ListIssues()` without SQL-level filtering for closed, claimed, or manually blocked issues, and completely omits parent dependency validation. Consequently, tasks whose prerequisite dependencies are still open appear as ready to be worked on.
4. **Identifier Ergonomics**: Issue IDs are generated as raw 32-character hexadecimal strings (128-bit UUID/random hex), which creates friction for human typing and prompt token bloat for autonomous AI agents.

## Goals / Non-Goals

**Goals:**
- Decompose the CLI into Domain-Driven Vertical Slice Architecture (DD-VSA) under `cmd/bd/` using `spf13/cobra`, with each command slice constrained below a 400 LOC ceiling.
- Introduce an `AppContext` struct injected during `PersistentPreRunE` to supply storage, domain services, database path, and I/O streams across all subcommands.
- Standardize all flags to POSIX syntax (supporting short flags like `-h`, `-t`, `-p`, `-a`, `-m` and long flags like `--help`, `--title`, `--priority`) with first-class `sbeads help <cmd>` navigation.
- Implement an optimized SQL Ready query in `backend/storage.go` that excludes closed, claimed, or blocked issues, and verifies all parent dependencies are closed.
- Implement a 3-character hash ID generator (`[0-9a-f]{3}` or alphanumeric) with configurable/default project prefix (e.g., `sbeads-a1b` or `issue-a1b`) and collision detection.
- Provide comprehensive unit, integration, and E2E test suites with in-memory buffers to guarantee zero regressions.

**Non-Goals:**
- Introducing an external RPC or client-server daemon (sbeads remains an embedded SQLite CLI).
- Implementing multi-tenant access control or cloud database backends.
- Replacing the underlying SQLite engine with an alternative database.
- Altering the historical data in existing databases beyond forward compatibility with short IDs.

## Decisions

### Decision 1: spf13/cobra DD-VSA with 400 LOC Ceiling
- **Choice**: Adopt `github.com/spf13/cobra` organizing subcommands into vertical slices (`root.go`, `create.go`, `ready.go`, `claim.go`, `close.go`, `show.go`, `dep.go`, `version.go`, `context.go`).
- **Rationale**: Cobra provides battle-tested subcommand routing, POSIX flag parsing via `pflag`, automatic help page generation, and execution hooks. Restricting each file to < 400 LOC enforces focused single-responsibility modules.
- **Alternatives Considered**:
  - *Keep `flag`*: Inflexible, requires custom subcommand dispatching, manual help generation, and poor POSIX compliance.
  - *`urfave/cli/v2`*: Functional, but Cobra is the de facto standard in Go infrastructure and provides cleaner test isolation via `cmd.SetOut()` / `cmd.SetErr()`.

### Decision 2: Persistent PreRun Lifecycle & AppContext
- **Choice**: Define an `AppContext` holding `Storage`, `ServiceDomain`, `DBPath`, and I/O streams (`Out`, `Err`). Initialize `AppContext` inside `rootCmd.PersistentPreRunE` and attach it to command contexts or closure state.
- **Rationale**: Isolates infrastructure setup (database discovery, migration check, connection opening) from individual command handlers. Enables command unit testing by allowing test harnesses to inject mock or in-memory contexts without hitting real disk files.
- **Alternatives Considered**:
  - *Global singleton variables*: Pollutes package namespace, causes race conditions during parallel tests, and inhibits clean teardown.

### Decision 3: Storage-Level Ready Query with Dependency SQL Filter
- **Choice**: Implement `GetReadyIssues()` directly in `backend/storage.go` executing:
  ```sql
  SELECT id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, created_at, updated_at
  FROM issues
  WHERE is_closed = 0
    AND is_claimed = 0
    AND is_blocked = 0
    AND NOT EXISTS (
      SELECT 1 FROM dependencies d
      JOIN issues p ON d.parent_id = p.id
      WHERE d.child_id = issues.id
        AND p.is_closed = 0
    )
  ORDER BY priority ASC, created_at ASC;
  ```
- **Rationale**: SQLite handles subquery filtering in microseconds. Executing filtering in SQL prevents fetching thousands of irrelevant rows into memory and avoids N+1 query overhead in the domain layer.
- **Alternatives Considered**:
  - *Domain-level filtering (in-memory)*: Requires querying all dependencies and parents in Go loops, resulting in N+1 database queries and race conditions under concurrent updates.

### Decision 4: 3-Character Hash ID Generation with Collision Handling
- **Choice**: Generate IDs using `<prefix>-<hash3>` where `<hash3>` is a 3-character hex/alnum string derived from cryptographic random bytes (providing 4,096 hex or 46,656 base36 combinations). The domain/storage layer checks for existing IDs and retries on collision up to 5 times.
- **Rationale**: For agile issue tracking (typically dozens to hundreds of active issues per repo), 3 characters provide maximum typing efficiency (`bd show a1b` or `bd claim a1b`). The collision retry ensures mathematical safety.
- **Alternatives Considered**:
  - *UUID v4 / 32-char hex*: Excessively verbose, awkward to copy/paste in terminals.
  - *Sequential integers (`#1`, `#2`)*: Prone to merge conflicts across distributed git branches. Short random hashes remain decentralized and collision-resistant across branches.

### Decision 5: Unified Help & POSIX Flag Conventions
- **Choice**: Leverage Cobra's command tree to automatically bind `-h` / `--help` on all commands, support `sbeads help <command>` interchangeably, and configure standard POSIX flags (e.g., `-t` / `--title`, `-p` / `--priority`, `-d` / `--description`, `-a` / `--assignee`, `-m` / `--agent`).
- **Rationale**: Aligns `sbeads` with UNIX conventions and makes CLI behavior intuitive for both humans and agentic tool runners.
- **Alternatives Considered**:
  - *Custom switch statement parsing*: Fragile, error-prone, and reinvents standard CLI behaviors.

## Risks / Trade-offs

- **[Risk] Namespace collision with 3-char IDs in large repositories**: With 16^3 = 4,096 possibilities (hex) or 36^3 = 46,656 (base36), projects exceeding several thousand issues have increased collision probability.
  → **Mitigation**: Implement automatic retry on insert collision; if collision rate exceeds threshold, support optional length expansion (e.g., 4 characters) via configuration.
- **[Risk] Existing 32-char hex IDs in pre-existing databases**: Existing databases may contain legacy 32-char IDs.
  → **Mitigation**: All database queries and CLI commands (`show`, `claim`, `close`, `dep`) query by exact ID string or prefix matching, ensuring complete backward compatibility with existing databases.
- **[Risk] SQLite lock contention on concurrent CLI invocations**: SQLite file locking under high concurrency.
  → **Mitigation**: Retain WAL mode (`PRAGMA journal_mode=WAL;`) and busy timeout in SQLite storage initialization.
