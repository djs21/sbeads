package domain

import (
	"time"

	"sbeads/internal/types"
	"sbeads/backend"
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

// ReadyIssues lists all unblocked issues
func (d *IssueDomain) ReadyIssues() ([]types.Issue, error) {
	issues, err := d.storage.ListIssues()
	if err != nil {
		return nil, err
	}
	return issues, nil
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
