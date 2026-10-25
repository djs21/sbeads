## Input: Analyzed beads README.md for feature mapping

## CONCLUSION: Top features for sbeads cherry-picking

**Rank 1: Issue Lifecycle (create → ready → claim → close)**
- Core workflow: `bd create` → `bd ready` → `bd update --claim` → `bd close`
- Hash-based IDs prevent merge collisions
- Priority P0 "critical" = highest priority work
- Dependencies prevent issues from being ready until resolved

**Rank 2: Persistent Memory**
- `bd prime` - workflow context injection
- `bd remember "insight"` - durable project facts

**Rank 3: Sync & Collaboration**
- `bd dolt push/pull` - cross-machine sync
- `bd connect` - remote `bd serve` over HTTP
- `bd ready --claim` - one-step workflow with claim

**Rank 4: Agent Integrations**
- `bd setup codex/claude/factory` - install agent tools/hooks

## Next: Create sbeads with minimal viable feature set based on user choice
