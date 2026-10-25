package domain

import (
	"sbeads/backend"
)

// DependencyDomain handles dependency business logic
type DependencyDomain struct {
	storage *backend.StorageImpl
}

// NewDependencyDomain creates a new dependency domain
func NewDependencyDomain(storage *backend.StorageImpl) *DependencyDomain {
	return &DependencyDomain{storage: storage}
}

// AddDependency adds a dependency (child depends on parent)
func (d *DependencyDomain) AddDependency(childID, parentID string) error {
	if childID == parentID {
		return ErrSelfDependency
	}

	// Check if dependency already exists
	has, err := d.HasDependency(childID, parentID)
	if err != nil {
		return err
	}
	if has {
		return ErrDependencyExists
	}

	// Add dependency
	return d.storage.AddDependency(childID, parentID)
}

// RemoveDependency removes a dependency
func (d *DependencyDomain) RemoveDependency(childID, parentID string) error {
	has, err := d.HasDependency(childID, parentID)
	if err != nil || !has {
		return ErrDependencyNotFound
	}

	return nil
}

// GetDependencies gets all dependencies for an issue
func (d *DependencyDomain) GetDependencies(issueID string) ([]string, error) {
	return d.storage.GetDependencies(issueID)
}

// HasDependency checks if an issue has a dependency on another
func (d *DependencyDomain) HasDependency(childID, parentID string) (bool, error) {
	deps, err := d.GetDependencies(childID)
	if err != nil {
		return false, err
	}
	for _, dep := range deps {
		if dep == parentID {
			return true, nil
		}
	}
	return false, nil
}
