package backend

import (
	"path/filepath"
	"testing"

	"sbeads/internal/types"
)

func TestReadyIssuesFiltersClosed(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_ready_closed.db")
	if err := CreateTables(dbPath); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}
	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	openIssue := &types.Issue{
		ID:        "open-1",
		Title:     "Open Issue",
		IssueType: "task",
		Priority:  1,
		IsClosed:  0,
	}
	closedIssue := &types.Issue{
		ID:        "closed-1",
		Title:     "Closed Issue",
		IssueType: "task",
		Priority:  1,
		IsClosed:  1,
	}
	if err := storage.CreateIssue(openIssue); err != nil {
		t.Fatalf("failed to create open issue: %v", err)
	}
	if err := storage.CreateIssue(closedIssue); err != nil {
		t.Fatalf("failed to create closed issue: %v", err)
	}

	ready, err := storage.GetReadyIssues()
	if err != nil {
		t.Fatalf("GetReadyIssues failed: %v", err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready issue, got %d", len(ready))
	}
	if ready[0].ID != "open-1" {
		t.Errorf("expected issue open-1, got %s", ready[0].ID)
	}
}

func TestReadyIssuesFiltersClaimedAndBlocked(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_ready_claimed_blocked.db")
	if err := CreateTables(dbPath); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}
	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	readyIssue := &types.Issue{
		ID:        "ready-1",
		Title:     "Ready Issue",
		IssueType: "task",
		Priority:  1,
		IsBlocked: 0,
		IsClaimed: 0,
		IsClosed:  0,
	}
	claimedIssue := &types.Issue{
		ID:        "claimed-1",
		Title:     "Claimed Issue",
		IssueType: "task",
		Priority:  1,
		IsBlocked: 0,
		IsClaimed: 1,
		IsClosed:  0,
	}
	blockedIssue := &types.Issue{
		ID:        "blocked-1",
		Title:     "Blocked Issue",
		IssueType: "task",
		Priority:  1,
		IsBlocked: 1,
		IsClaimed: 0,
		IsClosed:  0,
	}

	for _, issue := range []*types.Issue{readyIssue, claimedIssue, blockedIssue} {
		if err := storage.CreateIssue(issue); err != nil {
			t.Fatalf("failed to create issue %s: %v", issue.ID, err)
		}
	}

	ready, err := storage.GetReadyIssues()
	if err != nil {
		t.Fatalf("GetReadyIssues failed: %v", err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected exactly 1 ready issue, got %d", len(ready))
	}
	if ready[0].ID != "ready-1" {
		t.Errorf("expected issue ready-1, got %s", ready[0].ID)
	}
}

func TestReadyIssuesDependencyResolution(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_ready_deps.db")
	if err := CreateTables(dbPath); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}
	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	parent := &types.Issue{
		ID:        "parent-1",
		Title:     "Parent Task",
		IssueType: "task",
		Priority:  2,
		IsClosed:  0,
	}
	child := &types.Issue{
		ID:        "child-1",
		Title:     "Child Task",
		IssueType: "task",
		Priority:  1,
		IsClosed:  0,
	}

	if err := storage.CreateIssue(parent); err != nil {
		t.Fatalf("failed to create parent: %v", err)
	}
	if err := storage.CreateIssue(child); err != nil {
		t.Fatalf("failed to create child: %v", err)
	}
	if err := storage.AddDependency(child.ID, parent.ID); err != nil {
		t.Fatalf("failed to add dependency: %v", err)
	}

	// Initially, child is blocked by open parent
	ready, err := storage.GetReadyIssues()
	if err != nil {
		t.Fatalf("GetReadyIssues failed: %v", err)
	}
	for _, issue := range ready {
		if issue.ID == child.ID {
			t.Fatalf("child %s should not be ready while parent is open", child.ID)
		}
	}

	// Close parent
	if err := storage.CloseIssue(parent.ID); err != nil {
		t.Fatalf("failed to close parent: %v", err)
	}

	// Child must now be ready
	readyAfterClose, err := storage.GetReadyIssues()
	if err != nil {
		t.Fatalf("GetReadyIssues after close failed: %v", err)
	}
	foundChild := false
	for _, issue := range readyAfterClose {
		if issue.ID == child.ID {
			foundChild = true
			break
		}
	}
	if !foundChild {
		t.Fatalf("expected child %s to be ready after parent is closed", child.ID)
	}
}
