package backend

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"sbeads/internal/types"
)

// CanonicalizeIssueString serializes critical issue fields deterministically for HMAC computation.
func CanonicalizeIssueString(issue *types.Issue) string {
	if issue == nil {
		return ""
	}
	// Format: id|title|issue_type|priority|is_blocked|is_closed|is_claimed|agent|assignee|description
	return fmt.Sprintf("%s|%s|%s|%d|%d|%d|%d|%s|%s|%s",
		issue.ID,
		issue.Title,
		issue.IssueType,
		issue.Priority,
		issue.IsBlocked,
		issue.IsClosed,
		issue.IsClaimed,
		issue.Agent,
		issue.Assignee,
		issue.Description,
	)
}

// ComputeIssueHMAC calculates the HMAC-SHA256 signature for an issue using key.
func ComputeIssueHMAC(key []byte, issue *types.Issue) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(CanonicalizeIssueString(issue)))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyIssueHMAC verifies that expectedSignature matches the calculated HMAC-SHA256 signature.
func VerifyIssueHMAC(key []byte, issue *types.Issue, expectedSignature string) bool {
	actualSignature := ComputeIssueHMAC(key, issue)
	return hmac.Equal([]byte(actualSignature), []byte(expectedSignature))
}
