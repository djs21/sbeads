package main

import (
	"strings"
	"testing"
)

func TestUnclaimCommand(t *testing.T) {
	setupTestDB(t)

	// Create an issue for testing
	createOut, err := executeCmd("create", "-t", "Test Unclaim Issue", "-p", "1")
	if err != nil {
		t.Fatalf("unexpected error creating issue: %v", err)
	}
	parts := strings.Split(createOut, "Created issue: ")
	issueID := strings.TrimSpace(strings.Split(parts[1], " ")[0])

	t.Run("Unclaim without issue-id fails", func(t *testing.T) {
		_, err := executeCmd("unclaim")
		if err == nil || !strings.Contains(err.Error(), "--issue-id is required") {
			t.Fatalf("expected missing issue-id error, got: %v", err)
		}
	})

	t.Run("Unclaim non-existent issue fails", func(t *testing.T) {
		_, err := executeCmd("unclaim", "--issue-id", "non-existent")
		if err == nil {
			t.Fatal("expected error unclaiming non-existent issue, got nil")
		}
	})

	t.Run("Unclaim not-claimed issue fails", func(t *testing.T) {
		_, err := executeCmd("unclaim", "--issue-id", issueID)
		if err == nil {
			t.Fatal("expected error unclaiming not-claimed issue, got nil")
		}
	})

	t.Run("Unclaim with --issue-id flag", func(t *testing.T) {
		// First claim the issue
		_, err := executeCmd("claim", "--issue-id", issueID, "--assignee", "alice", "--agent", "testbot")
		if err != nil {
			t.Fatalf("unexpected error claiming issue: %v", err)
		}

		// Unclaim with --issue-id
		out, err := executeCmd("unclaim", "--issue-id", issueID)
		if err != nil {
			t.Fatalf("unexpected error unclaiming issue: %v", err)
		}

		expected := "Unclaimed issue: " + issueID + " (Test Unclaim Issue)"
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q, got %q", expected, out)
		}

		// Verify unclaiming again fails because it is already unclaimed
		_, err = executeCmd("unclaim", "--issue-id", issueID)
		if err == nil {
			t.Fatal("expected error unclaiming already unclaimed issue, got nil")
		}
	})

	t.Run("Unclaim with -i short flag", func(t *testing.T) {
		// Claim again
		_, err := executeCmd("claim", "-i", issueID, "-a", "bob")
		if err != nil {
			t.Fatalf("unexpected error claiming issue: %v", err)
		}

		// Unclaim with -i
		out, err := executeCmd("unclaim", "-i", issueID)
		if err != nil {
			t.Fatalf("unexpected error unclaiming issue: %v", err)
		}

		expected := "Unclaimed issue: " + issueID + " (Test Unclaim Issue)"
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q, got %q", expected, out)
		}
	})

	t.Run("Unclaim via positional argument", func(t *testing.T) {
		// Claim again
		_, err := executeCmd("claim", issueID)
		if err != nil {
			t.Fatalf("unexpected error claiming issue: %v", err)
		}

		// Unclaim via positional argument
		out, err := executeCmd("unclaim", issueID)
		if err != nil {
			t.Fatalf("unexpected error in positional unclaim: %v", err)
		}

		expected := "Unclaimed issue: " + issueID + " (Test Unclaim Issue)"
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q, got %q", expected, out)
		}
	})
}
