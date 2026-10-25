package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureOutput intercepts stdout for tests
func captureOutput(f func() error) (string, error) {
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outChan <- buf.String()
	}()

	err := f()

	w.Close()
	os.Stdout = origStdout
	out := <-outChan
	return out, err
}

func TestCLIRoutingAndValidation(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	t.Setenv("SBEADS_DB", dbPath)

	t.Run("No arguments returns error", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads"})
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(out, "Usage: sbeads") {
			t.Errorf("expected usage output, got: %s", out)
		}
	})

	t.Run("Unknown command returns error", func(t *testing.T) {
		_, err := captureOutput(func() error {
			return execute([]string{"sbeads", "unknown_cmd"})
		})
		if err == nil || !strings.Contains(err.Error(), "unknown command: unknown_cmd") {
			t.Fatalf("expected unknown command error, got: %v", err)
		}
	})

	t.Run("Version command prints version", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "version"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "sbeads version " + Version
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q in output, got %q", expected, out)
		}
	})

	t.Run("Create without title fails", func(t *testing.T) {
		_, err := captureOutput(func() error {
			return execute([]string{"sbeads", "create"})
		})
		if err == nil || !strings.Contains(err.Error(), "--title is required") {
			t.Fatalf("expected missing title error, got: %v", err)
		}
	})

	t.Run("Claim without issue-id fails", func(t *testing.T) {
		_, err := captureOutput(func() error {
			return execute([]string{"sbeads", "claim"})
		})
		if err == nil || !strings.Contains(err.Error(), "--issue-id is required") {
			t.Fatalf("expected missing issue-id error, got: %v", err)
		}
	})

	t.Run("Close without issue-id fails", func(t *testing.T) {
		_, err := captureOutput(func() error {
			return execute([]string{"sbeads", "close"})
		})
		if err == nil || !strings.Contains(err.Error(), "--issue-id is required") {
			t.Fatalf("expected missing issue-id error, got: %v", err)
		}
	})

	t.Run("Show without issue-id fails", func(t *testing.T) {
		_, err := captureOutput(func() error {
			return execute([]string{"sbeads", "show"})
		})
		if err == nil || !strings.Contains(err.Error(), "--issue-id is required") {
			t.Fatalf("expected missing issue-id error, got: %v", err)
		}
	})

	t.Run("Show non-existent issue fails", func(t *testing.T) {
		_, err := captureOutput(func() error {
			return execute([]string{"sbeads", "show", "--issue-id", "non-existent"})
		})
		if err == nil || !strings.Contains(err.Error(), "issue not found") {
			t.Fatalf("expected issue not found error, got: %v", err)
		}
	})

	t.Run("Dep without subcommand fails", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "dep"})
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(out, "Usage: sbeads dep") {
			t.Errorf("expected dep usage output, got: %s", out)
		}
	})

	t.Run("Dep with unknown subcommand fails", func(t *testing.T) {
		_, err := captureOutput(func() error {
			return execute([]string{"sbeads", "dep", "invalid"})
		})
		if err == nil || !strings.Contains(err.Error(), "unknown dep command: invalid") {
			t.Fatalf("expected unknown dep command error, got: %v", err)
		}
	})

	t.Run("Dep add without args fails", func(t *testing.T) {
		_, err := captureOutput(func() error {
			return execute([]string{"sbeads", "dep", "add"})
		})
		if err == nil || !strings.Contains(err.Error(), "--child and --parent are required") {
			t.Fatalf("expected child and parent required error, got: %v", err)
		}
	})

	t.Run("Dep list without issue-id fails", func(t *testing.T) {
		_, err := captureOutput(func() error {
			return execute([]string{"sbeads", "dep", "list"})
		})
		if err == nil || !strings.Contains(err.Error(), "--issue-id is required") {
			t.Fatalf("expected issue-id required error, got: %v", err)
		}
	})
}

