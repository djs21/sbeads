package domain

import (
	"strings"
	"time"

	"sbeads/backend"
	"sbeads/internal/types"
)

// IssueDomain handles issue business logic
type IssueDomain struct {
	storage *backend.StorageImpl
}

// NewIssueDomain creates a new issue domain
func NewIssueDomain(storage *backend.StorageImpl) *IssueDomain {
	return &IssueDomain{storage: storage}
}

// CreateIssue creates a new issue with validation
func (d *IssueDomain) CreateIssue(issue *types.Issue) (*types.Issue, error) {
	// Validate
	if issue.Title == "" {
		return nil, ErrEmptyTitle
	}
	if issue.Priority < 0 {
		return nil, ErrNegativePriority
	}
	if issue.IsBlocked == 0 && issue.IsClosed == 1 {
		return nil, ErrClosedBlocked
	}

	// Persist
	err := d.storage.CreateIssue(issue)
	if err != nil {
		return nil, err
	}

	// Return created issue with timestamps
	issue.CreatedAt = time.Now()
	issue.UpdatedAt = time.Now()
	return issue, nil
}

// CloseIssue closes an issue
func (d *IssueDomain) CloseIssue(id string) (*types.Issue, error) {
	issue, err := d.storage.GetIssue(id)
	if err != nil {
		return nil, err
	}

	// Check if already closed
	if issue.IsClosed == 1 {
		return issue, ErrAlreadyClosed
	}

	// Mark closed
	err = d.storage.CloseIssue(id)
	if err != nil {
		return nil, err
	}

	issue.IsClosed = 1
	issue.UpdatedAt = time.Now()
	return issue, nil
}

// ReadyIssues lists all actionable ready issues (unblocked, unclaimed, unclosed, and dependencies resolved)
func (d *IssueDomain) ReadyIssues() ([]types.Issue, error) {
	return d.storage.GetReadyIssues()
}

// ClaimIssue claims an issue (sets assignee and in_progress)
func (d *IssueDomain) ClaimIssue(id, assignee string, agent string) (*types.Issue, error) {
	issue, err := d.storage.GetIssue(id)
	if err != nil {
		return nil, err
	}

	// Check if already claimed
	if issue.IsClaimed == 1 {
		return issue, ErrAlreadyClaimed
	}

	// Claim the issue
	err = d.storage.ClaimIssue(id, assignee, agent)
	if err != nil {
		return nil, err
	}

	issue.IsClaimed = 1
	issue.UpdatedAt = time.Now()
	return issue, nil
}

// Unclaim releases a claimed issue
func (d *IssueDomain) Unclaim(id string) (*types.Issue, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrEmptyID
	}
	issue, err := d.storage.GetIssue(id)
	if err != nil {
		return nil, err
	}
	if issue.IsClaimed == 0 {
		return nil, ErrNotClaimed
	}
	if err := d.storage.UnclaimIssue(issue.ID); err != nil {
		return nil, err
	}
	issue.IsClaimed = 0
	issue.IsBlocked = 0
	issue.Agent = ""
	issue.Assignee = ""
	issue.UpdatedAt = time.Now()
	return issue, nil
}
