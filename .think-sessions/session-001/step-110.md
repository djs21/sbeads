## Input: User asked "what is bd prime used for?"

## CONCLUSION: bd prime = workflow context injection

**What `bd prime` does:**
1. Reads project memories from `bd remember "insight"`
2. Injects them into agent's working context
3. Prints workflow guidance based on project state
4. Enables agent to understand: open issues, dependencies, project history

**Example:**
```bash
# Store fact
bd remember "Project uses TypeScript"

# Later, agent sees this fact automatically
bd prime
```

**Core beads features:**

**1. Issue Lifecycle** (create → ready → claim → close)
- `bd create "Fix bug" -p 0 -t bug`
- `bd ready`
- `bd update bd-a1b2 --claim`
- `bd close bd-a1b2 "Fixed"`

**2. Persistent Memory** (what agent remembers about project)
- `bd remember "Project uses TypeScript"`
- `bd prime` (injects memories into workflow)

**3. Sync & Collaboration**
- `bd dolt push/pull` - cross-machine sync
- `bd connect` - remote `bd serve` over HTTP
- `bd ready --claim` - one-step workflow

**4. Agent Integrations**
- `bd setup codex/claude/factory` - install agent tools

## Next: User needs to choose which feature to implement for sbeads