func TestCLIWorkflows(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "workflow.db")
	t.Setenv("SBEADS_DB", dbPath)

	var issue1ID, issue2ID string

	t.Run("Ready on empty DB shows message", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "ready"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "No ready issues") {
			t.Errorf("expected 'No ready issues', got: %s", out)
		}
	})

	t.Run("Create issue with flags", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "create", "--title", "First Issue", "--issue-type", "bug", "--priority", "2", "--description", "First desc"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Created issue: ") {
			t.Fatalf("expected 'Created issue: ', got: %s", out)
		}
		parts := strings.Split(out, "Created issue: ")
		if len(parts) >= 2 {
			idPart := strings.Split(parts[1], " ")[0]
			issue1ID = strings.TrimSpace(idPart)
		}
		if issue1ID == "" {
			t.Fatalf("failed to extract issue ID from output: %s", out)
		}
	})

	t.Run("Create issue with positional arg", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "create", "Second Issue", "--priority", "1"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		parts := strings.Split(out, "Created issue: ")
		if len(parts) >= 2 {
			idPart := strings.Split(parts[1], " ")[0]
			issue2ID = strings.TrimSpace(idPart)
		}
		if issue2ID == "" {
			t.Fatalf("failed to extract issue ID from output: %s", out)
		}
	})

	t.Run("Ready lists both created issues", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "ready"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, issue1ID) || !strings.Contains(out, issue2ID) {
			t.Errorf("expected both issues in ready list, got: %s", out)
		}
	})

	t.Run("Show issue details", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "show", "--issue-id", issue1ID})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Title: First Issue") {
			t.Errorf("expected Title: First Issue, got: %s", out)
		}
		if !strings.Contains(out, "Type: bug") {
			t.Errorf("expected Type: bug, got: %s", out)
		}
	})

	t.Run("Claim issue", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "claim", "--issue-id", issue1ID, "--assignee", "bob", "--agent", "testbot"})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Claimed issue: "+issue1ID) {
			t.Errorf("expected claim confirmation, got: %s", out)
		}
	})

	t.Run("Add and list dependency", func(t *testing.T) {
		// Empty dep list first
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "dep", "list", "--issue-id", issue2ID})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "No dependencies") {
			t.Errorf("expected No dependencies, got: %s", out)
		}

		// Add dependency: issue2 depends on issue1
		out, err = captureOutput(func() error {
			return execute([]string{"sbeads", "dep", "add", "--child", issue2ID, "--parent", issue1ID})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Added dependency: "+issue2ID+" -> "+issue1ID) {
			t.Errorf("expected dependency added confirmation, got: %s", out)
		}

		// List dependency
		out, err = captureOutput(func() error {
			return execute([]string{"sbeads", "dep", "list", "--issue-id", issue2ID})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, issue1ID) {
			t.Errorf("expected %s in dep list, got: %s", issue1ID, out)
		}
	})

	t.Run("Close issue", func(t *testing.T) {
		out, err := captureOutput(func() error {
			return execute([]string{"sbeads", "close", "--issue-id", issue1ID})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Closed issue: "+issue1ID) {
			t.Errorf("expected close confirmation, got: %s", out)
		}

		// Verify closed in show
		out, err = captureOutput(func() error {
			return execute([]string{"sbeads", "show", "--issue-id", issue1ID})
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Closed: 1") {
			t.Errorf("expected Closed: 1, got: %s", out)
		}
	})
}

func TestDatabaseResolutionAndAncestorDiscovery(t *testing.T) {
	// 1. Environment variable override: SBEADS_DIR
	t.Run("SBEADS_DIR directory override", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("SBEADS_DIR", tempDir)
		t.Setenv("SBEADS_DB", "")

		got := findDatabasePath(t.TempDir())
		expected := filepath.Join(tempDir, "sbeads.db")
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})

	t.Run("SBEADS_DIR direct db file override", func(t *testing.T) {
		tempDir := t.TempDir()
		customDB := filepath.Join(tempDir, "custom.db")
		t.Setenv("SBEADS_DIR", customDB)
		t.Setenv("SBEADS_DB", "")

		got := findDatabasePath(t.TempDir())
		if got != customDB {
			t.Errorf("expected %s, got %s", customDB, got)
		}
	})

	// 2. Environment variable override: SBEADS_DB
	t.Run("SBEADS_DB direct file override", func(t *testing.T) {
		tempDir := t.TempDir()
		customDB := filepath.Join(tempDir, "direct.db")
		t.Setenv("SBEADS_DIR", "")
		t.Setenv("SBEADS_DB", customDB)

		got := findDatabasePath(t.TempDir())
		if got != customDB {
			t.Errorf("expected %s, got %s", customDB, got)
		}
	})

	t.Run("SBEADS_DB directory override", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("SBEADS_DIR", "")
		t.Setenv("SBEADS_DB", tempDir)

		got := findDatabasePath(t.TempDir())
		expected := filepath.Join(tempDir, "sbeads.db")
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})

	// 3. Ancestor search: walk upwards from current working directory
	t.Run("Ancestor discovery finds .sbeads in parent directory", func(t *testing.T) {
		root := t.TempDir()
		sbeadsDir := filepath.Join(root, ".sbeads")
		if err := os.Mkdir(sbeadsDir, 0755); err != nil {
			t.Fatalf("failed to create .sbeads: %v", err)
		}

		nested := filepath.Join(root, "level1", "level2", "level3")
		if err := os.MkdirAll(nested, 0755); err != nil {
			t.Fatalf("failed to create nested dirs: %v", err)
		}

		t.Setenv("SBEADS_DIR", "")
		t.Setenv("SBEADS_DB", "")

		got := findDatabasePath(nested)
		expected := filepath.Join(sbeadsDir, "sbeads.db")
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})

	// 4. Default when no ancestor has .sbeads
	t.Run("Default path when no .sbeads in ancestors", func(t *testing.T) {
		t.Setenv("SBEADS_DIR", "")
		t.Setenv("SBEADS_DB", "")

		got := findDatabasePath(t.TempDir())
		expected := filepath.Join(".sbeads", "sbeads.db")
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})

	// 5. Automatic creation of .sbeads directory when executing commands
	t.Run("Automatic creation of .sbeads directory during command execution", func(t *testing.T) {
		tempDir := t.TempDir()
		origWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to getwd: %v", err)
		}
		defer os.Chdir(origWd)

		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		t.Setenv("SBEADS_DIR", "")
		t.Setenv("SBEADS_DB", "")

		// Execute create command
		_, err = captureOutput(func() error {
			return execute([]string{"sbeads", "create", "Auto dir issue"})
		})
		if err != nil {
			t.Fatalf("unexpected error executing create: %v", err)
		}

		dotSbeads := filepath.Join(tempDir, ".sbeads")
		fi, err := os.Stat(dotSbeads)
		if err != nil || !fi.IsDir() {
			t.Fatalf("expected .sbeads directory to be created in %s", tempDir)
		}

		dbFile := filepath.Join(dotSbeads, "sbeads.db")
		if _, err := os.Stat(dbFile); err != nil {
			t.Fatalf("expected sbeads.db file to exist in %s", dotSbeads)
		}
	})
}
