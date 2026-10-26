package types

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// Base36Charset contains 36 alphanumeric characters [0-9a-z] (36^3 = 46,656 combinations)
const Base36Charset = "0123456789abcdefghijklmnopqrstuvwxyz"

// Issue represents a single issue in the tracker
type Issue struct {
	ID          string
	Title       string
	IssueType   string
	Priority    int
	Description string
	IsBlocked   int
	IsClosed    int
	IsClaimed   int
	Agent       string
	Assignee    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Dependency represents a dependency between issues
type Dependency struct {
	ChildID  string
	ParentID string
}

// GenerateShortHash generates a cryptographically random 3-character Base36 hash
func GenerateShortHash() string {
	b := make([]byte, 3)
	max := big.NewInt(int64(len(Base36Charset)))
	for i := 0; i < 3; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			b[i] = Base36Charset[i%len(Base36Charset)]
			continue
		}
		b[i] = Base36Charset[n.Int64()]
	}
	return string(b)
}

// GenerateIDWithPrefix generates an ID formatted as <prefix>-<hash3>
func GenerateIDWithPrefix(prefix string) string {
	if prefix == "" {
		prefix = "sbeads"
	}
	return fmt.Sprintf("%s-%s", prefix, GenerateShortHash())
}

// GenerateID generates a default ID formatted as sbeads-xxx
func GenerateID() string {
	return GenerateIDWithPrefix("sbeads")
}

