package backend

import (
	"path/filepath"
	"testing"

	"sbeads/internal/types"
)

func TestStorage_AgentAssigneePersistenceAndUnclaim(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_agent_persistence.db")

	if err := CreateTables(dbPath); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}

	// Idempotent migration check: calling CreateTables second time must succeed
	if err := CreateTables(dbPath); err != nil {
		t.Fatalf("Second CreateTables failed: %v", err)
	}

	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()

	issue := &types.Issue{
		ID:        "sbeads-abc",
		Title:     "Agent Task",
		IssueType: "task",
		Priority:  1,
		Agent:     "initial-agent",
		Assignee:  "initial-assignee",
	}

	if err := storage.CreateIssue(issue); err != nil {
		t.Fatalf("CreateIssue failed: %v", err)
	}

	fetched, err := storage.GetIssue("sbeads-abc")
	if err != nil {
		t.Fatalf("GetIssue failed: %v", err)
	}
	if fetched.Agent != "initial-agent" || fetched.Assignee != "initial-assignee" {
		t.Errorf("expected agent/assignee preserved, got agent=%q assignee=%q", fetched.Agent, fetched.Assignee)
	}

	// Test ClaimIssue updates agent and assignee
	if err := storage.ClaimIssue("sbeads-abc", "dev-user", "worker-1"); err != nil {
		t.Fatalf("ClaimIssue failed: %v", err)
	}

	claimed, err := storage.GetIssue("sbeads-abc")
	if err != nil {
		t.Fatalf("GetIssue after claim failed: %v", err)
	}
	if claimed.IsClaimed != 1 || claimed.Agent != "worker-1" || claimed.Assignee != "dev-user" {
		t.Errorf("claim did not update properly: %+v", claimed)
	}

	// Test UnclaimIssue releases task
	if err := storage.UnclaimIssue("sbeads-abc"); err != nil {
		t.Fatalf("UnclaimIssue failed: %v", err)
	}

	unclaimed, err := storage.GetIssue("sbeads-abc")
	if err != nil {
		t.Fatalf("GetIssue after unclaim failed: %v", err)
	}
	if unclaimed.IsClaimed != 0 || unclaimed.Agent != "" || unclaimed.Assignee != "" {
		t.Errorf("unclaim did not reset fields properly: %+v", unclaimed)
	}

	// Unclaim again should return error (not claimed)
	if err := storage.UnclaimIssue("sbeads-abc"); err == nil {
		t.Errorf("expected error unclaiming already unclaimed issue, got nil")
	}

	// Unclaim non-existent issue should return error (not found)
	if err := storage.UnclaimIssue("non-existent"); err == nil {
		t.Errorf("expected error unclaiming non-existent issue, got nil")
	}
}
