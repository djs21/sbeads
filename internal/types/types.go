package types

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateID generates a random ID

// Issue represents a single issue in the tracker
type Issue struct {
	ID            string
	Title         string
	IssueType     string
	Priority      int
	Description   string
	IsBlocked     int
	IsClosed      int
	IsClaimed     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Dependency represents a dependency between issues
type Dependency struct {
	ChildID string
	ParentID string
}

// GenerateID generates a random ID
func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
