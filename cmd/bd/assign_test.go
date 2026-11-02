package main

import (
	"strings"
	"testing"
)

func TestAssignCommand(t *testing.T) {
	setupTestDB(t)

	// Create an issue
	out, err := executeCmd("create", "Assignable Task", "-p", "1")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	id := extractCreatedID(out)

	// 1. Assign to a user
	outAssign, err := executeCmd("assign", id, "--assignee", "alice")
	if err != nil {
		t.Fatalf("assign failed: %v", err)
	}
	if !strings.Contains(outAssign, "Assigned issue: "+id) {
		t.Errorf("unexpected assign output: %s", outAssign)
	}

	// Verify show contains Assignee: alice
	outShow, err := executeCmd("show", id)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(outShow, "Assignee: alice") {
		t.Errorf("expected Assignee: alice in show, got:\n%s", outShow)
	}

	// 2. Assign with agent flag
	_, err = executeCmd("assign", id, "--agent", "agent-x")
	if err != nil {
		t.Fatalf("assign with agent failed: %v", err)
	}

	outShow2, err := executeCmd("show", id)
	if err != nil {
		t.Fatalf("second show failed: %v", err)
	}
	if !strings.Contains(outShow2, "Agent: @agent-x") {
		t.Errorf("expected Agent: @agent-x in show, got:\n%s", outShow2)
	}

	// 3. Assign without flags should fail
	_, err = executeCmd("assign", id)
	if err == nil {
		t.Fatal("expected assign without flags to fail")
	}

	// 4. Assign without ID should fail
	_, err = executeCmd("assign", "--assignee", "bob")
	if err == nil {
		t.Fatal("expected assign without ID to fail")
	}
}
