package backend

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

const (
	AuthKeyFileName = "auth.key"
	AuthKeyLength   = 32
)

// EnsureAuthKey loads or generates a cryptographic HMAC auth key in keyDir.
// It strictly enforces 0600 file permissions. If permissions are too permissive
// (for instance after being pulled from a git repository on another machine),
// it automatically corrects them to 0600 and logs a security notification to stderr.
func EnsureAuthKey(keyDir string) ([]byte, error) {
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return nil, fmt.Errorf("creating auth key directory: %w", err)
	}

	keyPath := filepath.Join(keyDir, AuthKeyFileName)
	info, err := os.Stat(keyPath)
	if err == nil {
		// Enforce strict 0600 permission
		currentMode := info.Mode().Perm()
		if currentMode != 0600 {
			fmt.Fprintf(os.Stderr, "security notice: %s has permissive permissions (%04o); enforcing 0600\n", keyPath, currentMode)
			if err := os.Chmod(keyPath, 0600); err != nil {
				return nil, fmt.Errorf("enforcing strict 0600 permissions on %s: %w", keyPath, err)
			}
		}

		keyHex, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("reading auth key: %w", err)
		}
		rawKey, err := hex.DecodeString(string(keyHex))
		if err != nil || len(rawKey) != AuthKeyLength {
			return nil, fmt.Errorf("invalid auth key content in %s", keyPath)
		}
		return rawKey, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("checking auth key: %w", err)
	}

	// Generate 32 cryptographically secure random bytes
	rawKey := make([]byte, AuthKeyLength)
	if _, err := rand.Read(rawKey); err != nil {
		return nil, fmt.Errorf("generating random auth key: %w", err)
	}

	keyHex := hex.EncodeToString(rawKey)
	// Write with strict 0600 permissions directly
	if err := os.WriteFile(keyPath, []byte(keyHex), 0600); err != nil {
		return nil, fmt.Errorf("saving auth key: %w", err)
	}

	return rawKey, nil
}
