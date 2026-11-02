# Backend

## Purpose

SQLite persistence engine and database operations for sbeads.

## Ownership

- Storage interface definitions (`Storage`) and SQLite implementation (`StorageImpl`)
- Database migrations, table creation, and indices
- SQLite connection lifecycle, WAL configuration, and busy timeouts
- Issue CRUD, dependency persistence, transactional deletion (with cascade), and query execution

## Local Contracts

- SQLite driver must use pure-Go implementation `modernc.org/sqlite` with `CGO_ENABLED=0`
- Issue mutation records are signed with HMAC-SHA256 using `.sbeads/auth.key` (strictly enforcing 0600 file permissions)
- Connections configure WAL mode and busy timeout via connection string parameters: `_journal_mode=WAL&_busy_timeout=5000`
- Connection pool limits open connections to 1 (`db.SetMaxOpenConns(1)`) to avoid file locking conflicts
- Storage creates missing parent directories automatically with permissions `0755`
- Default database file resolves to `.sbeads/sbeads.db`

## Work Guidance

- Ensure all SQL statements use query parameters rather than string concatenation
- Handle closed database handles cleanly during application shutdown
- Maintain retry logic for ID collisions when inserting new records

## Verification

- Run test suite: `go test ./backend/...`

## Child DOX Index

(No child DOX files)
