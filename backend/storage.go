package backend

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"sbeads/internal/types"
)

// Storage defines the storage interface
type Storage interface {
	CreateIssue(issue *types.Issue) error
	GetIssue(id string) (*types.Issue, error)
	ListIssues() ([]types.Issue, error)
	CloseIssue(id string) error
	ClaimIssue(id, assignee, agent string) error
	GetDependencies(issueID string) ([]string, error)
	AddDependency(childID, parentID string) error
	Close() error
}

// StorageImpl implements the Storage interface
type StorageImpl struct {
	db *sql.DB
}

// NewStorage creates a new database storage
func NewStorage(path string) (*StorageImpl, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_journal_mode=WAL", path))
	if err != nil {
		return nil, err
	}
	return &StorageImpl{db: db}, nil
}

// CreateIssue inserts a new issue
func (s *StorageImpl) CreateIssue(issue *types.Issue) error {
	if issue.CreatedAt.IsZero() {
		issue.CreatedAt = time.Now()
	}
	if issue.UpdatedAt.IsZero() {
		issue.UpdatedAt = time.Now()
	}
	_, err := s.db.Exec("INSERT INTO issues (id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		issue.ID, issue.Title, issue.IssueType, issue.Priority, issue.Description,
		issue.IsBlocked, issue.IsClosed, issue.IsClaimed, issue.CreatedAt.Format(time.RFC3339), issue.UpdatedAt.Format(time.RFC3339))
	return err
}

// GetIssue retrieves an issue by ID
func (s *StorageImpl) GetIssue(id string) (*types.Issue, error) {
	var issue types.Issue
	var createdAtStr, updatedAtStr string
	err := s.db.QueryRow("SELECT title, issue_type, priority, description, is_blocked, is_closed, is_claimed, created_at, updated_at FROM issues WHERE id = ?",
		id).Scan(&issue.Title, &issue.IssueType, &issue.Priority, &issue.Description,
		&issue.IsBlocked, &issue.IsClosed, &issue.IsClaimed, &createdAtStr, &updatedAtStr)
	if err != nil {
		return nil, err
	}
	issue.ID = id
	if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
		issue.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
		issue.UpdatedAt = t
	}
	return &issue, nil
}

// ListIssues returns all issues
func (s *StorageImpl) ListIssues() ([]types.Issue, error) {
	var issues []types.Issue
	rows, err := s.db.Query("SELECT id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, created_at, updated_at FROM issues")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var issue types.Issue
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&issue.ID, &issue.Title, &issue.IssueType, &issue.Priority, &issue.Description,
			&issue.IsBlocked, &issue.IsClosed, &issue.IsClaimed, &createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			issue.CreatedAt = t
		}
		if t, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			issue.UpdatedAt = t
		}
		issues = append(issues, issue)
	}
	if err = rows.Err(); err != nil {
		return issues, err
	}
	return issues, nil
}

// CloseIssue marks an issue as closed
func (s *StorageImpl) CloseIssue(id string) error {
	_, err := s.db.Exec("UPDATE issues SET is_closed = ?, updated_at = ? WHERE id = ?",
		1, time.Now().Format(time.RFC3339), id)
	return err
}

// ClaimIssue claims an issue for an agent
func (s *StorageImpl) ClaimIssue(id, assignee, agent string) error {
	_, err := s.db.Exec("UPDATE issues SET is_claimed = 1, is_blocked = 1, updated_at = ? WHERE id = ?",
		time.Now().Format(time.RFC3339), id)
	return err
}

// GetDependencies returns all dependencies for an issue
func (s *StorageImpl) GetDependencies(issueID string) ([]string, error) {
	var deps []string
	rows, err := s.db.Query("SELECT parent_id FROM dependencies WHERE child_id = ?", issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var depID string
		if err := rows.Scan(&depID); err != nil {
			return nil, err
		}
		deps = append(deps, depID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return deps, nil
}

// AddDependency adds a dependency
func (s *StorageImpl) AddDependency(childID, parentID string) error {
	_, err := s.db.Exec("INSERT INTO dependencies (child_id, parent_id) VALUES (?, ?)", childID, parentID)
	return err
}

// Close closes the database connection
func (s *StorageImpl) Close() error {
	return s.db.Close()
}

// CreateTables creates the SQLite tables
func CreateTables(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating database directory: %w", err)
	}
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_journal_mode=WAL", path))
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
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS dependencies (
		child_id TEXT NOT NULL,
		parent_id TEXT NOT NULL,
		PRIMARY KEY (child_id, parent_id)
	);`
	_, err = db.Exec(schema)
	return err
}
