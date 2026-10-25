package types

import (
	"encoding/hex"
	"testing"
)

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	if len(id1) != 32 {
		t.Errorf("expected 32 character hex string, got length %d (%s)", len(id1), id1)
	}

	if id1 == id2 {
		t.Errorf("expected unique IDs, got duplicate: %s", id1)
	}

	if _, err := hex.DecodeString(id1); err != nil {
		t.Errorf("expected valid hex string, decode error: %v", err)
	}
}
