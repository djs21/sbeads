package domain

import (
	"path/filepath"
	"testing"

	"sbeads/backend"
	"sbeads/internal/types"
)

func setupTestStorage(t *testing.T) *backend.StorageImpl {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	if err := backend.CreateTables(dbPath); err != nil {
		t.Fatalf("setup CreateTables failed: %v", err)
	}
	s, err := backend.NewStorage(dbPath)
	if err != nil {
		t.Fatalf("setup NewStorage failed: %v", err)
	}
	t.Cleanup(func() {
		s.Close()
	})
	return s
}

func TestIssueDomain_CreateIssue(t *testing.T) {
	storage := setupTestStorage(t)
	domain := NewIssueDomain(storage)

	tests := []struct {
		name        string
		issue       *types.Issue
		expectedErr error
	}{
		{
			name: "Valid issue",
			issue: &types.Issue{
				ID:        "issue-valid",
				Title:     "Valid Title",
				IssueType: "bug",
				Priority:  1,
			},
			expectedErr: nil,
		},
		{
			name: "Empty title",
			issue: &types.Issue{
				ID:        "issue-no-title",
				Title:     "",
				IssueType: "task",
				Priority:  0,
			},
			expectedErr: ErrEmptyTitle,
		},
		{
			name: "Negative priority",
			issue: &types.Issue{
				ID:        "issue-neg-prio",
				Title:     "Title",
				IssueType: "task",
				Priority:  -1,
			},
			expectedErr: ErrNegativePriority,
		},
		{
			name: "Closed but not blocked",
			issue: &types.Issue{
				ID:        "issue-closed-unblocked",
				Title:     "Title",
				IssueType: "task",
				Priority:  0,
				IsBlocked: 0,
				IsClosed:  1,
			},
			expectedErr: ErrClosedBlocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := domain.CreateIssue(tt.issue)
			if tt.expectedErr != nil {
				if err != tt.expectedErr {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
				if created != nil {
					t.Fatalf("expected nil created issue on error, got %v", created)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if created == nil {
					t.Fatal("expected non-nil created issue")
				}
				if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
					t.Errorf("expected non-zero timestamps, got CreatedAt=%v, UpdatedAt=%v", created.CreatedAt, created.UpdatedAt)
				}
			}
		})
	}
}

func TestIssueDomain_CloseIssue(t *testing.T) {
	storage := setupTestStorage(t)
	domain := NewIssueDomain(storage)

	// Create an issue to close
	issue := &types.Issue{
		ID:        "issue-to-close",
		Title:     "To Close",
		IssueType: "task",
		Priority:  1,
	}
	if _, err := domain.CreateIssue(issue); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	t.Run("Successfully close issue", func(t *testing.T) {
		closed, err := domain.CloseIssue(issue.ID)
		if err != nil {
			t.Fatalf("unexpected error closing issue: %v", err)
		}
		if closed.IsClosed != 1 {
			t.Errorf("expected IsClosed=1, got %d", closed.IsClosed)
		}
	})

	t.Run("Error when closing already closed issue", func(t *testing.T) {
		_, err := domain.CloseIssue(issue.ID)
		if err != ErrAlreadyClosed {
			t.Fatalf("expected ErrAlreadyClosed, got %v", err)
		}
	})

	t.Run("Error closing non-existent issue", func(t *testing.T) {
		_, err := domain.CloseIssue("non-existent-id")
		if err == nil {
			t.Fatal("expected error for non-existent issue, got nil")
		}
	})
}

func TestIssueDomain_ClaimIssue(t *testing.T) {
	storage := setupTestStorage(t)
	domain := NewIssueDomain(storage)

	issue := &types.Issue{
		ID:        "issue-to-claim",
		Title:     "To Claim",
		IssueType: "task",
		Priority:  1,
	}
	if _, err := domain.CreateIssue(issue); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	t.Run("Successfully claim issue", func(t *testing.T) {
		claimed, err := domain.ClaimIssue(issue.ID, "user1", "agent1")
		if err != nil {
			t.Fatalf("unexpected error claiming issue: %v", err)
		}
		if claimed.IsClaimed != 1 {
			t.Errorf("expected IsClaimed=1, got %d", claimed.IsClaimed)
		}
	})

	t.Run("Error when claiming already claimed issue", func(t *testing.T) {
		_, err := domain.ClaimIssue(issue.ID, "user2", "agent2")
		if err != ErrAlreadyClaimed {
			t.Fatalf("expected ErrAlreadyClaimed, got %v", err)
		}
	})

	t.Run("Error claiming non-existent issue", func(t *testing.T) {
		_, err := domain.ClaimIssue("non-existent-id", "u", "a")
		if err == nil {
			t.Fatal("expected error for non-existent issue, got nil")
		}
	})
}

func TestIssueDomain_ReadyIssues(t *testing.T) {
	storage := setupTestStorage(t)
	domain := NewIssueDomain(storage)

	initial, err := domain.ReadyIssues()
	if err != nil {
		t.Fatalf("ReadyIssues failed: %v", err)
	}
	initialCount := len(initial)

	issue := &types.Issue{
		ID:        "ready-issue-1",
		Title:     "Ready Issue",
		IssueType: "task",
		Priority:  2,
	}
	if _, err := domain.CreateIssue(issue); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	issues, err := domain.ReadyIssues()
	if err != nil {
		t.Fatalf("ReadyIssues failed: %v", err)
	}
	if len(issues) != initialCount+1 {
		t.Errorf("expected %d issues, got %d", initialCount+1, len(issues))
	}
}
