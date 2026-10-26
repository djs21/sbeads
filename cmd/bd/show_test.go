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

	t.Run("Show issue with flag and open status", func(t *testing.T) {
		out, err := executeCmd("show", "--issue-id", issueID)
		if err != nil {
			t.Fatalf("unexpected error showing issue: %v", err)
		}
		expectedLines := []string{
			"Issue: " + issueID,
			"Title: Detailed Task",
			"Type: feature",
			"Priority: 2",
			"Status: OPEN [O]",
			"Description: Show test description",
		}
		for _, line := range expectedLines {
			if !strings.Contains(out, line) {
				t.Errorf("expected %q in show output, got: %s", line, out)
			}
		}
		if strings.Contains(out, "Blocked by:") {
			t.Errorf("expected no 'Blocked by:' section, got: %s", out)
		}
		if strings.Contains(out, "Blocks:") {
			t.Errorf("expected no 'Blocks:' section, got: %s", out)
		}
	})

	t.Run("Show claimed issue displays CLAIMED status and Agent", func(t *testing.T) {
		claimCreateOut, err := executeCmd("create", "-t", "Claimed Task", "-p", "1")
		if err != nil {
			t.Fatalf("failed to create claimable issue: %v", err)
		}
		partsClaim := strings.Split(claimCreateOut, "Created issue: ")
		claimID := strings.TrimSpace(strings.Split(partsClaim[1], " ")[0])

		if _, err := executeCmd("claim", claimID, "--agent", "worker-42"); err != nil {
			t.Fatalf("failed to claim issue: %v", err)
		}

		out, err := executeCmd("show", claimID)
		if err != nil {
			t.Fatalf("unexpected error showing claimed issue: %v", err)
		}
		if !strings.Contains(out, "Status: CLAIMED [C]") {
			t.Errorf("expected 'Status: CLAIMED [C]', got: %s", out)
		}
		if !strings.Contains(out, "Agent: @worker-42") {
			t.Errorf("expected 'Agent: @worker-42', got: %s", out)
		}
	})

	t.Run("Show issue with Blocked by and Blocks dependency trees", func(t *testing.T) {
		parentOut, err := executeCmd("create", "-t", "Parent Blocker Task", "-p", "3")
		if err != nil {
			t.Fatalf("failed to create parent issue: %v", err)
		}
		parentID := strings.TrimSpace(strings.Split(strings.Split(parentOut, "Created issue: ")[1], " ")[0])

		targetOut, err := executeCmd("create", "-t", "Target Mid Task", "-p", "2")
		if err != nil {
			t.Fatalf("failed to create target issue: %v", err)
		}
		targetID := strings.TrimSpace(strings.Split(strings.Split(targetOut, "Created issue: ")[1], " ")[0])

		dependentOut, err := executeCmd("create", "-t", "Child Dependent Task", "-p", "1")
		if err != nil {
			t.Fatalf("failed to create dependent issue: %v", err)
		}
		dependentID := strings.TrimSpace(strings.Split(strings.Split(dependentOut, "Created issue: ")[1], " ")[0])

		if _, err := executeCmd("dep", "add", "--child", targetID, "--parent", parentID); err != nil {
			t.Fatalf("failed to add blocker dependency: %v", err)
		}
		if _, err := executeCmd("dep", "add", "--child", dependentID, "--parent", targetID); err != nil {
			t.Fatalf("failed to add dependent dependency: %v", err)
		}

		out, err := executeCmd("show", targetID)
		if err != nil {
			t.Fatalf("unexpected error showing blocked issue: %v", err)
		}
		if !strings.Contains(out, "Status: BLOCKED [B]") {
			t.Errorf("expected 'Status: BLOCKED [B]', got: %s", out)
		}
		if !strings.Contains(out, "Blocked by:") {
			t.Errorf("expected 'Blocked by:' section, got: %s", out)
		}
		if !strings.Contains(out, parentID+": Parent Blocker Task") {
			t.Errorf("expected blocker %s in 'Blocked by:', got: %s", parentID, out)
		}
		if !strings.Contains(out, "Blocks:") {
			t.Errorf("expected 'Blocks:' section, got: %s", out)
		}
		if !strings.Contains(out, dependentID+": Child Dependent Task") {
			t.Errorf("expected dependent %s in 'Blocks:', got: %s", dependentID, out)
		}

		// Close parent blocker: target should unblock and not show Blocked by:
		if _, err := executeCmd("close", parentID); err != nil {
			t.Fatalf("failed to close parent issue: %v", err)
		}

		outAfterClose, err := executeCmd("show", targetID)
		if err != nil {
			t.Fatalf("unexpected error showing target after parent close: %v", err)
		}
		if !strings.Contains(outAfterClose, "Status: OPEN [O]") {
			t.Errorf("expected 'Status: OPEN [O]' after parent closed, got: %s", outAfterClose)
		}
		if strings.Contains(outAfterClose, "Blocked by:") {
			t.Errorf("expected no 'Blocked by:' after parent closed, got: %s", outAfterClose)
		}
		if !strings.Contains(outAfterClose, "Blocks:") {
			t.Errorf("expected 'Blocks:' still present, got: %s", outAfterClose)
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
