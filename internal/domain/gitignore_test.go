package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGitignoreProtection(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("mkdir .git failed: %v", err)
	}

	// 1. Initially no .gitignore: should create one and add entries
	appended, err := EnsureGitignoreProtection(tempDir)
	if err != nil {
		t.Fatalf("EnsureGitignoreProtection failed: %v", err)
	}
	if !appended {
		t.Errorf("expected appended to be true")
	}

	gitignorePath := filepath.Join(tempDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("reading .gitignore failed: %v", err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, ".sbeads/auth.key") || !strings.Contains(strContent, ".sbeads/*.db") {
		t.Errorf("missing expected entries in .gitignore: %s", strContent)
	}

	// 2. Second call: all entries exist, should be idempotent (no duplicate appends)
	appendedSecond, err := EnsureGitignoreProtection(tempDir)
	if err != nil {
		t.Fatalf("second EnsureGitignoreProtection failed: %v", err)
	}
	if appendedSecond {
		t.Errorf("expected appended to be false when all entries already exist")
	}

	// 3. If repo already ignores .sbeads/ completely
	tempDir2 := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDir2, ".git"), 0755); err != nil {
		t.Fatalf("mkdir .git failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir2, ".gitignore"), []byte(".sbeads/\n"), 0644); err != nil {
		t.Fatalf("writing .gitignore failed: %v", err)
	}

	appended3, err := EnsureGitignoreProtection(tempDir2)
	if err != nil {
		t.Fatalf("EnsureGitignoreProtection with blanket ignore failed: %v", err)
	}
	if appended3 {
		t.Errorf("expected false when .sbeads/ is already globally ignored")
	}
}
