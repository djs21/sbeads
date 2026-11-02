# bd CLI

## Purpose

Command-line interface and subcommand implementations for the `sbeads` / `bd` tool.

## Ownership

- CLI commands: `claim`, `close`, `context`, `create`, `dep`, `list`, `ready`, `root`, `show`, `unclaim`, and `version`
- Flag parsing and validation
- Database ancestor discovery for `.sbeads/sbeads.db`
- Formatted console output, status badges, and CLI test helpers

## Local Contracts

- Commands initialize database and domain services via `AppContext` during `PersistentPreRunE`
- Database connections close in `PersistentPostRunE`
- Database location follows resolution precedence: explicit flag, `SBEADS_DB` / `SBEADS_DIR` environment variables, ancestor search for `.sbeads/`, falling back to local `.sbeads/sbeads.db`
- Status badges follow canonical representations: `[O]` for OPEN, `[C]` for CLAIMED, `[B]` for BLOCKED, `[X]` for CLOSED
- Non-zero exit code returned on application errors

## Work Guidance

- Use Cobra command definitions with `SilenceUsage: true` and `SilenceErrors: true`
- Delegate business validation to `internal/domain` rather than handling state changes in CLI commands
- Keep output formats predictable for machine parsing by AI agents

## Verification

- Run test suite: `go test ./cmd/bd/...`

## Child DOX Index

(No child DOX files)
