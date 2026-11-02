package backend

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sbeads/internal/types"
)

func TestStorage_HMACIntegrityAndTamperDetection(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_hmac.db")

	if err := CreateTables(dbPath); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	issue := &types.Issue{
		ID:          "sbeads-test",
		Title:       "Original Protected Issue",
		IssueType:   "task",
		Priority:    1,
		Description: "Important task",
		IsBlocked:   0,
		IsClosed:    0,
		IsClaimed:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 1. Create issue (should automatically compute mutation_hash)
	if err := storage.CreateIssue(issue); err != nil {
		t.Fatalf("CreateIssue failed: %v", err)
	}

	// 2. Read issue normally: must succeed
	retrieved, err := storage.GetIssue(issue.ID)
	if err != nil {
		t.Fatalf("GetIssue failed on untouched issue: %v", err)
	}
	if retrieved.Title != issue.Title {
		t.Errorf("expected title %s, got %s", issue.Title, retrieved.Title)
	}

	// 3. Tamper directly in SQLite database via raw SQL (simulating external agent edit)
	_, err = storage.db.Exec("UPDATE issues SET title = 'Hacked Title via raw sqlite3' WHERE id = ?", issue.ID)
	if err != nil {
		t.Fatalf("direct SQL update failed: %v", err)
	}

	// 4. GetIssue must now detect tamper and reject!
	_, err = storage.GetIssue(issue.ID)
	if err == nil || !strings.Contains(err.Error(), "tamper detected") {
		t.Fatalf("expected 'tamper detected' error after raw SQL edit, got: %v", err)
	}

	// 5. ListIssues must also detect tamper and report error
	_, err = storage.ListIssues()
	if err == nil || !strings.Contains(err.Error(), "tamper detected") {
		t.Fatalf("expected 'tamper detected' error on ListIssues, got: %v", err)
	}

	// 6. Legitimate CLI mutation (UpdateIssue) must re-sign and succeed
	legitUpdate := &types.Issue{
		ID:          issue.ID,
		Title:       "Legitimately Repaired Title",
		IssueType:   "task",
		Priority:    2,
		Description: "Legit desc",
		IsBlocked:   0,
	}
	if err := storage.UpdateIssue(legitUpdate); err != nil {
		t.Fatalf("UpdateIssue on legitimate mutation failed: %v", err)
	}

	// Read again after legitimate update: must succeed!
	repaired, err := storage.GetIssue(issue.ID)
	if err != nil {
		t.Fatalf("GetIssue after legitimate update failed: %v", err)
	}
	if repaired.Title != "Legitimately Repaired Title" {
		t.Errorf("expected updated title, got: %s", repaired.Title)
	}
}
