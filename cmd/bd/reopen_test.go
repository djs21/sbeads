package main

import (
	"strings"
	"testing"
)

func TestReopenCommand(t *testing.T) {
	setupTestDB(t)

	// Create and close an issue
	out, err := executeCmd("create", "Reopenable Task", "-p", "1")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	id := extractCreatedID(out)

	_, err = executeCmd("close", id)
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// Verify it is closed
	outShow, err := executeCmd("show", id)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(outShow, "Status: CLOSED [X]") {
		t.Fatalf("expected CLOSED status, got:\n%s", outShow)
	}

	// 1. Reopen the issue
	outReopen, err := executeCmd("reopen", id)
	if err != nil {
		t.Fatalf("reopen command failed: %v", err)
	}
	if !strings.Contains(outReopen, "Reopened issue: "+id) {
		t.Errorf("unexpected reopen output: %s", outReopen)
	}

	// 2. Verify with show: status should be OPEN [O]
	outShowAfter, err := executeCmd("show", id)
	if err != nil {
		t.Fatalf("second show failed: %v", err)
	}
	if !strings.Contains(outShowAfter, "Status: OPEN [O]") {
		t.Errorf("expected OPEN [O] after reopen, got:\n%s", outShowAfter)
	}

	// 3. Reopening an already open issue should return an error
	_, err = executeCmd("reopen", id)
	if err == nil {
		t.Fatal("expected reopening already open issue to fail")
	}

	// 4. Reopening without ID should return an error
	_, err = executeCmd("reopen")
	if err == nil {
		t.Fatal("expected reopen without ID to fail")
	}
}
