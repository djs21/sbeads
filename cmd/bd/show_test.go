package main

import (
	"strings"
	"testing"
)

func TestShowCommand(t *testing.T) {
	setupTestDB(t)

	createOut, err := executeCmd("create", "-t", "Detailed Task", "-y", "feature", "-p", "2", "-d", "Show test description")
	if err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}
	parts := strings.Split(createOut, "Created issue: ")
	issueID := strings.TrimSpace(strings.Split(parts[1], " ")[0])

	t.Run("Show issue with flag", func(t *testing.T) {
		out, err := executeCmd("show", "--issue-id", issueID)
		if err != nil {
			t.Fatalf("unexpected error showing issue: %v", err)
		}
		expectedLines := []string{
			"Issue: " + issueID,
			"Title: Detailed Task",
			"Type: feature",
			"Priority: 2",
			"Description: Show test description",
			"Blocked: 0",
			"Closed: 0",
			"Claimed: 0",
		}
		for _, line := range expectedLines {
			if !strings.Contains(out, line) {
				t.Errorf("expected %q in show output, got: %s", line, out)
			}
		}
	})

	t.Run("Show issue with positional argument", func(t *testing.T) {
		out, err := executeCmd("show", issueID)
		if err != nil {
			t.Fatalf("unexpected error showing issue: %v", err)
		}
		if !strings.Contains(out, "Title: Detailed Task") {
			t.Errorf("expected Title in output, got: %s", out)
		}
	})

	t.Run("Show without issue-id fails", func(t *testing.T) {
		_, err := executeCmd("show")
		if err == nil || !strings.Contains(err.Error(), "--issue-id is required") {
			t.Fatalf("expected missing issue-id error, got: %v", err)
		}
	})

	t.Run("Show non-existent issue fails", func(t *testing.T) {
		_, err := executeCmd("show", "--issue-id", "non-existent")
		if err == nil || !strings.Contains(err.Error(), "issue not found") {
			t.Fatalf("expected issue not found error, got: %v", err)
		}
	})
}
