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

	t.Run("CreateWithPrefix custom", func(t *testing.T) {
		customIssue, err := svc.CreateWithPrefix("godzilla", "Godzilla Task", "task", 2, "Desc", 0)
		if err != nil {
			t.Fatalf("CreateWithPrefix failed: %v", err)
		}
		if !strings.HasPrefix(customIssue.ID, "godzilla-") {
			t.Errorf("expected ID starting with godzilla-, got %s", customIssue.ID)
		}
	})

	t.Run("RemoveDependency and Delete issue", func(t *testing.T) {
		child, err := svc.Create("Child Task", "task", 1, "", 0)
		if err != nil {
			t.Fatalf("Create child failed: %v", err)
		}
		parent, err := svc.Create("Parent Task", "task", 1, "", 0)
		if err != nil {
			t.Fatalf("Create parent failed: %v", err)
		}
		if err := svc.AddDependency(child.ID, parent.ID); err != nil {
			t.Fatalf("AddDependency failed: %v", err)
		}

		// Deleting parent without cascade should fail
		if err := svc.Delete(parent.ID, false); err == nil {
			t.Fatal("expected error deleting parent issue with active dependents when cascade=false")
		}

		// Remove dependency
		if err := svc.RemoveDependency(child.ID, parent.ID); err != nil {
			t.Fatalf("RemoveDependency failed: %v", err)
		}

		// Now deleting parent without cascade should succeed
		if err := svc.Delete(parent.ID, false); err != nil {
			t.Fatalf("Delete parent failed: %v", err)
		}

		// Deleting with cascade
		p2, err := svc.Create("Parent 2", "task", 1, "", 0)
		if err != nil {
			t.Fatalf("Create p2 failed: %v", err)
		}
		c2, err := svc.Create("Child 2", "task", 1, "", 0)
		if err != nil {
			t.Fatalf("Create c2 failed: %v", err)
		}
		if err := svc.AddDependency(c2.ID, p2.ID); err != nil {
			t.Fatalf("AddDependency p2-c2 failed: %v", err)
		}
		if err := svc.Delete(p2.ID, true); err != nil {
			t.Fatalf("Delete p2 with cascade failed: %v", err)
		}
	})
}

func TestDetectProjectPrefix(t *testing.T) {
	tempDir := t.TempDir()
	// CleanPrefix tests
	if got := CleanPrefix("My_Project-Name 123!"); got != "my-project-name-123" {
		t.Errorf("unexpected CleanPrefix result: %s", got)
	}
	if got := CleanPrefix("---"); got != "sbeads" {
		t.Errorf("expected fallback 'sbeads', got %s", got)
	}

	// Detect in empty directory
	prefix := DetectProjectPrefix(tempDir)
	if prefix == "" {
		t.Errorf("expected non-empty prefix")
	}
}
