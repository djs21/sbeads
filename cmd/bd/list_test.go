package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func parseCreatedID(t *testing.T, output string) string {
	t.Helper()
	parts := strings.Split(output, "Created issue: ")
	if len(parts) < 2 {
		t.Fatalf("could not parse created issue ID from output: %s", output)
	}
	return strings.Fields(parts[1])[0]
}

func TestListCommand(t *testing.T) {
	setupTestDB(t)

	t.Run("Empty DB shows message or empty JSON", func(t *testing.T) {
		out, err := executeCmd("list")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "No issues found") {
			t.Errorf("expected 'No issues found', got: %s", out)
		}

		jsonOut, err := executeCmd("list", "--json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var list []any
		if err := json.Unmarshal([]byte(jsonOut), &list); err != nil {
			t.Fatalf("expected valid empty json array, got error: %v, raw: %s", err, jsonOut)
		}
		if len(list) != 0 {
			t.Errorf("expected empty array, got %d items", len(list))
		}
	})

	t.Run("Default listing only includes unclosed issues", func(t *testing.T) {
		setupTestDB(t)

		out1, err := executeCmd("create", "-t", "Active Task 1", "-p", "1")
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		id1 := parseCreatedID(t, out1)

		out2, err := executeCmd("create", "-t", "Closed Task", "-p", "2")
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		id2 := parseCreatedID(t, out2)

		out3, err := executeCmd("create", "-t", "Active Task 2", "-p", "3")
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		id3 := parseCreatedID(t, out3)

		// Close task 2
		if _, err := executeCmd("close", id2); err != nil {
			t.Fatalf("close failed: %v", err)
		}

		// Run default list
		listOut, err := executeCmd("list")
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}

		if !strings.Contains(listOut, id1) {
			t.Errorf("expected unclosed task %s in list output: %s", id1, listOut)
		}
		if !strings.Contains(listOut, id3) {
			t.Errorf("expected unclosed task %s in list output: %s", id3, listOut)
		}
		if strings.Contains(listOut, id2) {
			t.Errorf("did not expect closed task %s in default list output: %s", id2, listOut)
		}
	})

	t.Run("Flag --all includes closed and open issues", func(t *testing.T) {
		setupTestDB(t)

		out1, _ := executeCmd("create", "-t", "Open Item", "-p", "1")
		id1 := parseCreatedID(t, out1)

		out2, _ := executeCmd("create", "-t", "Resolved Item", "-p", "1")
		id2 := parseCreatedID(t, out2)
		executeCmd("close", id2)

		allOut, err := executeCmd("list", "--all")
		if err != nil {
			t.Fatalf("list --all failed: %v", err)
		}

		if !strings.Contains(allOut, id1) {
			t.Errorf("expected open task %s in --all output", id1)
		}
		if !strings.Contains(allOut, id2) {
			t.Errorf("expected closed task %s in --all output", id2)
		}
		if !strings.Contains(allOut, "[CLOSED]") {
			t.Errorf("expected [CLOSED] status badge in output: %s", allOut)
		}

		// Short flag -a
		shortAllOut, err := executeCmd("list", "-a")
		if err != nil {
			t.Fatalf("list -a failed: %v", err)
		}
		if !strings.Contains(shortAllOut, id1) || !strings.Contains(shortAllOut, id2) {
			t.Errorf("expected both tasks in -a output")
		}
	})

	t.Run("Flag --closed only shows closed issues", func(t *testing.T) {
		setupTestDB(t)

		out1, _ := executeCmd("create", "-t", "Pending Work", "-p", "1")
		id1 := parseCreatedID(t, out1)

		out2, _ := executeCmd("create", "-t", "Finished Work", "-p", "2")
		id2 := parseCreatedID(t, out2)
		executeCmd("close", id2)

		closedOut, err := executeCmd("list", "--closed")
		if err != nil {
			t.Fatalf("list --closed failed: %v", err)
		}

		if strings.Contains(closedOut, id1) {
			t.Errorf("did not expect open task %s in --closed output", id1)
		}
		if !strings.Contains(closedOut, id2) {
			t.Errorf("expected closed task %s in --closed output: %s", id2, closedOut)
		}
		if !strings.Contains(closedOut, "[CLOSED]") {
			t.Errorf("expected [CLOSED] in output: %s", closedOut)
		}

		// Short flag -c
		shortClosedOut, err := executeCmd("list", "-c")
		if err != nil {
			t.Fatalf("list -c failed: %v", err)
		}
		if !strings.Contains(shortClosedOut, id2) || strings.Contains(shortClosedOut, id1) {
			t.Errorf("unexpected output for -c: %s", shortClosedOut)
		}
	})

	t.Run("Flag --blocked shows explicitly blocked and dependency-blocked issues", func(t *testing.T) {
		setupTestDB(t)

		// 1. Normal open task
		out1, _ := executeCmd("create", "-t", "Free Task", "-p", "1")
		id1 := parseCreatedID(t, out1)

		// 2. Explicitly blocked task
		out2, _ := executeCmd("create", "-t", "Explicit Blocked", "-p", "2", "-b", "1")
		id2 := parseCreatedID(t, out2)

		// 3. Task blocked by unclosed dependency
		out3, _ := executeCmd("create", "-t", "Dep Blocked", "-p", "3")
		id3 := parseCreatedID(t, out3)
		executeCmd("dep", "add", id3, id1)

		blockedOut, err := executeCmd("list", "--blocked")
		if err != nil {
			t.Fatalf("list --blocked failed: %v", err)
		}

		if strings.Contains(blockedOut, id1) {
			t.Errorf("did not expect unblocked task %s in --blocked output", id1)
		}
		if !strings.Contains(blockedOut, id2) {
			t.Errorf("expected explicitly blocked task %s in --blocked output: %s", id2, blockedOut)
		}
		if !strings.Contains(blockedOut, id3) {
			t.Errorf("expected dependency blocked task %s in --blocked output: %s", id3, blockedOut)
		}
		if !strings.Contains(blockedOut, "[BLOCKED]") {
			t.Errorf("expected [BLOCKED] badge in output: %s", blockedOut)
		}

		// Short flag -b
		shortBlockedOut, err := executeCmd("list", "-b")
		if err != nil {
			t.Fatalf("list -b failed: %v", err)
		}
		if !strings.Contains(shortBlockedOut, id2) || !strings.Contains(shortBlockedOut, id3) {
			t.Errorf("unexpected output for -b: %s", shortBlockedOut)
		}
	})

	t.Run("Flag --tree renders hierarchical epic and child indentation", func(t *testing.T) {
		setupTestDB(t)

		// Create an Epic
		outEpic, err := executeCmd("create", "-t", "Observability Epic", "--type", "epic", "-p", "3")
		if err != nil {
			t.Fatalf("create epic failed: %v", err)
		}
		epicID := parseCreatedID(t, outEpic)

		// Create child tasks
		outC1, _ := executeCmd("create", "-t", "Subtask Alpha", "-p", "2")
		c1ID := parseCreatedID(t, outC1)

		outC2, _ := executeCmd("create", "-t", "Subtask Beta", "-p", "1")
		c2ID := parseCreatedID(t, outC2)

		// Add dependencies: Alpha and Beta depend on Epic
		if _, err := executeCmd("dep", "add", c1ID, epicID); err != nil {
			t.Fatalf("dep add failed: %v", err)
		}
		if _, err := executeCmd("dep", "add", c2ID, epicID); err != nil {
			t.Fatalf("dep add failed: %v", err)
		}

		// Create standalone root task
		outRoot, _ := executeCmd("create", "-t", "Standalone Task", "-p", "1")
		rootID := parseCreatedID(t, outRoot)

		treeOut, err := executeCmd("list", "--tree")
		if err != nil {
			t.Fatalf("list --tree failed: %v", err)
		}

		// Epic should have [EPIC] marker
		if !strings.Contains(treeOut, "[EPIC]") {
			t.Errorf("expected [EPIC] tag in tree output: %s", treeOut)
		}

		// Children should have tree connectors (├── or └──)
		hasBranch := strings.Contains(treeOut, "├── ") || strings.Contains(treeOut, "└── ")
		if !hasBranch {
			t.Errorf("expected tree branches '├── ' or '└── ' in output: %s", treeOut)
		}

		// Children should be indented
		lines := strings.Split(treeOut, "\n")
		var foundC1Indented, foundC2Indented, foundRootNotIndented bool
		for _, line := range lines {
			if strings.Contains(line, c1ID) && (strings.HasPrefix(line, "  ├── ") || strings.HasPrefix(line, "  └── ")) {
				foundC1Indented = true
			}
			if strings.Contains(line, c2ID) && (strings.HasPrefix(line, "  ├── ") || strings.HasPrefix(line, "  └── ")) {
				foundC2Indented = true
			}
			if strings.Contains(line, rootID) && !strings.HasPrefix(line, " ") {
				foundRootNotIndented = true
			}
		}

		if !foundC1Indented {
			t.Errorf("expected child %s to be indented under epic in lines:\n%s", c1ID, treeOut)
		}
		if !foundC2Indented {
			t.Errorf("expected child %s to be indented under epic in lines:\n%s", c2ID, treeOut)
		}
		if !foundRootNotIndented {
			t.Errorf("expected root task %s not to be indented in lines:\n%s", rootID, treeOut)
		}

		// Short flag -t
		shortTreeOut, err := executeCmd("list", "-t")
		if err != nil {
			t.Fatalf("list -t failed: %v", err)
		}
		if !strings.Contains(shortTreeOut, "[EPIC]") {
			t.Errorf("expected [EPIC] in -t output: %s", shortTreeOut)
		}
	})

	t.Run("Flag --json returns valid JSON in flat and tree modes", func(t *testing.T) {
		setupTestDB(t)

		outEpic, _ := executeCmd("create", "-t", "Epic JSON", "--type", "epic", "-p", "3")
		epicID := parseCreatedID(t, outEpic)

		outChild, _ := executeCmd("create", "-t", "Child JSON", "-p", "2")
		childID := parseCreatedID(t, outChild)
		executeCmd("dep", "add", childID, epicID)

		// 1. Flat JSON
		flatJSON, err := executeCmd("list", "--json")
		if err != nil {
			t.Fatalf("list --json failed: %v", err)
		}

		var flatItems []map[string]any
		if err := json.Unmarshal([]byte(flatJSON), &flatItems); err != nil {
			t.Fatalf("flat json invalid: %v, raw:\n%s", err, flatJSON)
		}
		if len(flatItems) != 2 {
			t.Errorf("expected 2 items in flat json, got %d", len(flatItems))
		}

		// Verify fields exist
		item := flatItems[0]
		if _, ok := item["id"]; !ok {
			t.Errorf("expected 'id' field in JSON item")
		}
		if _, ok := item["status"]; !ok {
			t.Errorf("expected 'status' field in JSON item")
		}

		// 2. Tree JSON
		treeJSON, err := executeCmd("list", "--tree", "--json")
		if err != nil {
			t.Fatalf("list --tree --json failed: %v", err)
		}

		var treeItems []map[string]any
		if err := json.Unmarshal([]byte(treeJSON), &treeItems); err != nil {
			t.Fatalf("tree json invalid: %v, raw:\n%s", err, treeJSON)
		}

		// Tree should have 1 root (the epic) containing children
		if len(treeItems) != 1 {
			t.Fatalf("expected 1 root epic item in tree json, got %d", len(treeItems))
		}
		rootItem := treeItems[0]
		if rootItem["id"] != epicID {
			t.Errorf("expected root ID %s, got %v", epicID, rootItem["id"])
		}

		children, ok := rootItem["children"].([]any)
		if !ok || len(children) != 1 {
			t.Fatalf("expected 1 child in root children, got: %v", rootItem["children"])
		}
		childItem := children[0].(map[string]any)
		if childItem["id"] != childID {
			t.Errorf("expected child ID %s, got %v", childID, childItem["id"])
		}
	})
}
