package domain

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9-]+`)

// CleanPrefix cleans a candidate string into a valid prefix:
// converts to lower case, replaces invalid characters with dash,
// trims leading and trailing dashes. Falls back to "sbeads" if empty.
func CleanPrefix(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanumericRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "sbeads"
	}
	return s
}

// DetectProjectPrefix walks up from startDir to discover git root or returns directory base
func DetectProjectPrefix(startDir string) string {
	if startDir == "" {
		cwd, err := os.Getwd()
		if err == nil {
			startDir = cwd
		}
	}
	if startDir == "" {
		return "sbeads"
	}

	curr := filepath.Clean(startDir)
	for {
		gitPath := filepath.Join(curr, ".git")
		sbeadsPath := filepath.Join(curr, ".sbeads")
		if _, err := os.Stat(gitPath); err == nil {
			base := filepath.Base(curr)
			return CleanPrefix(base)
		}
		if stat, err := os.Stat(sbeadsPath); err == nil && stat.IsDir() {
			base := filepath.Base(curr)
			return CleanPrefix(base)
		}

		parent := filepath.Dir(curr)
		if parent == curr || parent == "." || parent == "/" {
			break
		}
		curr = parent
	}

	base := filepath.Base(startDir)
	return CleanPrefix(base)
}
