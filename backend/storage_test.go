package backend

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"sbeads/internal/types"
)

func TestStorage(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_sbeads.db")

	// 1. Test CreateTables idempotency
	t.Run("CreateTables and Idempotency", func(t *testing.T) {
		if err := CreateTables(dbPath); err != nil {
			t.Fatalf("first CreateTables failed: %v", err)
		}

		// Second call must not fail due to duplicate init keys or unhandled errors
		if err := CreateTables(dbPath); err != nil {
			t.Fatalf("second CreateTables (idempotency) failed: %v", err)
		}
	})

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	// 2. Test CreateIssue and GetIssue
	var issueID string
	t.Run("CreateIssue and GetIssue", func(t *testing.T) {
		issue := &types.Issue{
			ID:          "test-issue-1",
			Title:       "Test Issue 1",
			IssueType:   "task",
			Priority:    1,
			Description: "Test Description",
			IsBlocked:   0,
			IsClosed:    0,
			IsClaimed:   0,
			CreatedAt:   time.Now().Truncate(time.Second),
			UpdatedAt:   time.Now().Truncate(time.Second),
		}

		if err := storage.CreateIssue(issue); err != nil {
			t.Fatalf("CreateIssue failed: %v", err)
		}
		issueID = issue.ID

		retrieved, err := storage.GetIssue(issueID)
		if err != nil {
			t.Fatalf("GetIssue failed: %v", err)
		}

		if retrieved.ID != issue.ID {
			t.Errorf("expected ID %s, got %s", issue.ID, retrieved.ID)
		}
		if retrieved.Title != issue.Title {
			t.Errorf("expected Title %s, got %s", issue.Title, retrieved.Title)
		}
		if retrieved.Priority != issue.Priority {
			t.Errorf("expected Priority %d, got %d", issue.Priority, retrieved.Priority)
		}
	})

	// 3. Test AddDependency and GetDependencies
	t.Run("AddDependency and GetDependencies", func(t *testing.T) {
		parentIssue := &types.Issue{
			ID:          "test-issue-parent",
			Title:       "Parent Issue",
			IssueType:   "task",
			Priority:    2,
			Description: "Parent Description",
			IsBlocked:   0,
			IsClosed:    0,
			IsClaimed:   0,
			CreatedAt:   time.Now().Truncate(time.Second),
			UpdatedAt:   time.Now().Truncate(time.Second),
		}
		if err := storage.CreateIssue(parentIssue); err != nil {
			t.Fatalf("CreateIssue for parent failed: %v", err)
		}

		if err := storage.AddDependency(issueID, parentIssue.ID); err != nil {
			t.Fatalf("AddDependency failed: %v", err)
		}

		deps, err := storage.GetDependencies(issueID)
		if err != nil {
			t.Fatalf("GetDependencies failed: %v", err)
		}
		if len(deps) == 0 {
			t.Fatalf("expected at least 1 dependency, got %d", len(deps))
		}
		if deps[0] != parentIssue.ID {
			t.Errorf("expected parent dependency %s, got %s", parentIssue.ID, deps[0])
		}
	})

	// 4. Test ListIssues
	t.Run("ListIssues", func(t *testing.T) {
		issues, err := storage.ListIssues()
		if err != nil {
			t.Fatalf("ListIssues failed: %v", err)
		}
		if len(issues) < 2 {
			t.Fatalf("expected at least 2 issues, got %d", len(issues))
		}
	})

	// 5. Test ClaimIssue and CloseIssue
	t.Run("ClaimIssue and CloseIssue", func(t *testing.T) {
		if err := storage.ClaimIssue(issueID, "assignee1", "agent1"); err != nil {
			t.Fatalf("ClaimIssue failed: %v", err)
		}
		claimed, err := storage.GetIssue(issueID)
		if err != nil {
			t.Fatalf("GetIssue after claim failed: %v", err)
		}
		if claimed.IsClaimed != 1 {
			t.Errorf("expected IsClaimed=1, got %d", claimed.IsClaimed)
		}

		if err := storage.CloseIssue(issueID); err != nil {
			t.Fatalf("CloseIssue failed: %v", err)
		}
		closed, err := storage.GetIssue(issueID)
		if err != nil {
			t.Fatalf("GetIssue after close failed: %v", err)
		}
		if closed.IsClosed != 1 {
			t.Errorf("expected IsClosed=1, got %d", closed.IsClosed)
		}
	})
}

func TestStorage_AutomaticDirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "deep", "nested", ".sbeads")
	dbPath := filepath.Join(nestedDir, "sbeads.db")

	if err := CreateTables(dbPath); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}

	fi, err := os.Stat(nestedDir)
	if err != nil || !fi.IsDir() {
		t.Fatalf("expected directory %s to be created automatically", nestedDir)
	}

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
}

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
		t.Fatalf("expected 1 ready issue, got %d", len(ready))
	}
	if ready[0].ID != "ready-1" {
		t.Errorf("expected ready-1, got %s", ready[0].ID)
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
		t.Fatalf("CreateIssue parent failed: %v", err)
	}
	if err := storage.CreateIssue(child); err != nil {
		t.Fatalf("CreateIssue child failed: %v", err)
	}
	if err := storage.AddDependency(child.ID, parent.ID); err != nil {
		t.Fatalf("AddDependency failed: %v", err)
	}

	// While parent is open, child must NOT be in ready
	ready, err := storage.GetReadyIssues()
	if err != nil {
		t.Fatalf("GetReadyIssues failed: %v", err)
	}
	for _, issue := range ready {
		if issue.ID == child.ID {
			t.Fatalf("child issue %s should NOT be ready while parent is open", child.ID)
		}
	}

	// Close parent
	if err := storage.CloseIssue(parent.ID); err != nil {
		t.Fatalf("CloseIssue parent failed: %v", err)
	}

	// After parent is closed, child MUST be in ready
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

func TestGetIssuePrefixMatch(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_prefix.db")
	if err := CreateTables(dbPath); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}
	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	issue := &types.Issue{
		ID:          "sbeads-123",
		Title:       "Target Task",
		IssueType:   "task",
		Priority:    1,
		Description: "Detailed task",
	}
	if err := storage.CreateIssue(issue); err != nil {
		t.Fatalf("CreateIssue failed: %v", err)
	}

	// Exact match
	exact, err := storage.GetIssue("sbeads-123")
	if err != nil {
		t.Fatalf("GetIssue exact match failed: %v", err)
	}
	if exact.ID != "sbeads-123" || exact.Title != "Target Task" {
		t.Errorf("unexpected exact result: %+v", exact)
	}

	// Short suffix/hash match: "123"
	byShort, err := storage.GetIssue("123")
	if err != nil {
		t.Fatalf("GetIssue short match failed: %v", err)
	}
	if byShort.ID != "sbeads-123" {
		t.Errorf("expected ID sbeads-123, got %s", byShort.ID)
	}

	// Prefix match: "sbeads-1"
	byPrefix, err := storage.GetIssue("sbeads-1")
	if err != nil {
		t.Fatalf("GetIssue prefix match failed: %v", err)
	}
	if byPrefix.ID != "sbeads-123" {
		t.Errorf("expected ID sbeads-123, got %s", byPrefix.ID)
	}

	// Non-existent ID
	if _, err := storage.GetIssue("nonexistent"); err == nil {
		t.Error("expected error for nonexistent ID, got nil")
	}
}

