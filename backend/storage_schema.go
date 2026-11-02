package backend

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// CreateTables creates the SQLite tables and applies column migrations.
func CreateTables(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating database directory: %w", err)
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", path))
	if err != nil {
		return err
	}
	defer db.Close()

	schema := `CREATE TABLE IF NOT EXISTS issues (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		issue_type TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 0,
		description TEXT,
		is_blocked INTEGER DEFAULT 0,
		is_closed INTEGER DEFAULT 0,
		is_claimed INTEGER DEFAULT 0,
		agent TEXT DEFAULT '',
		assignee TEXT DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS dependencies (
		child_id TEXT NOT NULL,
		parent_id TEXT NOT NULL,
		PRIMARY KEY (child_id, parent_id)
	);
	CREATE INDEX IF NOT EXISTS idx_dependencies_parent ON dependencies(parent_id);
	CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(is_closed, is_claimed, is_blocked);`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Idempotent column migrations for existing tables
	_, _ = db.Exec("ALTER TABLE issues ADD COLUMN agent TEXT DEFAULT '';")
	_, _ = db.Exec("ALTER TABLE issues ADD COLUMN assignee TEXT DEFAULT '';")
	_, _ = db.Exec("ALTER TABLE issues ADD COLUMN mutation_hash TEXT DEFAULT '';")
	return nil
}
