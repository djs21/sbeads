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
	for {
		candidate := filepath.Join(curr, ".sbeads")
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return filepath.Join(candidate, "sbeads.db")
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return filepath.Join(".sbeads", "sbeads.db")
}

func FindDatabasePath(startDir string) string {
	return findDatabasePath(startDir)
}
