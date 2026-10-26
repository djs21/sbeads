package domain

import (
	"sbeads/internal/types"
	"sbeads/backend"
)

// GenerateID generates a unique ID
func GenerateID() string {
	return types.GenerateID()
}

// ServiceDomain combines all domain logic
type ServiceDomain struct {
	issue    *IssueDomain
	dependency *DependencyDomain
}

// NewServiceDomain creates a new service domain
func NewServiceDomain(storage *backend.StorageImpl) *ServiceDomain {
	return &ServiceDomain{
		issue:    NewIssueDomain(storage),
		dependency: NewDependencyDomain(storage),
	}
}

// Create creates a new issue
func (s *ServiceDomain) Create(title, issueType string, priority int, description string, blocked int) (*types.Issue, error) {
	issue := &types.Issue{
		ID:        GenerateID(),
		Title:     title,
		IssueType: issueType,
		Priority:  priority,
		Description: description,
		IsBlocked: blocked,
		IsClosed:  0,
		IsClaimed: 0,
	}

	return s.issue.CreateIssue(issue)
}

// Close closes an issue
func (s *ServiceDomain) Close(id string) (*types.Issue, error) {
	return s.issue.CloseIssue(id)
}

// Ready returns all ready issues
func (s *ServiceDomain) Ready() ([]types.Issue, error) {
	return s.issue.ReadyIssues()
}

// Claim claims an issue
func (s *ServiceDomain) Claim(id string, assignee string, agent string) (*types.Issue, error) {
	return s.issue.ClaimIssue(id, assignee, agent)
}

// AddDependency adds a dependency
func (s *ServiceDomain) AddDependency(childID, parentID string) error {
	return s.dependency.AddDependency(childID, parentID)
}

// GetDependencies returns dependencies for an issue
func (s *ServiceDomain) GetDependencies(issueID string) ([]string, error) {
	return s.dependency.GetDependencies(issueID)
}
