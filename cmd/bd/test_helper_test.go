package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func executeCmd(args ...string) (string, error) {
	var buf bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func setupTestDB(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	t.Setenv("SBEADS_DB", dbPath)
	t.Setenv("SBEADS_DIR", "")
	return dbPath
}
