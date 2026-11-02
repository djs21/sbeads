# Internal Packages

## Purpose

Private package encapsulation boundary housing domain logic and shared types for sbeads.

## Ownership

- Core business logic and domain modeling in `internal/domain`
- Shared data structures, constants, and ID generation in `internal/types`

## Local Contracts

- Internal packages are strictly isolated to the sbeads module and cannot be imported by external Go projects
- Package `internal/types` holds pure data models without business logic or external dependencies
- Package `internal/domain` orchestrates business logic and graph operations using `backend.Storage`

## Work Guidance

- Prevent circular imports between internal packages
- Keep types focused on representation and domain packages focused on behavior

## Verification

- Run test suite: `go test ./internal/...`

## Child DOX Index

- [domain/AGENTS.md](domain/AGENTS.md): Domain rules, status computation, validation, and issue dependency graph
- [types/AGENTS.md](types/AGENTS.md): Core data models, constants, and Base36 ID generation
