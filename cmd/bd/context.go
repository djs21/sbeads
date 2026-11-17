package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"sbeads/backend"
	"sbeads/internal/domain"
)

type AppContext struct {
	DBPath  string
	Storage *backend.StorageImpl
	Service *domain.ServiceDomain
}

type appContextKey struct{}

func WithAppContext(ctx context.Context, appCtx *AppContext) context.Context {
	return context.WithValue(ctx, appContextKey{}, appCtx)
}

func GetAppContext(ctx context.Context) *AppContext {
	if ctx == nil {
		return nil
	}
	appCtx, _ := ctx.Value(appContextKey{}).(*AppContext)
	return appCtx
}

func getDBPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	return findDatabasePath(cwd)
}

func findDatabasePath(startDir string) string {
	if dir := os.Getenv("SBEADS_DIR"); dir != "" {
		if strings.HasSuffix(dir, ".db") {
			return dir
		}
		if fi, err := os.Stat(dir); err == nil && !fi.IsDir() {
			return dir
		}
		return filepath.Join(dir, "sbeads.db")
	}

	if p := os.Getenv("SBEADS_DB"); p != "" {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return filepath.Join(p, "sbeads.db")
		}
		return p
	}

	curr, err := filepath.Abs(startDir)
	if err != nil {
		curr = startDir
	}

	// Discovery is scoped to a single project. Outside a git repository the
	// search never leaves startDir, so a stray .sbeads/ in a parent directory
	// cannot hijack the database. Inside a repository the search stops at the
	// nearest .git entry, so .sbeads/ directories above it are never used.
	gitRoot := domain.FindGitRoot(curr)
	if gitRoot == "" {
		return filepath.Join(curr, ".sbeads", "sbeads.db")
	}

	for dir := curr; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, ".sbeads")
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return filepath.Join(candidate, "sbeads.db")
		}
		if dir == gitRoot {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	// All worktrees of one repository share a single database at the main
	// working tree root; a .sbeads/ inside a worktree (returned above) is a
	// deliberate per-worktree override.
	return filepath.Join(domain.MainWorktreeRoot(curr), ".sbeads", "sbeads.db")
}

func FindDatabasePath(startDir string) string {
	return findDatabasePath(startDir)
}
