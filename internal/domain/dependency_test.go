package domain

import (
	"testing"

	"sbeads/internal/types"
)

func TestDependencyDomain(t *testing.T) {
	storage := setupTestStorage(t)
	depDomain := NewDependencyDomain(storage)
	issueDomain := NewIssueDomain(storage)

	child := &types.Issue{ID: "child-1", Title: "Child Issue", IssueType: "task"}
	parent := &types.Issue{ID: "parent-1", Title: "Parent Issue", IssueType: "task"}
	parent2 := &types.Issue{ID: "parent-2", Title: "Parent Issue 2", IssueType: "task"}

	if _, err := issueDomain.CreateIssue(child); err != nil {
		t.Fatalf("failed to create child: %v", err)
	}
	if _, err := issueDomain.CreateIssue(parent); err != nil {
		t.Fatalf("failed to create parent: %v", err)
	}
	if _, err := issueDomain.CreateIssue(parent2); err != nil {
		t.Fatalf("failed to create parent2: %v", err)
	}

	t.Run("AddDependency self error", func(t *testing.T) {
		err := depDomain.AddDependency(child.ID, child.ID)
		if err != ErrSelfDependency {
			t.Fatalf("expected ErrSelfDependency, got %v", err)
		}
	})

	t.Run("AddDependency success", func(t *testing.T) {
		err := depDomain.AddDependency(child.ID, parent.ID)
		if err != nil {
			t.Fatalf("unexpected error adding dependency: %v", err)
		}

		has, err := depDomain.HasDependency(child.ID, parent.ID)
		if err != nil {
			t.Fatalf("unexpected error in HasDependency: %v", err)
		}
		if !has {
			t.Fatal("expected HasDependency to return true")
		}
	})

	t.Run("AddDependency duplicate error", func(t *testing.T) {
		err := depDomain.AddDependency(child.ID, parent.ID)
		if err != ErrDependencyExists {
			t.Fatalf("expected ErrDependencyExists, got %v", err)
		}
	})

	t.Run("HasDependency false for unlinked", func(t *testing.T) {
		has, err := depDomain.HasDependency(child.ID, parent2.ID)
		if err != nil {
			t.Fatalf("unexpected error in HasDependency: %v", err)
		}
		if has {
			t.Fatal("expected HasDependency to return false")
		}
	})

	t.Run("GetDependencies list", func(t *testing.T) {
		deps, err := depDomain.GetDependencies(child.ID)
		if err != nil {
			t.Fatalf("unexpected error getting dependencies: %v", err)
		}
		if len(deps) != 1 || deps[0] != parent.ID {
			t.Fatalf("expected [parent-1], got %v", deps)
		}
	})

	t.Run("RemoveDependency not found error", func(t *testing.T) {
		err := depDomain.RemoveDependency(child.ID, "non-existent-parent")
		if err != ErrDependencyNotFound {
			t.Fatalf("expected ErrDependencyNotFound, got %v", err)
		}
	})

	t.Run("RemoveDependency success", func(t *testing.T) {
		err := depDomain.RemoveDependency(child.ID, parent.ID)
		if err != nil {
			t.Fatalf("unexpected error in RemoveDependency: %v", err)
		}
	})
}
