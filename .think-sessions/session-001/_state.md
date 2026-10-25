## Task: Create sbeads - issue tracker with SQLite backend (Go)
## Progress: Step 3 of 5 — Fix database driver
## Completed: Changed all bool to int; fixed HandleCreate
## Status: in-progress
## Last Action: Changed blocked param to int and added conversion in HandleCreate
## Next Action: Build and test create command
## Key Files:
- cmd/bd/main.go: blocked param changed to int
- internal/types/types.go: bool->int
- internal/domain/issue.go: bool->int
- internal/domain/service.go: bool->int
## Decisions: Using Go flag package, sqlite3 driver
