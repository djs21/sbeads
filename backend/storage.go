package backend

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"sbeads/internal/types"
)

// Storage defines the storage interface
type Storage interface {
	CreateIssue(issue *types.Issue) error
	GetIssue(id string) (*types.Issue, error)
	ListIssues() ([]types.Issue, error)
	CloseIssue(id string) error
	ReopenIssue(id string) error
	ClaimIssue(id, assignee, agent string) error
	UnclaimIssue(id string) error
	GetDependencies(issueID string) ([]string, error)
	GetDependents(parentID string) ([]string, error)
	GetDependencyEdges(issueID string) (*DependencyEdges, error)
	GetReadyIssues() ([]types.Issue, error)
	AddDependency(childID, parentID string) error
	RemoveDependency(childID, parentID string) error
	DeleteIssue(id string, cascade bool) error
	UpdateIssue(issue *types.Issue) error
	ExportJSONL(w io.Writer) error
	Close() error
}

// StorageImpl implements the Storage interface
type StorageImpl struct {
	db      *sql.DB
	authKey []byte
}

// NewStorage creates a new database storage
func NewStorage(path string) (*StorageImpl, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	authKey, err := EnsureAuthKey(filepath.Dir(path))
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initializing auth key: %w", err)
	}

	return &StorageImpl{db: db, authKey: authKey}, nil
}

// CreateIssue inserts a new issue with retry on ID collision
func (s *StorageImpl) CreateIssue(issue *types.Issue) error {
	if issue.CreatedAt.IsZero() {
		issue.CreatedAt = time.Now()
	}
	if issue.UpdatedAt.IsZero() {
		issue.UpdatedAt = time.Now()
	}

	const maxRetries = 5
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err = s.db.Exec("INSERT INTO issues (id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, agent, assignee, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			issue.ID, issue.Title, issue.IssueType, issue.Priority, issue.Description,
			issue.IsBlocked, issue.IsClosed, issue.IsClaimed, issue.Agent, issue.Assignee, issue.CreatedAt.Format(time.RFC3339), issue.UpdatedAt.Format(time.RFC3339))
		if err == nil {
			if len(s.authKey) > 0 {
				hash := ComputeIssueHMAC(s.authKey, issue)
				_, _ = s.db.Exec("UPDATE issues SET mutation_hash = ? WHERE id = ?", hash, issue.ID)
			}
			return nil
		}
		// If collision on unique constraint and ID is generated (has prefix), regenerate hash and retry
		if strings.Contains(err.Error(), "UNIQUE constraint failed") && strings.Contains(issue.ID, "-") {
			parts := strings.Split(issue.ID, "-")
			issue.ID = types.GenerateIDWithPrefix(parts[0])
			continue
		}
		return err
	}
	return err
}

// GetIssue retrieves an issue by ID (exact match or prefix/short-ID match)
func (s *StorageImpl) GetIssue(id string) (*types.Issue, error) {
	var issue types.Issue
	var createdAtStr, updatedAtStr, mutationHash string

	// 1. Exact match
	query := "SELECT id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, agent, assignee, created_at, updated_at, COALESCE(mutation_hash, '') FROM issues WHERE id = ?"
	err := s.db.QueryRow(query, id).Scan(&issue.ID, &issue.Title, &issue.IssueType, &issue.Priority, &issue.Description,
		&issue.IsBlocked, &issue.IsClosed, &issue.IsClaimed, &issue.Agent, &issue.Assignee, &createdAtStr, &updatedAtStr, &mutationHash)

	// 2. Fallback to short hash or prefix match (e.g. sbeads-<id>, or <id>%)
	if err == sql.ErrNoRows {
		queryFallback := "SELECT id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, agent, assignee, created_at, updated_at, COALESCE(mutation_hash, '') FROM issues WHERE id LIKE ? OR id LIKE ? ORDER BY length(id) ASC LIMIT 1"
		err = s.db.QueryRow(queryFallback, "%-"+id, id+"%").Scan(&issue.ID, &issue.Title, &issue.IssueType, &issue.Priority, &issue.Description,
			&issue.IsBlocked, &issue.IsClosed, &issue.IsClaimed, &issue.Agent, &issue.Assignee, &createdAtStr, &updatedAtStr, &mutationHash)
	}

	if err != nil {
		return nil, err
	}

	if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
		issue.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
		issue.UpdatedAt = t
	}

	// Verify HMAC if key and hash exist
	if len(s.authKey) > 0 && mutationHash != "" {
		if !VerifyIssueHMAC(s.authKey, &issue, mutationHash) {
			return nil, fmt.Errorf("tamper detected: issue %s was modified outside sbeads CLI", issue.ID)
		}
	}

	return &issue, nil
}

