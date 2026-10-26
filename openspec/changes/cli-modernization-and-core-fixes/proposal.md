## Why

The current `sbeads` CLI implementation relies on a monolithic `cmd/bd/main.go` built on Go's standard `flag` package, which lacks standardized POSIX flag parsing (`-h`, `--help`), structured per-subcommand help (`sbeads help <cmd>`), and clean architectural separation. Concurrently, the core issue tracking engine suffers from two severe functional flaws:
1. The `ready` query currently returns all issues unconditionally, without filtering out closed, claimed, or manually blocked tasks, and fails to check whether parent dependencies remain open.
2. The ID generation produces unwieldy 32-character hexadecimal strings that are cumbersome for human developers and AI agents to reference in CLI workflows.

Modernizing the CLI architecture to Cobra with Domain-Driven Vertical Slice Architecture (DD-VSA), implementing a robust SQL-based Ready query with recursive dependency awareness, and adopting ergonomic 3-character hash IDs directly resolves these usability and correctness bottlenecks.

## What Changes

- **CLI Framework Modernization**: Migrate `cmd/bd` from a single monolithic file to `spf13/cobra` vertical slices (`root.go`, `create.go`, `ready.go`, `claim.go`, `close.go`, `show.go`, `dep.go`, `version.go`, `context.go`), enforcing a strict 400 LOC ceiling per file.
- **Application Context Lifecycle**: Introduce `AppContext` initialized in `PersistentPreRunE` on the root command to manage database resolution, backend storage, domain services, and output writer dependency injection.
- **Unified Help & POSIX Flags**: Support standard POSIX short/long flags (`-h`, `--help`, `-t`, `-p`, `-a`, `-m`) and per-subcommand help invocations via both `sbeads <cmd> --help` and `sbeads help <cmd>`.
- **Accurate Ready Query with Dependency Resolution**: Implement a single SQL query in storage that filters issues where `is_closed = 0 AND is_claimed = 0 AND is_blocked = 0 AND NOT EXISTS (SELECT 1 FROM dependencies d JOIN issues p ON d.parent_id = p.id WHERE d.child_id = issues.id AND p.is_closed = 0)`.
- **Ergonomic 3-Character Hash IDs**: Replace 32-char hex IDs with 3-character alphanumeric/hex hash identifiers with prefix support (e.g., `sbeads-xxx` or `issue-xxx`) including automatic collision detection and retry mechanisms.

## Capabilities

### New Capabilities
- `cli-framework`: Modular Cobra-based command hierarchy with DD-VSA vertical slices, POSIX flag support, and unified per-command help documentation.
- `ready-query`: Storage-level query filtering closed, claimed, blocked issues, and issues blocked by unclosed parent dependencies.
- `id-format`: Compact 3-character hash ID generation with project prefixing and collision resolution.

### Modified Capabilities

## Impact

- **CLI Interface**: Subcommands now accept POSIX short/long flags (`-h`, `--help`, `--title`, etc.) and support `sbeads help <subcommand>`.
- **Storage Layer**: `backend/storage.go` adds dedicated `GetReadyIssues()` query with dependency join filtering.
- **Domain Layer**: `internal/domain/issue.go` delegates ready listing to the storage engine's dependency-aware query.
- **Types & Identifiers**: `internal/types/types.go` updates ID generation routines to output 3-character hashes (`<prefix>-<hash3>`).
- **Dependencies**: Adds `github.com/spf13/cobra` to `go.mod`.
- **Source Structure**: Decomposes monolithic `cmd/bd/main.go` into modular command slice files in `cmd/bd/`.
