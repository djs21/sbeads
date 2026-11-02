# bd CLI

## Purpose

Command-line interface and subcommand implementations for the `sbeads` / `bd` tool.

## Ownership

- CLI commands: `claim`, `close`, `context`, `create`, `delete` (aliases `del`, `rm`), `dep` (subcommands `add`, `list`, `rm`), `list`, `ready`, `root`, `show`, `unclaim`, and `version`
- Flag parsing and validation
- Database ancestor discovery for `.sbeads/sbeads.db`
- Formatted console output, status badges, and CLI test helpers

## Local Contracts

- Structure follows Domain-Driven Vertical Slice Architecture (DD-VSA) where each command slice is an isolated file under cmd/bd/
- Strict source file ceiling: every source file and test file must stay under 400 LOC
- Commands initialize database and domain services via `AppContext` during `PersistentPreRunE`
- Database connections close in `PersistentPostRunE`
- Database location follows resolution precedence: explicit flag, `SBEADS_DB` / `SBEADS_DIR` environment variables, ancestor search for `.sbeads/`, falling back to local `.sbeads/sbeads.db`
- Status badges follow canonical representations: `[O]` for OPEN, `[C]` for CLAIMED, `[B]` for BLOCKED, `[X]` for CLOSED
- Non-zero exit code returned on application errors
- Delete command refuses to remove an issue if active dependent issues exist unless --force or --cascade is passed
- Create command determines issue prefix using flag --prefix, environment variable SBEADS_PREFIX, or active git repository folder name

## Work Guidance

- Use Cobra command definitions with `SilenceUsage: true` and `SilenceErrors: true`
- Delegate business validation to `internal/domain` rather than handling state changes in CLI commands
- Keep output formats predictable for machine parsing by AI agents

## Verification

- Run test suite: `go test ./cmd/bd/...`

## Child DOX Index

(No child DOX files)