// DependencyEdges holds the parent blockers and child dependents for an issue
type DependencyEdges struct {
	Blockers   []types.Issue
	Dependents []types.Issue
}

func scanIssues(rows *sql.Rows, authKey []byte) ([]types.Issue, error) {
	var issues []types.Issue
	for rows.Next() {
		var issue types.Issue
		var createdAtStr, updatedAtStr, mutationHash string
		if err := rows.Scan(&issue.ID, &issue.Title, &issue.IssueType, &issue.Priority, &issue.Description,
			&issue.IsBlocked, &issue.IsClosed, &issue.IsClaimed, &issue.Agent, &issue.Assignee, &createdAtStr, &updatedAtStr, &mutationHash); err != nil {
			return nil, err
		}
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			issue.CreatedAt = t
		}
		if t, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			issue.UpdatedAt = t
		}

		if len(authKey) > 0 && mutationHash != "" {
			if !VerifyIssueHMAC(authKey, &issue, mutationHash) {
				return nil, fmt.Errorf("tamper detected: issue %s was modified outside sbeads CLI", issue.ID)
			}
		}

		issues = append(issues, issue)
	}
	if err := rows.Err(); err != nil {
		return issues, err
	}
	return issues, nil
}

// ListIssues returns all issues
func (s *StorageImpl) ListIssues() ([]types.Issue, error) {
	rows, err := s.db.Query("SELECT id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, agent, assignee, created_at, updated_at, COALESCE(mutation_hash, '') FROM issues ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIssues(rows, s.authKey)
}

func (s *StorageImpl) recomputeAndSaveHMAC(id string) error {
	if len(s.authKey) == 0 {
		return nil
	}
	var issue types.Issue
	var createdAtStr, updatedAtStr string
	query := "SELECT id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, agent, assignee, created_at, updated_at FROM issues WHERE id = ?"
	err := s.db.QueryRow(query, id).Scan(&issue.ID, &issue.Title, &issue.IssueType, &issue.Priority, &issue.Description,
		&issue.IsBlocked, &issue.IsClosed, &issue.IsClaimed, &issue.Agent, &issue.Assignee, &createdAtStr, &updatedAtStr)
	if err != nil {
		return err
	}
	hash := ComputeIssueHMAC(s.authKey, &issue)
	_, err = s.db.Exec("UPDATE issues SET mutation_hash = ? WHERE id = ?", hash, id)
	return err
}

// CloseIssue marks an issue as closed
func (s *StorageImpl) CloseIssue(id string) error {
	_, err := s.db.Exec("UPDATE issues SET is_closed = ?, updated_at = ? WHERE id = ?",
		1, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return err
	}
	return s.recomputeAndSaveHMAC(id)
}

// ReopenIssue marks a closed issue as open (unclosed)
func (s *StorageImpl) ReopenIssue(id string) error {
	res, err := s.db.Exec("UPDATE issues SET is_closed = 0, updated_at = ? WHERE id = ? AND is_closed = 1",
		time.Now().Format(time.RFC3339), id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("issue %s is not closed or does not exist", id)
	}
	return s.recomputeAndSaveHMAC(id)
}

// ClaimIssue claims an issue for an agent
func (s *StorageImpl) ClaimIssue(id, assignee, agent string) error {
	res, err := s.db.Exec("UPDATE issues SET is_claimed = 1, agent = ?, assignee = ?, updated_at = ? WHERE id = ?",
		agent, assignee, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("issue not found: %s", id)
	}
	return s.recomputeAndSaveHMAC(id)
}

// UnclaimIssue releases a claimed issue back to unclaimed state
func (s *StorageImpl) UnclaimIssue(id string) error {
	res, err := s.db.Exec("UPDATE issues SET is_claimed = 0, agent = '', assignee = '', updated_at = ? WHERE id = ? AND is_claimed = 1",
		time.Now().Format(time.RFC3339), id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		var count int
		if err := s.db.QueryRow("SELECT COUNT(1) FROM issues WHERE id = ?", id).Scan(&count); err == nil && count == 0 {
			return fmt.Errorf("issue not found: %s", id)
		}
		return fmt.Errorf("issue %s is not claimed", id)
	}
	return s.recomputeAndSaveHMAC(id)
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

// GetDependents returns all child issue IDs that depend on the given parentID
func (s *StorageImpl) GetDependents(parentID string) ([]string, error) {
	var dependents []string
	rows, err := s.db.Query("SELECT child_id FROM dependencies WHERE parent_id = ? ORDER BY child_id ASC", parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var childID string
		if err := rows.Scan(&childID); err != nil {
			return nil, err
		}
		dependents = append(dependents, childID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dependents, nil
}

// GetDependencyEdges returns the blockers (parents) and dependents (children) for an issue
func (s *StorageImpl) GetDependencyEdges(issueID string) (*DependencyEdges, error) {
	blockersQuery := `SELECT p.id, p.title, p.issue_type, p.priority, p.description, p.is_blocked, p.is_closed, p.is_claimed, p.agent, p.assignee, p.created_at, p.updated_at, COALESCE(p.mutation_hash, '')
FROM issues p
JOIN dependencies d ON d.parent_id = p.id
WHERE d.child_id = ?
ORDER BY p.priority DESC, p.created_at ASC`

	blockersRows, err := s.db.Query(blockersQuery, issueID)
	if err != nil {
		return nil, fmt.Errorf("querying blockers: %w", err)
	}
	defer blockersRows.Close()

	blockers, err := scanIssues(blockersRows, s.authKey)
	if err != nil {
		return nil, fmt.Errorf("scanning blockers: %w", err)
	}

	dependentsQuery := `SELECT c.id, c.title, c.issue_type, c.priority, c.description, c.is_blocked, c.is_closed, c.is_claimed, c.agent, c.assignee, c.created_at, c.updated_at, COALESCE(c.mutation_hash, '')
FROM issues c
JOIN dependencies d ON d.child_id = c.id
WHERE d.parent_id = ?
ORDER BY c.priority DESC, c.created_at ASC`

	dependentsRows, err := s.db.Query(dependentsQuery, issueID)
	if err != nil {
		return nil, fmt.Errorf("querying dependents: %w", err)
	}
	defer dependentsRows.Close()

	dependents, err := scanIssues(dependentsRows, s.authKey)
	if err != nil {
		return nil, fmt.Errorf("scanning dependents: %w", err)
	}

	return &DependencyEdges{
		Blockers:   blockers,
		Dependents: dependents,
	}, nil
}

// AddDependency adds a dependency
func (s *StorageImpl) AddDependency(childID, parentID string) error {
	_, err := s.db.Exec("INSERT INTO dependencies (child_id, parent_id) VALUES (?, ?)", childID, parentID)
	return err
}

// RemoveDependency removes a specific dependency relationship between child and parent.
func (s *StorageImpl) RemoveDependency(childID, parentID string) error {
	res, err := s.db.Exec("DELETE FROM dependencies WHERE child_id = ? AND parent_id = ?", childID, parentID)
	if err != nil {
		return fmt.Errorf("deleting dependency: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("dependency not found: %s -> %s", childID, parentID)
	}
	return nil
}

// DeleteIssue deletes an issue by ID.
// If cascade is false and other issues depend on this issue, it returns an error.
// If cascade is true, all related dependency rows (both parent and child) are also deleted.
func (s *StorageImpl) DeleteIssue(id string, cascade bool) error {
	var exists int
	err := s.db.QueryRow("SELECT COUNT(1) FROM issues WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("checking issue existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("issue not found: %s", id)
	}

	if !cascade {
		var dependentCount int
		err := s.db.QueryRow("SELECT COUNT(1) FROM dependencies WHERE parent_id = ?", id).Scan(&dependentCount)
		if err != nil {
			return fmt.Errorf("checking dependents: %w", err)
		}
		if dependentCount > 0 {
			return fmt.Errorf("cannot delete issue %s: %d issues depend on it", id, dependentCount)
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM dependencies WHERE child_id = ? OR parent_id = ?", id, id); err != nil {
		return fmt.Errorf("deleting dependencies: %w", err)
	}

	res, err := tx.Exec("DELETE FROM issues WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting issue: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("issue not found: %s", id)
	}

	return tx.Commit()
}

// GetReadyIssues returns ready issues that are open, unclaimed, unblocked,
// and have no unclosed parent dependencies.
func (s *StorageImpl) GetReadyIssues() ([]types.Issue, error) {
	query := `SELECT id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, agent, assignee, created_at, updated_at, COALESCE(mutation_hash, '')
	FROM issues
	WHERE is_closed = 0
	  AND is_claimed = 0
	  AND is_blocked = 0
	  AND NOT EXISTS (
		SELECT 1 FROM dependencies d
		JOIN issues p ON d.parent_id = p.id
		WHERE d.child_id = issues.id AND p.is_closed = 0
	  )
	ORDER BY priority DESC, created_at ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanIssues(rows, s.authKey)
}

// UpdateIssue updates mutable fields of an existing issue.
func (s *StorageImpl) UpdateIssue(issue *types.Issue) error {
	issue.UpdatedAt = time.Now()
	var hash string
	if len(s.authKey) > 0 {
		hash = ComputeIssueHMAC(s.authKey, issue)
	}
	res, err := s.db.Exec(`UPDATE issues 
		SET title = ?, issue_type = ?, priority = ?, description = ?, is_blocked = ?, updated_at = ?, mutation_hash = ? 
		WHERE id = ?`,
		issue.Title, issue.IssueType, issue.Priority, issue.Description, issue.IsBlocked, issue.UpdatedAt.Format(time.RFC3339), hash, issue.ID)
	if err != nil {
		return fmt.Errorf("updating issue: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("issue not found: %s", issue.ID)
	}
	return nil
}

// ExportJSONL writes all issues and dependency rows formatted as line-by-line JSON.
func (s *StorageImpl) ExportJSONL(w io.Writer) error {
	issues, err := s.ListIssues()
	if err != nil {
		return fmt.Errorf("listing issues for export: %w", err)
	}

	encoder := json.NewEncoder(w)
	for _, iss := range issues {
		record := map[string]interface{}{
			"record_type": "issue",
			"id":          iss.ID,
			"title":       iss.Title,
			"issue_type":  iss.IssueType,
			"priority":    iss.Priority,
			"description": iss.Description,
			"is_blocked":  iss.IsBlocked,
			"is_closed":   iss.IsClosed,
			"is_claimed":  iss.IsClaimed,
			"agent":       iss.Agent,
			"assignee":    iss.Assignee,
			"created_at":  iss.CreatedAt.Format(time.RFC3339),
			"updated_at":  iss.UpdatedAt.Format(time.RFC3339),
		}
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("encoding issue %s: %w", iss.ID, err)
		}
	}

	rows, err := s.db.Query("SELECT child_id, parent_id FROM dependencies ORDER BY child_id, parent_id")
	if err != nil {
		return fmt.Errorf("querying dependencies for export: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var childID, parentID string
		if err := rows.Scan(&childID, &parentID); err != nil {
			return fmt.Errorf("scanning dependency for export: %w", err)
		}
		record := map[string]interface{}{
			"record_type": "dependency",
			"child_id":    childID,
			"parent_id":   parentID,
		}
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("encoding dependency: %w", err)
		}
	}

	return rows.Err()
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
