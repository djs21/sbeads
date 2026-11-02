package main

import (
	"strings"
	"testing"
)

func TestDeleteAndDepRemoveCommands(t *testing.T) {
	setupTestDB(t)

	// Create issues
	outA, err := executeCmd("create", "Parent Issue", "-p", "1")
	if err != nil {
		t.Fatalf("create Parent failed: %v", err)
	}
	idA := extractCreatedID(outA)

	outB, err := executeCmd("create", "Child Issue", "-p", "1")
	if err != nil {
		t.Fatalf("create Child failed: %v", err)
	}
	idB := extractCreatedID(outB)

	// Link B -> A
	_, err = executeCmd("dep", "add", idB, idA)
	if err != nil {
		t.Fatalf("dep add failed: %v", err)
	}

	// 1. Delete without force should fail because idB depends on idA
	_, err = executeCmd("delete", idA)
	if err == nil {
		t.Fatal("expected delete without force to fail when dependents exist")
	}

	// 2. Remove dependency using 'dep rm'
	outDepRm, err := executeCmd("dep", "rm", idB, idA)
	if err != nil {
		t.Fatalf("dep rm failed: %v", err)
	}
	if !strings.Contains(outDepRm, "Removed dependency:") {
		t.Errorf("unexpected dep rm output: %s", outDepRm)
	}

	// 3. Delete issue using alias 'rm'
	outRm, err := executeCmd("rm", idA)
	if err != nil {
		t.Fatalf("rm alias failed: %v", err)
	}
	if !strings.Contains(outRm, "Deleted issue:") {
		t.Errorf("unexpected rm output: %s", outRm)
	}

	// Verify idA is gone
	_, err = executeCmd("show", idA)
	if err == nil {
		t.Fatal("expected show on deleted issue to fail")
	}

	// 4. Create and delete with cascade using 'del -f'
	outC, err := executeCmd("create", "Parent C", "-p", "1")
	if err != nil {
		t.Fatalf("create Parent C failed: %v", err)
	}
	idC := extractCreatedID(outC)

	outD, err := executeCmd("create", "Child D", "-p", "1")
	if err != nil {
		t.Fatalf("create Child D failed: %v", err)
	}
	idD := extractCreatedID(outD)

	_, err = executeCmd("dep", "add", idD, idC)
	if err != nil {
		t.Fatalf("dep add D->C failed: %v", err)
	}

	outDel, err := executeCmd("del", idC, "-f")
	if err != nil {
		t.Fatalf("del -f failed: %v", err)
	}
	if !strings.Contains(outDel, "Deleted issue:") {
		t.Errorf("unexpected del output: %s", outDel)
	}

	// Verify dependency was removed from idD
	outShowD, err := executeCmd("show", idD)
	if err != nil {
		t.Fatalf("show D failed: %v", err)
	}
	if strings.Contains(outShowD, idC) {
		t.Errorf("expected idC to be removed from idD dependencies, got:\n%s", outShowD)
	}
}

func extractCreatedID(output string) string {
	// Format: "Created issue: sbeads-xxx (Title)"
	idx := strings.Index(output, "Created issue: ")
	if idx == -1 {
		return ""
	}
	part := output[idx+len("Created issue: "):]
	spaceIdx := strings.Index(part, " ")
	if spaceIdx != -1 {
		return part[:spaceIdx]
	}
	return strings.TrimSpace(part)
}
