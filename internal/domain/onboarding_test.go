package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureAgentsOnboarding(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Brand new repo without AGENTS.md: creates file
	created, err := EnsureAgentsOnboarding(tempDir)
	if err != nil {
		t.Fatalf("first EnsureAgentsOnboarding failed: %v", err)
	}
	if !created {
		t.Errorf("expected created to be true")
	}

	agentsFile := filepath.Join(tempDir, "AGENTS.md")
	content, err := os.ReadFile(agentsFile)
	if err != nil {
		t.Fatalf("reading AGENTS.md failed: %v", err)
	}
	if !strings.Contains(string(content), "sbeads ready") {
		t.Errorf("missing sbeads instructions in created AGENTS.md: %s", string(content))
	}

	// 2. Second invocation: already has sbeads, idempotent
	second, err := EnsureAgentsOnboarding(tempDir)
	if err != nil {
		t.Fatalf("second EnsureAgentsOnboarding failed: %v", err)
	}
	if second {
		t.Errorf("expected second call to return false")
	}

	// 3. Existing AGENTS.md without sbeads: appends cleanly
	tempDir2 := t.TempDir()
	origFile := filepath.Join(tempDir2, "AGENTS.md")
	if err := os.WriteFile(origFile, []byte("# Existing Custom Agent Rules\n- Rule 1\n"), 0644); err != nil {
		t.Fatalf("writing custom AGENTS.md failed: %v", err)
	}

	appended, err := EnsureAgentsOnboarding(tempDir2)
	if err != nil {
		t.Fatalf("appending EnsureAgentsOnboarding failed: %v", err)
	}
	if !appended {
		t.Errorf("expected appended to be true")
	}

	content2, err := os.ReadFile(origFile)
	if err != nil {
		t.Fatalf("reading appended AGENTS.md failed: %v", err)
	}
	strContent2 := string(content2)
	if !strings.Contains(strContent2, "Existing Custom Agent Rules") || !strings.Contains(strContent2, "Task Tracking with sbeads") {
		t.Errorf("expected both existing rules and sbeads onboarding, got:\n%s", strContent2)
	}
}
