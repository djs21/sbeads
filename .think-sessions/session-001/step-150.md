## Input: User asked why Dolt vs SQLite - pros/cons

## CONCLUSION: Dolt adds version control and atomicity over SQLite

**Dolt advantages:**
- Git-like branching (create `bd-a1b2-branch`, merge later)
- Cell-level merge (automatically merges conflicting changes)
- Cross-machine sync (pull from other agents' databases)
- Built-in history/audit trail per row
- Automatic conflict resolution for branching issues

**SQLite advantages:**
- Simpler, single-file database
- No external dependencies
- Faster startup (no git-like overhead)
- Easier to backup (single file)
- No conflicts possible (no branching)

**Why beads uses Dolt:**
- Agents need to **branch issues** without manual merge
- Multiple agents can work on same issue **independently**
- **Cross-machine sync** (pull issues from other agents)
- **Atomic operations** with conflict resolution
- Full audit trail of every change

**Trade-offs:**
- Dolt is **slower** (more features = more overhead)
- SQLite is **faster** (no git-like overhead)
- Dolt requires Dolt binary (external dependency)
- SQLite is standalone binary

## Next: Implement sbeads with Dolt backend (user's choice)
