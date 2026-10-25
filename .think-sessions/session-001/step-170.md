## Input: User confirmed Go implementation, wants latest Go version

## CONCLUSION: Use Go 1.23+ for modern features

**Latest Go:** 1.23.x (current stable)

**Go version strategy:**
- Use Go 1.23 (latest stable)
- Enable CGO disabled for pure Go build: `CGO_ENABLED=0 go build`
- Use modern Go features (modular, cleaner code)
- Follow beads codebase conventions

**Project Go version:**
- beads uses Go 1.21+ (modern)
- Uses CGO_ENABLED=0 for pure Go builds
- Uses modules (go.mod)

**Implementation plan:**
1. Initialize Go module with go 1.23
2. Create skeleton command structure
3. Implement SQLite backend (not Dolt)
4. Implement core commands: create, ready, claim, close, show, dep
5. Test with sample data

## Next: Implement sbeads with Go 1.23, SQLite backend
