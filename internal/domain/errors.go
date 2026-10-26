package domain

import "errors"

var (
	// Issue errors
	ErrEmptyTitle        = errors.New("empty title")
	ErrNegativePriority  = errors.New("priority must be non-negative")
	ErrClosedBlocked     = errors.New("cannot create closed issue as blocked")
	ErrAlreadyClosed     = errors.New("issue already closed")
	ErrAlreadyClaimed    = errors.New("issue already claimed")
	ErrEmptyID           = errors.New("id cannot be empty")
	ErrNotClaimed        = errors.New("issue is not claimed")

	// Dependency errors
	ErrSelfDependency    = errors.New("cannot depend on self")
	ErrDependencyExists  = errors.New("dependency already exists")
	ErrDependencyNotFound = errors.New("dependency not found")
)
