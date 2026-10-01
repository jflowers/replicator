package forge

import (
	"strings"
	"testing"
)

func TestSubtaskPrompt(t *testing.T) {
	prompt := SubtaskPrompt(
		"worker-1", "cell-abc", "epic-xyz",
		"Implement auth",
		[]string{"auth.go", "auth_test.go"},
		"Use JWT tokens",
	)

	checks := []string{
		"Implement auth",
		"worker-1",
		"cell-abc",
		"epic-xyz",
		"auth.go",
		"auth_test.go",
		"JWT tokens",
		"comms_reserve",
		"forge_complete",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("prompt missing %q", check)
		}
	}
}

func TestSubtaskPrompt_NoFiles(t *testing.T) {
	prompt := SubtaskPrompt("agent", "id", "epic", "Title", nil, "")
	if strings.Contains(prompt, "Files to Modify") {
		t.Error("should not show files section when empty")
	}
}

func TestSpawnSubtask(t *testing.T) {
	result, err := SpawnSubtask("cell-1", "epic-1", "Do thing", []string{"a.go"}, "desc", "ctx")
	if err != nil {
		t.Fatalf("SpawnSubtask: %v", err)
	}
	if result["status"] != "ready_to_spawn" {
		t.Errorf("status = %v, want %q", result["status"], "ready_to_spawn")
	}
	if result["bead_id"] != "cell-1" {
		t.Errorf("bead_id = %v, want %q", result["bead_id"], "cell-1")
	}
}

func TestSpawnSubtask_MissingBeadID(t *testing.T) {
	_, err := SpawnSubtask("", "epic-1", "title", nil, "", "")
	if err == nil {
		t.Error("expected error for missing bead_id")
	}
}

func TestSpawnSubtask_MissingEpicID(t *testing.T) {
	_, err := SpawnSubtask("cell-1", "", "title", nil, "", "")
	if err == nil {
		t.Error("expected error for missing epic_id")
	}
}

func TestSpawnSubtask_MissingTitle(t *testing.T) {
	_, err := SpawnSubtask("cell-1", "epic-1", "", nil, "", "")
	if err == nil {
		t.Error("expected error for missing title")
	}
}

func TestCompleteSubtask(t *testing.T) {
	store := testStore(t)

	// Create a cell to complete.
	if _, err := store.DB.Exec(
		"INSERT INTO beads (id, title, type, status) VALUES (?, ?, ?, ?)",
		"cell-test", "Test task", "task", "in_progress",
	); err != nil {
		t.Fatalf("insert cell: %v", err)
	}

	result, err := CompleteSubtask(store, "cell-test", "All done", []string{"a.go"})
	if err != nil {
		t.Fatalf("CompleteSubtask: %v", err)
	}
	if result["status"] != "completed" {
		t.Errorf("status = %v, want %q", result["status"], "completed")
	}

	// Verify event was recorded.
	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM events WHERE type = 'subtask_complete'").Scan(&count); err != nil {
		t.Fatalf("scan completion count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 subtask_complete event, got %d", count)
	}
}

func TestCompleteSubtask_MissingBeadID(t *testing.T) {
	store := testStore(t)
	_, err := CompleteSubtask(store, "", "result", nil)
	if err == nil {
		t.Error("expected error for missing bead_id")
	}
}
