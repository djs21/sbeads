package domain

import (
	"fmt"
	"strings"

	"sbeads/internal/types"
)

// Status represents the computed lifecycle state of an issue.
type Status string

const (
	// StatusOpen represents an issue that is open and ready to be worked on.
	StatusOpen Status = "OPEN"
	// StatusClaimed represents an issue currently claimed by an agent or assignee.
	StatusClaimed Status = "CLAIMED"
	// StatusBlocked represents an issue blocked explicitly or by unclosed parent dependencies.
	StatusBlocked Status = "BLOCKED"
	// StatusClosed represents an issue that has been resolved or completed.
	StatusClosed Status = "CLOSED"
)

// String returns the string representation of Status.
func (s Status) String() string {
	return string(s)
}

// Badge returns the bracketed status string, e.g. "[OPEN]".
func (s Status) Badge() string {
	return fmt.Sprintf("[%s]", s)
}

// ShortBadge returns a single-character standard badge: [O], [C], [B], or [X].
func (s Status) ShortBadge() string {
	switch s {
	case StatusClosed:
		return "[X]"
	case StatusClaimed:
		return "[C]"
	case StatusBlocked:
		return "[B]"
	default:
		return "[O]"
	}
}

// ComputeStatus calculates the effective status of an issue based on its flags and dependency state.
// Precedence order:
// 1. Closed: issue.IsClosed != 0 -> StatusClosed
// 2. Blocked: issue.IsBlocked != 0 or hasUnclosedDeps -> StatusBlocked
// 3. Claimed: issue.IsClaimed != 0 -> StatusClaimed
// 4. Default -> StatusOpen
func ComputeStatus(issue *types.Issue, hasUnclosedDeps bool) Status {
	if issue == nil {
		return StatusOpen
	}
	if issue.IsClosed != 0 {
		return StatusClosed
	}
	if issue.IsBlocked != 0 || hasUnclosedDeps {
		return StatusBlocked
	}
	if issue.IsClaimed != 0 {
		return StatusClaimed
	}
	return StatusOpen
}

// FormatOneLiner formats an issue into a single standardized line:
// <id> [<STATUS>] <[EPIC] jika epic> (P<priority>/<type>) <title> <[@agent jika ada agent/assignee]>
func FormatOneLiner(issue *types.Issue, status Status) string {
	if issue == nil {
		return ""
	}

	var parts []string
	if issue.ID != "" {
		parts = append(parts, issue.ID)
	}

	parts = append(parts, fmt.Sprintf("[%s]", status))

	if strings.EqualFold(issue.IssueType, "epic") {
		parts = append(parts, "[EPIC]")
	}

	issueType := issue.IssueType
	if issueType == "" {
		issueType = "task"
	}
	parts = append(parts, fmt.Sprintf("(P%d/%s)", issue.Priority, issueType))

	if issue.Title != "" {
		parts = append(parts, issue.Title)
	}

	agent := issue.Agent
	if agent == "" {
		agent = issue.Assignee
	}
	if agent != "" {
		agent = strings.TrimPrefix(agent, "@")
		parts = append(parts, fmt.Sprintf("[@%s]", agent))
	}

	return strings.Join(parts, " ")
}
