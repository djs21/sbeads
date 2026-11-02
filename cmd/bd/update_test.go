package main

import (
	"strings"
	"testing"
)

func TestUpdateCommand(t *testing.T) {
	setupTestDB(t)

	// Create an initial issue
	out, err := executeCmd("create", "Original Title", "-p", "1", "-d", "Original Desc", "-y", "task")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	id := extractCreatedID(out)

	// 1. Update title and priority
	outUp, err := executeCmd("update", id, "-t", "Updated Title", "-p", "3")
	if err != nil {
		t.Fatalf("update title and priority failed: %v", err)
	}
	if !strings.Contains(outUp, "Updated issue: "+id) {
		t.Errorf("unexpected update output: %s", outUp)
	}

	// Verify with show
	outShow, err := executeCmd("show", id)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(outShow, "Title: Updated Title") || !strings.Contains(outShow, "Priority: 3") {
		t.Errorf("fields not updated in show output:\n%s", outShow)
	}

	// 2. Update type to epic and description
	_, err = executeCmd("update", id, "-y", "epic", "-d", "New Epic Description")
	if err != nil {
		t.Fatalf("update type and desc failed: %v", err)
	}

	outShow2, err := executeCmd("show", id)
	if err != nil {
		t.Fatalf("second show failed: %v", err)
	}
	if !strings.Contains(outShow2, "Type: epic") || !strings.Contains(outShow2, "New Epic Description") {
		t.Errorf("type or desc not updated in show output:\n%s", outShow2)
	}

	// 3. Update without any flag should fail
	_, err = executeCmd("update", id)
	if err == nil {
		t.Fatal("expected update without flags to return error")
	}

	// 4. Update with empty title should fail
	_, err = executeCmd("update", id, "-t", "   ")
	if err == nil {
		t.Fatal("expected update with empty title to return error")
	}
}
