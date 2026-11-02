package backend

import (
	"fmt"

	"sbeads/internal/types"
)

// DependencyEdges holds the parent blockers and child dependents for an issue
type DependencyEdges struct {
	Blockers   []types.Issue
	Dependents []types.Issue
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
