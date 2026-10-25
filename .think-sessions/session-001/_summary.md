## Completed Work Summary

1. Fixed flag.Parse() order - moved before all flag definitions
2. Added sqlite3 driver import
3. Changed Issue struct bool fields to int
4. Updated domain layer to use int instead of bool
5. Fixed HandleCreate to pass int value

## In Progress
- Testing create command with sqlite3

## Key Files Modified
- backend/storage.go: sqlite driver, removed ? placeholders, fixed boolean handling
- internal/types/types.go: bool -> int
- internal/domain/issue.go: bool -> int
- internal/domain/service.go: bool -> int
- cmd/bd/main.go: blocked param changed to int
