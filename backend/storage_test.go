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
