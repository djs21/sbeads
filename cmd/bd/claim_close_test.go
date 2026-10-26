package main

import (
	"strings"
	"testing"
)

func TestClaimAndCloseCommand(t *testing.T) {
	setupTestDB(t)

	// Create an issue for testing
	createOut, err := executeCmd("create", "-t", "Work Item", "-p", "1")
	if err != nil {
		t.Fatalf("unexpected error creating issue: %v", err)
	}
	parts := strings.Split(createOut, "Created issue: ")
	issueID := strings.TrimSpace(strings.Split(parts[1], " ")[0])

	t.Run("Claim without issue-id fails", func(t *testing.T) {
		_, err := executeCmd("claim")
		if err == nil || !strings.Contains(err.Error(), "--issue-id is required") {
			t.Fatalf("expected missing issue-id error, got: %v", err)
		}
	})

	t.Run("Claim with flags", func(t *testing.T) {
		out, err := executeCmd("claim", "--issue-id", issueID, "--assignee", "alice", "--agent", "testbot")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "Claimed issue: " + issueID + " (Work Item) by testbot"
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q, got %q", expected, out)
		}
	})

	t.Run("Claim already claimed issue fails", func(t *testing.T) {
		_, err := executeCmd("claim", "--issue-id", issueID, "--assignee", "bob")
		if err == nil {
			t.Fatal("expected error claiming already claimed issue, got nil")
		}
	})

	t.Run("Close without issue-id fails", func(t *testing.T) {
		_, err := executeCmd("close")
		if err == nil || !strings.Contains(err.Error(), "--issue-id is required") {
			t.Fatalf("expected missing issue-id error, got: %v", err)
		}
	})

	t.Run("Close with flag", func(t *testing.T) {
		out, err := executeCmd("close", "--issue-id", issueID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "Closed issue: " + issueID + " (Work Item)"
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q, got %q", expected, out)
		}
	})

	t.Run("Close already closed issue fails", func(t *testing.T) {
		_, err := executeCmd("close", "--issue-id", issueID)
		if err == nil {
			t.Fatal("expected error closing already closed issue, got nil")
		}
	})

	t.Run("Claim non-existent issue fails", func(t *testing.T) {
		_, err := executeCmd("claim", "--issue-id", "non-existent")
		if err == nil {
			t.Fatal("expected error claiming non-existent issue, got nil")
		}
	})

	t.Run("Close non-existent issue fails", func(t *testing.T) {
		_, err := executeCmd("close", "--issue-id", "non-existent")
		if err == nil {
			t.Fatal("expected error closing non-existent issue, got nil")
		}
	})

	t.Run("Positional argument fallback for claim and close", func(t *testing.T) {
		outNew, err := executeCmd("create", "-t", "Second Item")
		if err != nil {
			t.Fatalf("failed creating second item: %v", err)
		}
		pNew := strings.Split(outNew, "Created issue: ")
		newID := strings.TrimSpace(strings.Split(pNew[1], " ")[0])

		claimOut, err := executeCmd("claim", newID, "-a", "charlie")
		if err != nil {
			t.Fatalf("unexpected error in positional claim: %v", err)
		}
		if !strings.Contains(claimOut, "Claimed issue: "+newID) {
			t.Errorf("expected claim confirmation, got: %s", claimOut)
		}

		closeOut, err := executeCmd("close", newID)
		if err != nil {
			t.Fatalf("unexpected error in positional close: %v", err)
		}
		if !strings.Contains(closeOut, "Closed issue: "+newID) {
			t.Errorf("expected close confirmation, got: %s", closeOut)
		}
	})
}
