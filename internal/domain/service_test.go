package domain

import (
	"strings"
	"testing"
)

func TestServiceDomain(t *testing.T) {
	storage := setupTestStorage(t)
	svc := NewServiceDomain(storage)

	t.Run("GenerateID format", func(t *testing.T) {
		id := GenerateID()
		if !strings.HasPrefix(id, "sbeads-") {
			t.Errorf("expected id to start with 'sbeads-', got %s", id)
		}
		if len(id) != 10 { // "sbeads-" (7) + 3 chars = 10
			t.Errorf("expected id length 10, got %d (%s)", len(id), id)
		}
	})

	var issue1ID, issue2ID string
	t.Run("Create issue", func(t *testing.T) {
		issue1, err := svc.Create("Issue 1", "feature", 3, "Desc 1", 0)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if issue1.Title != "Issue 1" || issue1.IssueType != "feature" || issue1.Priority != 3 {
			t.Errorf("unexpected issue field values: %+v", issue1)
		}
		issue1ID = issue1.ID

		issue2, err := svc.Create("Issue 2", "bug", 1, "Desc 2", 0)
		if err != nil {
			t.Fatalf("Create second issue failed: %v", err)
		}
		issue2ID = issue2.ID
	})

	t.Run("Ready issues list", func(t *testing.T) {
		ready, err := svc.Ready()
		if err != nil {
			t.Fatalf("Ready failed: %v", err)
		}
		if len(ready) < 2 {
			t.Errorf("expected at least 2 ready issues, got %d", len(ready))
		}
	})

	t.Run("Claim issue", func(t *testing.T) {
		claimed, err := svc.Claim(issue1ID, "alice", "agent-x")
		if err != nil {
			t.Fatalf("Claim failed: %v", err)
		}
		if claimed.IsClaimed != 1 {
			t.Errorf("expected IsClaimed=1, got %d", claimed.IsClaimed)
		}
	})

	t.Run("Unclaim issue", func(t *testing.T) {
		unclaimed, err := svc.Unclaim(issue1ID)
		if err != nil {
			t.Fatalf("Unclaim failed: %v", err)
		}
		if unclaimed.IsClaimed != 0 || unclaimed.Agent != "" || unclaimed.Assignee != "" {
			t.Errorf("expected unclaimed issue, got %+v", unclaimed)
		}
	})

	t.Run("AddDependency and GetDependencies", func(t *testing.T) {
		err := svc.AddDependency(issue1ID, issue2ID)
		if err != nil {
			t.Fatalf("AddDependency failed: %v", err)
		}

		deps, err := svc.GetDependencies(issue1ID)
		if err != nil {
			t.Fatalf("GetDependencies failed: %v", err)
		}
		if len(deps) != 1 || deps[0] != issue2ID {
			t.Errorf("expected [%s], got %v", issue2ID, deps)
		}
	})

	t.Run("Close issue", func(t *testing.T) {
		closed, err := svc.Close(issue1ID)
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
		if closed.IsClosed != 1 {
			t.Errorf("expected IsClosed=1, got %d", closed.IsClosed)
		}
	})
}
