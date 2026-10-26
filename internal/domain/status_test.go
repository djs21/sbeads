package domain

import (
	"testing"

	"sbeads/internal/types"
)

func TestComputeStatus(t *testing.T) {
	tests := []struct {
		name            string
		issue           *types.Issue
		hasUnclosedDeps bool
		want            Status
	}{
		{
			name:            "nil issue defaults to open",
			issue:           nil,
			hasUnclosedDeps: false,
			want:            StatusOpen,
		},
		{
			name: "open issue without blockers or claims",
			issue: &types.Issue{
				ID:        "sbeads-001",
				IsClosed:  0,
				IsBlocked: 0,
				IsClaimed: 0,
			},
			hasUnclosedDeps: false,
			want:            StatusOpen,
		},
		{
			name: "closed issue takes top precedence",
			issue: &types.Issue{
				ID:        "sbeads-002",
				IsClosed:  1,
				IsBlocked: 1,
				IsClaimed: 1,
			},
			hasUnclosedDeps: true,
			want:            StatusClosed,
		},
		{
			name: "closed issue without flags",
			issue: &types.Issue{
				ID:       "sbeads-003",
				IsClosed: 1,
			},
			hasUnclosedDeps: false,
			want:            StatusClosed,
		},
		{
			name: "blocked via is_blocked flag",
			issue: &types.Issue{
				ID:        "sbeads-004",
				IsClosed:  0,
				IsBlocked: 1,
				IsClaimed: 0,
			},
			hasUnclosedDeps: false,
			want:            StatusBlocked,
		},
		{
			name: "blocked via unclosed dependencies",
			issue: &types.Issue{
				ID:        "sbeads-005",
				IsClosed:  0,
				IsBlocked: 0,
				IsClaimed: 0,
			},
			hasUnclosedDeps: true,
			want:            StatusBlocked,
		},
		{
			name: "blocked takes precedence over claimed",
			issue: &types.Issue{
				ID:        "sbeads-006",
				IsClosed:  0,
				IsBlocked: 1,
				IsClaimed: 1,
			},
			hasUnclosedDeps: false,
			want:            StatusBlocked,
		},
		{
			name: "claimed issue without blockers",
			issue: &types.Issue{
				ID:        "sbeads-007",
				IsClosed:  0,
				IsBlocked: 0,
				IsClaimed: 1,
				Agent:     "worker-1",
			},
			hasUnclosedDeps: false,
			want:            StatusClaimed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeStatus(tt.issue, tt.hasUnclosedDeps)
			if got != tt.want {
				t.Errorf("ComputeStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusMethods(t *testing.T) {
	cases := []struct {
		status     Status
		str        string
		badge      string
		shortBadge string
	}{
		{StatusOpen, "OPEN", "[OPEN]", "[O]"},
		{StatusClaimed, "CLAIMED", "[CLAIMED]", "[C]"},
		{StatusBlocked, "BLOCKED", "[BLOCKED]", "[B]"},
		{StatusClosed, "CLOSED", "[CLOSED]", "[X]"},
	}

	for _, c := range cases {
		if c.status.String() != c.str {
			t.Errorf("String() = %v, want %v", c.status.String(), c.str)
		}
		if c.status.Badge() != c.badge {
			t.Errorf("Badge() = %v, want %v", c.status.Badge(), c.badge)
		}
		if c.status.ShortBadge() != c.shortBadge {
			t.Errorf("ShortBadge() = %v, want %v", c.status.ShortBadge(), c.shortBadge)
		}
	}
}

func TestFormatOneLiner(t *testing.T) {
	tests := []struct {
		name   string
		issue  *types.Issue
		status Status
		want   string
	}{
		{
			name:   "nil issue returns empty string",
			issue:  nil,
			status: StatusOpen,
			want:   "",
		},
		{
			name: "open task formatting without agent",
			issue: &types.Issue{
				ID:        "sbeads-123",
				Title:     "Setup project",
				IssueType: "task",
				Priority:  1,
			},
			status: StatusOpen,
			want:   "sbeads-123 [OPEN] (P1/task) Setup project",
		},
		{
			name: "claimed bug with agent",
			issue: &types.Issue{
				ID:        "sbeads-456",
				Title:     "Fix crash",
				IssueType: "bug",
				Priority:  0,
				Agent:     "worker-1",
			},
			status: StatusClaimed,
			want:   "sbeads-456 [CLAIMED] (P0/bug) Fix crash [@worker-1]",
		},
		{
			name: "claimed bug with agent having leading @",
			issue: &types.Issue{
				ID:        "sbeads-456",
				Title:     "Fix crash",
				IssueType: "bug",
				Priority:  0,
				Agent:     "@worker-1",
			},
			status: StatusClaimed,
			want:   "sbeads-456 [CLAIMED] (P0/bug) Fix crash [@worker-1]",
		},
		{
			name: "claimed issue falling back to assignee",
			issue: &types.Issue{
				ID:        "sbeads-456",
				Title:     "Fix crash",
				IssueType: "bug",
				Priority:  0,
				Assignee:  "alice",
			},
			status: StatusClaimed,
			want:   "sbeads-456 [CLAIMED] (P0/bug) Fix crash [@alice]",
		},
		{
			name: "epic issue adds [EPIC] tag",
			issue: &types.Issue{
				ID:        "sbeads-789",
				Title:     "User authentication",
				IssueType: "epic",
				Priority:  2,
			},
			status: StatusOpen,
			want:   "sbeads-789 [OPEN] [EPIC] (P2/epic) User authentication",
		},
		{
			name: "case-insensitive epic issue type",
			issue: &types.Issue{
				ID:        "sbeads-789",
				Title:     "User authentication",
				IssueType: "EPIC",
				Priority:  2,
			},
			status: StatusOpen,
			want:   "sbeads-789 [OPEN] [EPIC] (P2/EPIC) User authentication",
		},
		{
			name: "epic issue with agent",
			issue: &types.Issue{
				ID:        "sbeads-789",
				Title:     "User authentication",
				IssueType: "epic",
				Priority:  2,
				Agent:     "lead",
			},
			status: StatusClaimed,
			want:   "sbeads-789 [CLAIMED] [EPIC] (P2/epic) User authentication [@lead]",
		},
		{
			name: "blocked status formatting",
			issue: &types.Issue{
				ID:        "sbeads-999",
				Title:     "Waiting for upstream API",
				IssueType: "task",
				Priority:  3,
			},
			status: StatusBlocked,
			want:   "sbeads-999 [BLOCKED] (P3/task) Waiting for upstream API",
		},
		{
			name: "closed status formatting",
			issue: &types.Issue{
				ID:        "sbeads-999",
				Title:     "Completed task",
				IssueType: "task",
				Priority:  1,
			},
			status: StatusClosed,
			want:   "sbeads-999 [CLOSED] (P1/task) Completed task",
		},
		{
			name: "empty issue type defaults to task",
			issue: &types.Issue{
				ID:       "sbeads-111",
				Title:    "No type specified",
				Priority: 1,
			},
			status: StatusOpen,
			want:   "sbeads-111 [OPEN] (P1/task) No type specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatOneLiner(tt.issue, tt.status)
			if got != tt.want {
				t.Errorf("FormatOneLiner() = %q, want %q", got, tt.want)
			}
		})
	}
}
