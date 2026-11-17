package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sbeads/internal/domain"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
}

// newMainRepoWithWorktree creates a hermetic repository at tmp/alpha-repo plus
// a linked worktree at tmp/feature, and registers cleanup for the worktree.
func newMainRepoWithWorktree(t *testing.T, tmp string) (mainRoot, worktree string) {
	t.Helper()
	mainRoot = filepath.Join(tmp, "alpha-repo")
	if err := os.MkdirAll(mainRoot, 0755); err != nil {
		t.Fatalf("mkdir main repo: %v", err)
	}
	for _, args := range [][]string{
		{"-c", "init.defaultBranch=main", "init", "-q", mainRoot},
		{"-C", mainRoot, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-qm", "init", "--allow-empty"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	worktree = filepath.Join(tmp, "feature")
	if out, err := exec.Command("git", "-C", mainRoot, "worktree", "add", worktree, "-b", "feat").CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\n%s", err, out)
	}
	t.Cleanup(func() {
		_ = exec.Command("git", "-C", mainRoot, "worktree", "remove", "--force", worktree).Run()
	})
	return mainRoot, worktree
}

func TestWorktreeSharesDatabaseWithMainRepo(t *testing.T) {
	requireGit(t)
	mainRoot, worktree := newMainRepoWithWorktree(t, t.TempDir())

	got := findDatabasePath(worktree)
	want := filepath.Join(mainRoot, ".sbeads", "sbeads.db")
	if got != want {
		t.Errorf("findDatabasePath(worktree) = %q, want %q", got, want)
	}
	if strings.HasPrefix(got, worktree) {
		t.Errorf("findDatabasePath resolved into the worktree itself: %q", got)
	}
}

func TestWorktreeLocalSbeadsOverridesSharedDatabase(t *testing.T) {
	requireGit(t)
	_, worktree := newMainRepoWithWorktree(t, t.TempDir())

	if err := os.MkdirAll(filepath.Join(worktree, ".sbeads"), 0755); err != nil {
		t.Fatalf("mkdir worktree .sbeads: %v", err)
	}
	got := findDatabasePath(worktree)
	want := filepath.Join(worktree, ".sbeads", "sbeads.db")
	if got != want {
		t.Errorf("findDatabasePath(worktree with local .sbeads) = %q, want %q", got, want)
	}
}

func TestWorktreePrefixStaysMainRepoName(t *testing.T) {
	requireGit(t)
	_, worktree := newMainRepoWithWorktree(t, t.TempDir())

	if got, want := domain.DetectProjectPrefix(worktree), "alpha-repo"; got != want {
		t.Errorf("DetectProjectPrefix(worktree) = %q, want %q", got, want)
	}
}
