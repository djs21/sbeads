package backend

import (
	"testing"
	"time"

	"sbeads/internal/types"
)

func TestHMAC_ComputeAndVerify(t *testing.T) {
	key := []byte("01234567890123456789012345678901") // 32 bytes key

	issue := &types.Issue{
		ID:          "sbeads-abc",
		Title:       "Sample Issue",
		IssueType:   "task",
		Priority:    2,
		Description: "Detailed task specification",
		IsBlocked:   0,
		IsClosed:    0,
		IsClaimed:   0,
		Agent:       "",
		Assignee:    "",
		CreatedAt:   time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC),
	}

	// 1. Compute HMAC
	sig := ComputeIssueHMAC(key, issue)
	if len(sig) != 64 { // SHA-256 in hex string = 64 chars
		t.Fatalf("expected 64 hex characters, got %d (%s)", len(sig), sig)
	}

	// 2. Verify valid HMAC
	if !VerifyIssueHMAC(key, issue, sig) {
		t.Errorf("expected valid HMAC verification to return true")
	}

	// 3. Tamper detection: modifying title must fail verification
	tamperedIssue := *issue
	tamperedIssue.Title = "Tampered Title"
	if VerifyIssueHMAC(key, &tamperedIssue, sig) {
		t.Errorf("expected tampered title to fail verification")
	}

	// 4. Tamper detection: modifying priority must fail verification
	tamperedPriority := *issue
	tamperedPriority.Priority = 9
	if VerifyIssueHMAC(key, &tamperedPriority, sig) {
		t.Errorf("expected tampered priority to fail verification")
	}

	// 5. Tamper detection: modifying is_closed must fail verification
	tamperedClosed := *issue
	tamperedClosed.IsClosed = 1
	if VerifyIssueHMAC(key, &tamperedClosed, sig) {
		t.Errorf("expected tampered is_closed to fail verification")
	}

	// 6. Wrong key must fail verification
	wrongKey := []byte("wrong_key_1234567890123456789012")
	if VerifyIssueHMAC(wrongKey, issue, sig) {
		t.Errorf("expected verification with wrong key to fail")
	}
}
