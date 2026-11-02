package domain

import (
	"os"
	"path/filepath"
	"strings"
)

const SbeadsOnboardingSnippet = `
## Task Tracking with sbeads

This repository uses sbeads for lightweight, dependency-aware local issue tracking and multi-agent coordination.

- **Check ready tasks**: Run ` + "`sbeads ready`" + ` to view unblocked tasks.
- **Claim a task**: Run ` + "`sbeads claim <id> --agent <name>`" + ` before working.
- **Release a claim**: Run ` + "`sbeads unclaim <id>`" + ` if aborted or blocked.
- **Inspect details**: Run ` + "`sbeads show <id>`" + ` to view computed status and dependency edges.
- **Close completed work**: Run ` + "`sbeads close <id>`" + ` after tests pass.
- **Mutations & safety**: Use ` + "`sbeads update`" + `, ` + "`sbeads reopen`" + `, ` + "`sbeads delete`" + `, and ` + "`sbeads dep add/rm`" + `. Never modify database SQLite files directly.
`

// EnsureAgentsOnboarding checks if AGENTS.md exists in workspaceDir (or git root).
// If it does not exist, it creates a new AGENTS.md containing the sbeads onboarding instructions.
// If it exists but does not mention sbeads, it cleanly appends the snippet.
func EnsureAgentsOnboarding(workspaceDir string) (bool, error) {
	if workspaceDir == "" {
		cwd, err := os.Getwd()
		if err == nil {
			workspaceDir = cwd
		}
	}

	root := findGitRoot(workspaceDir)
	if root == "" {
		if filepath.Base(workspaceDir) == ".sbeads" {
			root = filepath.Dir(workspaceDir)
		} else {
			root = workspaceDir
		}
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	data, err := os.ReadFile(agentsPath)
	if err == nil {
		content := string(data)
		if strings.Contains(content, "sbeads") {
			// Already contains sbeads instructions
			return false, nil
		}

		var sb strings.Builder
		sb.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString(SbeadsOnboardingSnippet)

		if err := os.WriteFile(agentsPath, []byte(sb.String()), 0644); err != nil {
			return false, err
		}
		return true, nil
	}

	if !os.IsNotExist(err) {
		return false, err
	}

	// File does not exist: create brand new AGENTS.md
	header := "# Project Instructions\n"
	fullDoc := header + SbeadsOnboardingSnippet

	if err := os.WriteFile(agentsPath, []byte(fullDoc), 0644); err != nil {
		return false, err
	}
	return true, nil
}
