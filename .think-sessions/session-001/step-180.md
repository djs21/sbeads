## Input: Go version is 1.27.3 (newer than beads' 1.26.5)

## CONCLUSION: Use Go 1.27 with modern features

**Go version:** 1.27.3 (latest)

**Implementation plan:**
1. Initialize go.mod with Go 1.27
2. Create skeleton: main.go, SQLite backend, core commands
3. Implement: create, ready, claim, close, show, dep
4. Use SQLite for storage (not Dolt)
5. Enable CGO_ENABLED=0 for pure Go build

**Skeleton structure:**
```
sbeads/
├── go.mod
├── main.go
├── cmd/
│   └── bd/
│       ├── main.go
│       ├── create.go
│       ├── ready.go
│       ├── claim.go
│       ├── close.go
│       ├── show.go
│       └── dep/
│           ├── add.go
│           ├── remove.go
│           └── list.go
└── backend/
    └── sqlite.go
```

## Next: Create skeleton and implement SQLite backend
