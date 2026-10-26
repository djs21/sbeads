package main

import (
	"strings"
	"testing"
)

func TestReadyCommand(t *testing.T) {
	setupTestDB(t)

	t.Run("Ready on empty DB shows message", func(t *testing.T) {
		out, err := executeCmd("ready")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "No ready issues") {
			t.Errorf("expected 'No ready issues', got: %s", out)
		}
	})

	t.Run("Ready lists created issues", func(t *testing.T) {
		out1, err := executeCmd("create", "-t", "Issue Alpha", "-p", "1")
		if err != nil {
			t.Fatalf("unexpected error creating alpha: %v", err)
		}
		out2, err := executeCmd("create", "-t", "Issue Beta", "-p", "2")
		if err != nil {
			t.Fatalf("unexpected error creating beta: %v", err)
		}

		parts1 := strings.Split(out1, "Created issue: ")
		id1 := strings.Split(parts1[1], " ")[0]
		parts2 := strings.Split(out2, "Created issue: ")
		id2 := strings.Split(parts2[1], " ")[0]

		readyOut, err := executeCmd("ready")
		if err != nil {
			t.Fatalf("unexpected error running ready: %v", err)
		}

		expected1 := id1 + " [OPEN] (P1/task) Issue Alpha"
		if !strings.Contains(readyOut, expected1) {
			t.Errorf("expected %q in ready list, got: %s", expected1, readyOut)
		}
		expected2 := id2 + " [OPEN] (P2/task) Issue Beta"
		if !strings.Contains(readyOut, expected2) {
			t.Errorf("expected %q in ready list, got: %s", expected2, readyOut)
		}
	})

	t.Run("Ready help flag", func(t *testing.T) {
		out, err := executeCmd("ready", "--help")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "List all ready issues") {
			t.Errorf("expected description in help, got: %s", out)
		}
	})
}
