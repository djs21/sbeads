## Input: User asked "will this take lot of time?"

## CONCLUSION: Yes, it will take time - but sbeads is minimal

**Why it takes time:**
- Need full Dolt backend (SQL database with version control)
- Need atomic operations (`--claim` sets assignee + in_progress)
- Need dependency graph tracking
- Need audit trails for all commands
- Need hash-based IDs (bd-a1b2) to prevent collisions

**What makes it faster:**
- No sync/pull (saves complex Dolt remote handling)
- No agent hooks (no installation code)
- No memory system (no remember/prime)
- No HTTP backend (no server code)
- No backup/compact (saves space management)

**Estimated timeline:**
- Skeleton structure: 30-60 minutes
- Core commands (create/update/close/ready/show): 2-4 hours
- Dependencies: 1-2 hours
- Testing + polish: 1-2 hours
- **Total: ~5-8 hours** (vs 20+ for full beads)

## Next: Implement sbeads with skeleton first, then fill in
