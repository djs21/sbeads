package backend

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"sbeads/internal/types"
)

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

// ImportJSONL reads issues and dependency records from JSONL and writes them into storage.
func (s *StorageImpl) ImportJSONL(r io.Reader) (int, int, error) {
	scanner := bufio.NewScanner(r)
	var issuesCount, depsCount int

	tx, err := s.db.Begin()
	if err != nil {
		return 0, 0, fmt.Errorf("starting import transaction: %w", err)
	}
	defer tx.Rollback()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return 0, 0, fmt.Errorf("parsing json line: %w", err)
		}

		recordType, _ := raw["record_type"].(string)
		if recordType == "dependency" {
			childID, _ := raw["child_id"].(string)
			parentID, _ := raw["parent_id"].(string)
			if childID != "" && parentID != "" {
				_, err := tx.Exec("INSERT OR IGNORE INTO dependencies (child_id, parent_id) VALUES (?, ?)", childID, parentID)
				if err != nil {
					return 0, 0, fmt.Errorf("inserting dependency %s -> %s: %w", childID, parentID, err)
				}
				depsCount++
			}
			continue
		}

		// Otherwise treat as issue
		id, _ := raw["id"].(string)
		title, _ := raw["title"].(string)
		issueType, _ := raw["issue_type"].(string)
		priorityFloat, _ := raw["priority"].(float64)
		desc, _ := raw["description"].(string)
		blockedFloat, _ := raw["is_blocked"].(float64)
		closedFloat, _ := raw["is_closed"].(float64)
		claimedFloat, _ := raw["is_claimed"].(float64)
		agent, _ := raw["agent"].(string)
		assignee, _ := raw["assignee"].(string)
		createdAtStr, _ := raw["created_at"].(string)
		updatedAtStr, _ := raw["updated_at"].(string)

		if id == "" || title == "" {
			continue
		}

		iss := types.Issue{
			ID:          id,
			Title:       title,
			IssueType:   issueType,
			Priority:    int(priorityFloat),
			Description: desc,
			IsBlocked:   int(blockedFloat),
			IsClosed:    int(closedFloat),
			IsClaimed:   int(claimedFloat),
			Agent:       agent,
			Assignee:    assignee,
		}

		var hash string
		if len(s.authKey) > 0 {
			hash = ComputeIssueHMAC(s.authKey, &iss)
		}

		if createdAtStr == "" {
			createdAtStr = time.Now().Format(time.RFC3339)
		}
		if updatedAtStr == "" {
			updatedAtStr = time.Now().Format(time.RFC3339)
		}

		_, err := tx.Exec(`INSERT INTO issues (id, title, issue_type, priority, description, is_blocked, is_closed, is_claimed, agent, assignee, created_at, updated_at, mutation_hash)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				title = excluded.title,
				issue_type = excluded.issue_type,
				priority = excluded.priority,
				description = excluded.description,
				is_blocked = excluded.is_blocked,
				is_closed = excluded.is_closed,
				is_claimed = excluded.is_claimed,
				agent = excluded.agent,
				assignee = excluded.assignee,
				updated_at = excluded.updated_at,
				mutation_hash = excluded.mutation_hash`,
			iss.ID, iss.Title, iss.IssueType, iss.Priority, iss.Description,
			iss.IsBlocked, iss.IsClosed, iss.IsClaimed, iss.Agent, iss.Assignee,
			createdAtStr, updatedAtStr, hash)
		if err != nil {
			return 0, 0, fmt.Errorf("upserting issue %s: %w", iss.ID, err)
		}
		issuesCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("reading JSONL stream: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("committing import transaction: %w", err)
	}

	return issuesCount, depsCount, nil
}
