## Input: User uses beads alone, agents update sequentially, wants SQLite

## CONCLUSION: SQLite is perfect for your use case

**Why SQLite is better for YOU:**
- You work alone (no team sync needed)
- Agents update sequentially (no conflicts)
- Simpler = faster startup
- No Dolt binary dependency
- Single file = easier backup
- No branch/merge logic needed
- Faster for your workflow

**SQLite advantages for solo use:**
- No Dolt overhead
- Simpler codebase
- Faster
- Easier to backup (single file)
- No version control needed
- No conflict resolution logic

**What we keep from beads:**
- Issue lifecycle (create → ready → claim → close)
- Dependency management
- Hash-based IDs (still useful)
- Audit trails
- Atomic operations (still needed for consistency)

**What we drop:**
- Dolt backend → SQLite
- Branching/merging
- Cross-machine sync
- Team features

## Next: Implement sbeads with SQLite backend
