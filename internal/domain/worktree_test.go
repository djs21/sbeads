package domain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMainWorktreeRoot(t *testing.T) {
	t.Run("normal repo returns its own root", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, ".git"), 0755); err != nil {
			t.Fatalf("create .git: %v", err)
		}
		nested := filepath.Join(root, "a", "b")
		if err := os.MkdirAll(nested, 0755); err != nil {
			t.Fatalf("mkdir nested: %v", err)
		}
		if got, want := MainWorktreeRoot(nested), root; got != want {
			t.Errorf("MainWorktreeRoot = %q, want %q", got, want)
		}
	})

	t.Run("worktree resolves to main working tree root", func(t *testing.T) {
		tmp := t.TempDir()
		mainRoot := filepath.Join(tmp, "main")
		gitDir := filepath.Join(mainRoot, ".git", "worktrees", "feat")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			t.Fatalf("mkdir gitdir: %v", err)
		}
		// Relative commondir marker, as git writes it: <main>/.git/worktrees/../../.git
		if err := os.WriteFile(filepath.Join(gitDir, "commondir"), []byte("../../.git"), 0644); err != nil {
			t.Fatalf("write commondir: %v", err)
		}

		wtRoot := filepath.Join(tmp, "worktree")
		if err := os.MkdirAll(wtRoot, 0755); err != nil {
			t.Fatalf("mkdir worktree: %v", err)
		}
		gitFile := "gitdir: " + gitDir + "\n"
		if err := os.WriteFile(filepath.Join(wtRoot, ".git"), []byte(gitFile), 0644); err != nil {
			t.Fatalf("write .git file: %v", err)
		}

		nested := filepath.Join(wtRoot, "src")
		if err := os.MkdirAll(nested, 0755); err != nil {
			t.Fatalf("mkdir nested: %v", err)
		}
		if got, want := MainWorktreeRoot(nested), mainRoot; got != want {
			t.Errorf("MainWorktreeRoot = %q, want %q", got, want)
		}
	})

	t.Run("submodule is its own project", func(t *testing.T) {
		tmp := t.TempDir()
		// <parent>/.git/modules/sub — a gitdir with NO commondir file.
		modDir := filepath.Join(tmp, "parent", ".git", "modules", "sub")
		if err := os.MkdirAll(modDir, 0755); err != nil {
			t.Fatalf("mkdir modules: %v", err)
		}
		sub := filepath.Join(tmp, "parent", "sub")
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatalf("mkdir sub: %v", err)
		}
		gitFile := "gitdir: " + modDir + "\n"
		if err := os.WriteFile(filepath.Join(sub, ".git"), []byte(gitFile), 0644); err != nil {
			t.Fatalf("write .git file: %v", err)
		}
		if got, want := MainWorktreeRoot(sub), sub; got != want {
			t.Errorf("MainWorktreeRoot = %q, want %q", got, want)
		}
	})

	t.Run("garbage .git file degrades to nearest root", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ".git"), []byte("not a gitdir line"), 0644); err != nil {
			t.Fatalf("write .git: %v", err)
		}
		if got, want := MainWorktreeRoot(root), root; got != want {
			t.Errorf("MainWorktreeRoot = %q, want %q", got, want)
		}
	})

	t.Run("no git entry returns empty", func(t *testing.T) {
		if got := MainWorktreeRoot(t.TempDir()); got != "" {
			t.Errorf("MainWorktreeRoot = %q, want empty", got)
		}
	})
}
