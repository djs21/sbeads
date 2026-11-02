package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportCommand(t *testing.T) {
	setupTestDB(t)

	// Create 2 issues and a dependency
	out1, err := executeCmd("create", "Parent Export Issue", "-p", "2", "-d", "Desc Parent", "-y", "task")
	if err != nil {
		t.Fatalf("create 1 failed: %v", err)
	}
	id1 := extractCreatedID(out1)

	out2, err := executeCmd("create", "Child Export Issue", "-p", "1", "-d", "Desc Child", "-y", "bug")
	if err != nil {
		t.Fatalf("create 2 failed: %v", err)
	}
	id2 := extractCreatedID(out2)

	_, err = executeCmd("dep", "add", id2, id1)
	if err != nil {
		t.Fatalf("dep add failed: %v", err)
	}

	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "exported.jsonl")

	// 1. Run export command
	outExport, err := executeCmd("export", "--output", exportFile)
	if err != nil {
		t.Fatalf("export command failed: %v", err)
	}
	if !strings.Contains(outExport, "Exported") {
		t.Errorf("unexpected export output: %s", outExport)
	}

	// 2. Verify file content
	f, err := os.Open(exportFile)
	if err != nil {
		t.Fatalf("opening exported file failed: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Should contain lines for issues and dependencies
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 JSONL records (2 issues + 1 dep), got %d", len(lines))
	}

	foundID1, foundDep := false, false
	for _, l := range lines {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(l), &obj); err != nil {
			t.Fatalf("invalid json line %s: %v", l, err)
		}
		if obj["id"] == id1 {
			foundID1 = true
		}
		if obj["child_id"] == id2 && obj["parent_id"] == id1 {
			foundDep = true
		}
	}

	if !foundID1 {
		t.Errorf("expected issue %s in export JSONL", id1)
	}
	if !foundDep {
		t.Errorf("expected dependency %s -> %s in export JSONL", id2, id1)
	}
}
