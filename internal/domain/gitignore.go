package domain

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RequiredGitignoreEntries lists standard local patterns that must not be checked into git.
var RequiredGitignoreEntries = []string{
	".sbeads/*.db",
	".sbeads/*.db-wal",
	".sbeads/*.db-shm",
	".sbeads/auth.key",
}

// EnsureGitignoreProtection checks git root directory for .gitignore.
// If .gitignore exists, it appends any missing entries from RequiredGitignoreEntries.
// If .gitignore does not exist, it creates one with the required entries.
func EnsureGitignoreProtection(workspaceDir string) (bool, error) {
	if workspaceDir == "" {
		cwd, err := os.Getwd()
		if err == nil {
			workspaceDir = cwd
		}
	}

	// Locate git root
	gitRoot := findGitRoot(workspaceDir)
	if gitRoot == "" {
		// Not inside a git repository; no gitignore to manage
		return false, nil
	}

	gitignorePath := filepath.Join(gitRoot, ".gitignore")
	existingContent := ""
	existingLines := make(map[string]bool)

	if data, err := os.ReadFile(gitignorePath); err == nil {
		existingContent = string(data)
		scanner := bufio.NewScanner(strings.NewReader(existingContent))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				existingLines[line] = true
			}
		}
	}

	var missingEntries []string
	for _, entry := range RequiredGitignoreEntries {
		// Check both exact entry and generic .sbeads/ entry
		if !existingLines[entry] && !existingLines[".sbeads"] && !existingLines[".sbeads/"] {
			missingEntries = append(missingEntries, entry)
		}
	}

	if len(missingEntries) == 0 {
		return false, nil
	}

	var sb strings.Builder
	if existingContent != "" && !strings.HasSuffix(existingContent, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("\n# sbeads local database and secret keys\n")
	for _, entry := range missingEntries {
		sb.WriteString(entry)
		sb.WriteString("\n")
	}

	f, err := os.OpenFile(gitignorePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return false, fmt.Errorf("opening .gitignore for append: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(sb.String()); err != nil {
		return false, fmt.Errorf("writing to .gitignore: %w", err)
	}

	return true, nil
}

func findGitRoot(startDir string) string {
	curr := filepath.Clean(startDir)
	for {
		if stat, err := os.Stat(filepath.Join(curr, ".git")); err == nil {
			if stat.IsDir() || stat.Mode().IsRegular() { // Regular file in worktrees/submodules
				return curr
			}
		}
		parent := filepath.Dir(curr)
		if parent == curr || parent == "." || parent == "/" {
			break
		}
		curr = parent
	}
	return ""
}
