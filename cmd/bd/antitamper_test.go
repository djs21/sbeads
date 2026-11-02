package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndToEnd_AntiTamperAndOnboarding(t *testing.T) {
	tempDir := t.TempDir()
	dbDir := filepath.Join(tempDir, ".sbeads")
	dbPath := filepath.Join(dbDir, "sbeads.db")

	// 1. Create tables and check auto-onboarding
	outCreate, err := executeCmdWithDB(dbPath, "create", "Tamper Test Issue", "-p", "2")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	id := extractCreatedID(outCreate)

	// Check AGENTS.md auto creation
	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("expected AGENTS.md to be created: %v", err)
	}
	if !strings.Contains(string(content), "sbeads ready") {
		t.Errorf("expected onboarding guide in AGENTS.md, got:\n%s", string(content))
	}

	// 2. Tamper directly in SQLite
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	_, err = db.Exec("UPDATE issues SET title = 'Corrupted Title by Rogue Process' WHERE id = ?", id)
	_ = db.Close()
	if err != nil {
		t.Fatalf("raw SQL update failed: %v", err)
	}

	// 3. Command show must detect tamper and fail
	_, err = executeCmdWithDB(dbPath, "show", id)
	if err == nil || !strings.Contains(err.Error(), "tamper detected") {
		t.Fatalf("expected tamper detected error on show, got: %v", err)
	}

	// 4. Command ready must also detect tamper and fail
	_, err = executeCmdWithDB(dbPath, "ready")
	if err == nil || !strings.Contains(err.Error(), "tamper detected") {
		t.Fatalf("expected tamper detected error on ready, got: %v", err)
	}
}

func executeCmdWithDB(dbPath string, args ...string) (string, error) {
	fullArgs := append([]string{"--db", dbPath}, args...)
	return executeCmd(fullArgs...)
}
