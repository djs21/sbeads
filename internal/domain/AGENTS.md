# Domain Logic

## Purpose

Business logic, issue lifecycle transitions, dependency graph resolution, and domain error definitions.

## Ownership

- `ServiceDomain`, `IssueDomain`, and `DependencyDomain` services
- Status computation rules (`CLOSED`, `BLOCKED`, `CLAIMED`, `OPEN`)
- Issue claim and unclaim validation
- Dependency cycle detection and blocking rules
- Domain errors (`ErrIssueNotFound`, `ErrCircularDependency`, `ErrSelfDependency`, `ErrParentNotClosed`)

## Local Contracts

- Status calculation precedence:
  1. `CLOSED` when `IsClosed != 0`
  2. `BLOCKED` when `IsBlocked != 0` or any prerequisite parent issue is not closed
  3. `CLAIMED` when `IsClaimed != 0`
  4. `OPEN` when unblocked and unclaimed
- Direct circular dependencies and transitive cycles are rejected on creation
- Issues cannot depend on themselves (`ErrSelfDependency`)
- Issues must exist in storage before dependencies can be created

## Work Guidance

- Enforce all business rules within domain services before invoking backend storage
- Maintain clear domain error types for programmatic evaluation by callers

## Verification

- Run test suite: `go test ./internal/domain/...`

## Child DOX Index

(No child DOX files)
