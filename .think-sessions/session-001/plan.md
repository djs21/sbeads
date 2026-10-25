## sbeads - Domain Driven Vertical Slice Architecture

### Files to create in order:

1. **Backend (Data Layer)**
   - `backend/storage.go` - SQLite storage layer (done)
   
2. **Domain (Business Logic)**
   - `internal/domain/issue.go` - Issue domain object
   - `internal/domain/dependency.go` - Dependency domain object
   - `internal/domain/service.go` - Issue service (business logic)
   
3. **Application (CLI Commands)**
   - `cmd/bd/main.go` - CLI entry point
   - `cmd/bd/create.go` - Create command
   - `cmd/bd/ready.go` - Ready command
   - `cmd/bd/claim.go` - Claim command
   - `cmd/bd/close.go` - Close command
   - `cmd/bd/show.go` - Show command
   - `cmd/bd/dep/add.go` - Add dependency
   - `cmd/bd/dep/list.go` - List dependencies

### Architecture:
- **Backend**: Pure Go, SQLite, no external deps
- **Domain**: Business logic, domain objects, no CLI
- **Application**: CLI commands, integration with backend/domain
