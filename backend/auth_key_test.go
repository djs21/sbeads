package backend

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureAuthKey_CreationAndStrictPermissions(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Initial creation
	key1, err := EnsureAuthKey(tempDir)
	if err != nil {
		t.Fatalf("first EnsureAuthKey failed: %v", err)
	}
	if len(key1) != AuthKeyLength {
		t.Fatalf("expected key length %d, got %d", AuthKeyLength, len(key1))
	}

	keyPath := filepath.Join(tempDir, AuthKeyFileName)
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat on created key failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600 file mode, got %04o", perm)
	}

	// 2. Re-reading existing key with correct permissions
	key2, err := EnsureAuthKey(tempDir)
	if err != nil {
		t.Fatalf("second EnsureAuthKey failed: %v", err)
	}
	if !bytes.Equal(key1, key2) {
		t.Errorf("expected key to be stable across reads")
	}

	// 3. Simulating repository pull with permissive permissions (0644 or 0755)
	if err := os.Chmod(keyPath, 0644); err != nil {
		t.Fatalf("chmod 0644 failed: %v", err)
	}

	key3, err := EnsureAuthKey(tempDir)
	if err != nil {
		t.Fatalf("EnsureAuthKey with permissive permissions failed: %v", err)
	}
	if !bytes.Equal(key1, key3) {
		t.Errorf("expected key content to remain identical")
	}

	// Check that mode was strictly enforced back to 0600
	infoAfter, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat after enforcement failed: %v", err)
	}
	if perm := infoAfter.Mode().Perm(); perm != 0600 {
		t.Errorf("expected permission to be enforced to 0600, got %04o", perm)
	}
}
