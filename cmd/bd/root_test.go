package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	setupTestDB(t)

	t.Run("No arguments returns error and usage", func(t *testing.T) {
		out, err := executeCmd()
		if err == nil {
			t.Fatal("expected error when no arguments provided, got nil")
		}
		if !strings.Contains(out, "Usage: sbeads") {
			t.Errorf("expected usage output, got: %s", out)
		}
	})

	t.Run("Unknown command returns error", func(t *testing.T) {
		_, err := executeCmd("unknown_cmd")
		if err == nil || !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("expected unknown command error, got: %v", err)
		}
	})

	t.Run("Version command prints version", func(t *testing.T) {
		out, err := executeCmd("version")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "sbeads version " + Version
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q in output, got %q", expected, out)
		}
	})

	t.Run("Root help flag prints help", func(t *testing.T) {
		out, err := executeCmd("--help")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Usage:") {
			t.Errorf("expected Usage in help output, got: %s", out)
		}
	})

	t.Run("Subcommand help syntax", func(t *testing.T) {
		out, err := executeCmd("help", "create")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "create [title]") {
			t.Errorf("expected create usage in help output, got: %s", out)
		}
	})

	t.Run("Persistent db flag overrides default db", func(t *testing.T) {
		customDB := filepath.Join(t.TempDir(), "custom_flag.db")
		out, err := executeCmd("--db", customDB, "create", "-t", "Flag DB Issue")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Created issue: ") {
			t.Errorf("expected created issue confirmation, got: %s", out)
		}
		if _, err := os.Stat(customDB); err != nil {
			t.Errorf("expected custom database file to exist at %s", customDB)
		}
	})
}

func TestDatabaseResolutionAndAncestorDiscovery(t *testing.T) {
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

	t.Run("Default path when no .sbeads in ancestors", func(t *testing.T) {
		t.Setenv("SBEADS_DIR", "")
		t.Setenv("SBEADS_DB", "")

		got := findDatabasePath(t.TempDir())
		expected := filepath.Join(".sbeads", "sbeads.db")
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})

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

		out, err := executeCmd("create", "Auto dir issue")
		if err != nil {
			t.Fatalf("unexpected error executing create: %v (out: %s)", err, out)
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
