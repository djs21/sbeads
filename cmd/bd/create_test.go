package main

import (
	"strings"
	"testing"
)

func TestCreateCommand(t *testing.T) {
	setupTestDB(t)

	t.Run("Create issue with long flags", func(t *testing.T) {
		out, err := executeCmd("create", "--title", "Long Flag Issue", "--issue-type", "bug", "--priority", "2", "--description", "Detailed description", "--blocked", "0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Created issue: ") || !strings.Contains(out, "Long Flag Issue") {
			t.Errorf("expected created issue output, got: %s", out)
		}
	})

	t.Run("Create issue with short flags", func(t *testing.T) {
		out, err := executeCmd("create", "-t", "Short Flag Issue", "-y", "feature", "-p", "1", "-d", "Short desc", "-b", "0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Created issue: ") || !strings.Contains(out, "Short Flag Issue") {
			t.Errorf("expected created issue output, got: %s", out)
		}
	})

	t.Run("Create issue with positional title fallback", func(t *testing.T) {
		out, err := executeCmd("create", "Positional Title Issue", "-p", "3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Created issue: ") || !strings.Contains(out, "Positional Title Issue") {
			t.Errorf("expected created issue output, got: %s", out)
		}
	})

	t.Run("Create without title fails", func(t *testing.T) {
		_, err := executeCmd("create")
		if err == nil || !strings.Contains(err.Error(), "--title is required") {
			t.Fatalf("expected missing title error, got: %v", err)
		}
	})

	t.Run("Create with negative priority fails", func(t *testing.T) {
		_, err := executeCmd("create", "-t", "Negative Priority Issue", "-p", "-1")
		if err == nil || !strings.Contains(err.Error(), "priority must be non-negative") {
			t.Fatalf("expected negative priority error, got: %v", err)
		}
	})
}
