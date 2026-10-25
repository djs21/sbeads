## Input: User wants sbeads to serve ONLY issue lifecycle + dependency management

## CONCLUSION: Yes, this is achievable - core issue tracker

**Target features to implement:**

**Issue Lifecycle:**
- `bd create "Title" -p PRIORITY -t TYPE`
- `bd update <id> --claim` (atomic: assignee + in_progress)
- `bd close <id>` (mark done)
- `bd ready` (list unblocked issues)
- `bd show <id>` (details + audit trail)

**Dependency Management:**
- `bd dep add <child> <parent>` (link tasks)
- `bd dep tree` (show dependency graph)
- `bd dep remove` (unlink)
- `bd dep list` (list dependencies for an issue)

**What to exclude (not needed):**
- Sync (`dolt push/pull`, `connect`)
- Agent integrations (`hooks`, `setup`)
- Memory (`remember`, `prime`)
- Backup/compact/upgrade
- HTTP backend
- WISPs, audits, metrics

## Next: Create sbeads with issue tracker + deps only
