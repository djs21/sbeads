package main

import (
	"strings"
	"testing"
)

func TestDepCommand(t *testing.T) {
	setupTestDB(t)

	// Create two issues for dependency relations
	out1, err := executeCmd("create", "-t", "Parent Task")
	if err != nil {
		t.Fatalf("failed to create parent issue: %v", err)
	}
	p1 := strings.Split(out1, "Created issue: ")
	parentID := strings.TrimSpace(strings.Split(p1[1], " ")[0])

	out2, err := executeCmd("create", "-t", "Child Task")
	if err != nil {
		t.Fatalf("failed to create child issue: %v", err)
	}
	p2 := strings.Split(out2, "Created issue: ")
	childID := strings.TrimSpace(strings.Split(p2[1], " ")[0])

	t.Run("Dep without subcommand fails", func(t *testing.T) {
		out, err := executeCmd("dep")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected dep usage in output, got: %s", out)
		}
	})

	t.Run("Dep with unknown subcommand fails", func(t *testing.T) {
		_, err := executeCmd("dep", "invalid")
		if err == nil || !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("expected unknown command error, got: %v", err)
		}
	})

	t.Run("Dep add without args fails", func(t *testing.T) {
		_, err := executeCmd("dep", "add")
		if err == nil || !strings.Contains(err.Error(), "--child and --parent are required") {
			t.Fatalf("expected child and parent required error, got: %v", err)
		}
	})

	t.Run("Dep list without issue-id fails", func(t *testing.T) {
		_, err := executeCmd("dep", "list")
		if err == nil || !strings.Contains(err.Error(), "--issue-id is required") {
			t.Fatalf("expected issue-id required error, got: %v", err)
		}
	})

	t.Run("Dep list initially empty", func(t *testing.T) {
		out, err := executeCmd("dep", "list", "--issue-id", childID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "No dependencies") {
			t.Errorf("expected 'No dependencies', got: %s", out)
		}
	})

	t.Run("Dep add with flags", func(t *testing.T) {
		out, err := executeCmd("dep", "add", "--child", childID, "--parent", parentID)
		if err != nil {
			t.Fatalf("unexpected error adding dependency: %v", err)
		}
		expected := "Added dependency: " + childID + " -> " + parentID
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q, got %q", expected, out)
		}
	})

	t.Run("Dep list shows added dependency", func(t *testing.T) {
		out, err := executeCmd("dep", "list", "--issue-id", childID)
		if err != nil {
			t.Fatalf("unexpected error listing dependencies: %v", err)
		}
		if !strings.Contains(out, parentID) {
			t.Errorf("expected %s in dep list, got: %s", parentID, out)
		}
	})

	t.Run("Dep add duplicate dependency fails", func(t *testing.T) {
		_, err := executeCmd("dep", "add", "--child", childID, "--parent", parentID)
		if err == nil {
			t.Fatal("expected error adding duplicate dependency, got nil")
		}
	})

	t.Run("Dep add self dependency fails", func(t *testing.T) {
		_, err := executeCmd("dep", "add", "--child", childID, "--parent", childID)
		if err == nil {
			t.Fatal("expected error adding self dependency, got nil")
		}
	})

	t.Run("Dep add with positional arguments", func(t *testing.T) {
		out3, err := executeCmd("create", "-t", "Second Child")
		if err != nil {
			t.Fatalf("failed creating second child: %v", err)
		}
		p3 := strings.Split(out3, "Created issue: ")
		child2ID := strings.TrimSpace(strings.Split(p3[1], " ")[0])

		addOut, err := executeCmd("dep", "add", child2ID, parentID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(addOut, "Added dependency: "+child2ID+" -> "+parentID) {
			t.Errorf("expected positional dep add confirmation, got: %s", addOut)
		}

		listOut, err := executeCmd("dep", "list", child2ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(listOut, parentID) {
			t.Errorf("expected %s in positional dep list, got: %s", parentID, listOut)
		}
	})
}
