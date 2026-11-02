package main

import (
	"strings"
	"testing"
)

func TestBlockAndUnblockCommands(t *testing.T) {
	setupTestDB(t)

	// Create an unblocked task
	out, err := executeCmd("create", "Blockable Task", "-p", "1")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	id := extractCreatedID(out)

	// 1. Initially it should be in ready queue
	outReady, err := executeCmd("ready")
	if err != nil {
		t.Fatalf("ready failed: %v", err)
	}
	if !strings.Contains(outReady, id) {
		t.Fatalf("expected issue %s to be in ready queue", id)
	}

	// 2. Block the issue explicitly
	outBlock, err := executeCmd("block", id)
	if err != nil {
		t.Fatalf("block command failed: %v", err)
	}
	if !strings.Contains(outBlock, "Blocked issue: "+id) {
		t.Errorf("unexpected block output: %s", outBlock)
	}

	// Verify show displays BLOCKED [B]
	outShow, err := executeCmd("show", id)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(outShow, "Status: BLOCKED [B]") {
		t.Errorf("expected BLOCKED status, got:\n%s", outShow)
	}

	// Verify it is no longer in ready queue
	outReady2, err := executeCmd("ready")
	if err != nil {
		t.Fatalf("second ready failed: %v", err)
	}
	if strings.Contains(outReady2, id) {
		t.Errorf("blocked issue %s should not appear in ready queue", id)
	}

	// 3. Unblock the issue
	outUnblock, err := executeCmd("unblock", id)
	if err != nil {
		t.Fatalf("unblock command failed: %v", err)
	}
	if !strings.Contains(outUnblock, "Unblocked issue: "+id) {
		t.Errorf("unexpected unblock output: %s", outUnblock)
	}

	// Verify show displays OPEN [O]
	outShow2, err := executeCmd("show", id)
	if err != nil {
		t.Fatalf("show after unblock failed: %v", err)
	}
	if !strings.Contains(outShow2, "Status: OPEN [O]") {
		t.Errorf("expected OPEN status after unblock, got:\n%s", outShow2)
	}

	// Verify it is back in ready queue
	outReady3, err := executeCmd("ready")
	if err != nil {
		t.Fatalf("third ready failed: %v", err)
	}
	if !strings.Contains(outReady3, id) {
		t.Errorf("unblocked issue %s should be in ready queue", id)
	}
}
