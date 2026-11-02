package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestImportCommand(t *testing.T) {
	tempDir := t.TempDir()
	sourceDB := filepath.Join(tempDir, "source.db")
	targetDB := filepath.Join(tempDir, "target.db")
	jsonlFile := filepath.Join(tempDir, "issues.jsonl")

	// 1. Create source issues & dep in sourceDB
	outA, err := executeCmdWithDB(sourceDB, "create", "Task Alpha", "-p", "3", "-y", "feature")
	if err != nil {
		t.Fatalf("create source A failed: %v", err)
	}
	idA := extractCreatedID(outA)

	outB, err := executeCmdWithDB(sourceDB, "create", "Task Beta", "-p", "1", "-y", "bug")
	if err != nil {
		t.Fatalf("create source B failed: %v", err)
	}
	idB := extractCreatedID(outB)

	_, err = executeCmdWithDB(sourceDB, "dep", "add", idB, idA)
	if err != nil {
		t.Fatalf("dep add source failed: %v", err)
	}

	// 2. Export from sourceDB to JSONL
	_, err = executeCmdWithDB(sourceDB, "export", "-o", jsonlFile)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// 3. Import from JSONL into fresh targetDB
	outImport, err := executeCmdWithDB(targetDB, "import", "-i", jsonlFile)
	if err != nil {
		t.Fatalf("import command failed: %v", err)
	}
	if !strings.Contains(outImport, "Imported") {
		t.Errorf("unexpected import output: %s", outImport)
	}

	// 4. Verify targetDB contains imported issues with valid HMAC and relationships
	outShowA, err := executeCmdWithDB(targetDB, "show", idA)
	if err != nil {
		t.Fatalf("show imported A failed (possible tamper/HMAC issue): %v", err)
	}
	if !strings.Contains(outShowA, "Task Alpha") || !strings.Contains(outShowA, "Blocks:\n  "+idB) {
		t.Errorf("unexpected show output for imported A:\n%s", outShowA)
	}

	outShowB, err := executeCmdWithDB(targetDB, "show", idB)
	if err != nil {
		t.Fatalf("show imported B failed: %v", err)
	}
	if !strings.Contains(outShowB, "Task Beta") || !strings.Contains(outShowB, "Blocked by:\n  "+idA) {
		t.Errorf("unexpected show output for imported B:\n%s", outShowB)
	}
}
