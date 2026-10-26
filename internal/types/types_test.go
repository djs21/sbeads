package types

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		t.Fatalf("expected format <prefix>-<hash3>, got %s", id)
	}
	if parts[0] != "sbeads" {
		t.Errorf("expected default prefix 'sbeads', got %s", parts[0])
	}
	if len(parts[1]) != 3 {
		t.Errorf("expected 3-character hash, got length %d (%s)", len(parts[1]), parts[1])
	}

	const base36 = "0123456789abcdefghijklmnopqrstuvwxyz"
	for _, ch := range parts[1] {
		if !strings.ContainsRune(base36, ch) {
			t.Errorf("expected base36 character, got %c in %s", ch, parts[1])
		}
	}
}

func TestGenerateIDWithPrefix(t *testing.T) {
	id := GenerateIDWithPrefix("issue")
	if !strings.HasPrefix(id, "issue-") {
		t.Errorf("expected prefix 'issue-', got %s", id)
	}
	hash := strings.TrimPrefix(id, "issue-")
	if len(hash) != 3 {
		t.Errorf("expected 3-character hash, got %s", hash)
	}
}

func TestGenerateID_Randomness(t *testing.T) {
	seen := make(map[string]bool)
	duplicates := 0
	const iterations = 100
	for i := 0; i < iterations; i++ {
		id := GenerateID()
		if seen[id] {
			duplicates++
		}
		seen[id] = true
	}
	// With 46,656 combinations and 100 samples, duplicate count should be low (< 5)
	if duplicates > 5 {
		t.Errorf("too many duplicates (%d) in %d iterations", duplicates, iterations)
	}
}

