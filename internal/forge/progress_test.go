package forge

import (
	"testing"
)

func TestProgress(t *testing.T) {
	store := testStore(t)

	err := Progress(store, "proj-1", "worker-1", "cell-1", "in_progress", 50, "halfway done", []string{"a.go"})
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}

	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM events WHERE type = 'forge_progress'").Scan(&count); err != nil {
		t.Fatalf("scan event count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 progress event, got %d", count)
	}
}

func TestProgress_MissingBeadID(t *testing.T) {
	store := testStore(t)
	err := Progress(store, "proj", "agent", "", "in_progress", 0, "", nil)
	if err == nil {
		t.Error("expected error for missing bead_id")
	}
}

func TestComplete(t *testing.T) {
	store := testStore(t)

	// Create a cell to complete.
	if _, err := store.DB.Exec(
		"INSERT INTO beads (id, title, type, status) VALUES (?, ?, ?, ?)",
		"cell-comp", "Complete me", "task", "in_progress",
	); err != nil {
		t.Fatalf("insert cell: %v", err)
	}

	result, err := Complete(store, "proj-1", "worker-1", "cell-comp", "All done", []string{"a.go"}, "good", false, false)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if result["status"] != "completed" {
		t.Errorf("status = %v, want %q", result["status"], "completed")
	}

	// Verify event.
	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM events WHERE type = 'forge_complete'").Scan(&count); err != nil {
		t.Fatalf("scan event count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 complete event, got %d", count)
	}

	// Verify cell status.
	var status string
	if err := store.DB.QueryRow("SELECT status FROM beads WHERE id = 'cell-comp'").Scan(&status); err != nil {
		t.Fatalf("scan cell status: %v", err)
	}
	if status != "closed" {
		t.Errorf("cell status = %q, want %q", status, "closed")
	}
}

func TestComplete_MissingBeadID(t *testing.T) {
	store := testStore(t)
	_, err := Complete(store, "proj", "agent", "", "summary", nil, "", false, false)
	if err == nil {
		t.Error("expected error for missing bead_id")
	}
}

func TestStatus(t *testing.T) {
	store := testStore(t)

	// Create an epic with subtasks.
	statements := []struct {
		query string
		args  []any
	}{
		{"INSERT INTO beads (id, title, type, status) VALUES (?, ?, ?, ?)", []any{"epic-1", "Epic", "epic", "open"}},
		{"INSERT INTO beads (id, title, type, status, parent_id) VALUES (?, ?, ?, ?, ?)", []any{"sub-1", "Sub 1", "task", "open", "epic-1"}},
		{"INSERT INTO beads (id, title, type, status, parent_id) VALUES (?, ?, ?, ?, ?)", []any{"sub-2", "Sub 2", "task", "closed", "epic-1"}},
		{"INSERT INTO beads (id, title, type, status, parent_id) VALUES (?, ?, ?, ?, ?)", []any{"sub-3", "Sub 3", "task", "in_progress", "epic-1"}},
	}
	for _, statement := range statements {
		if _, err := store.DB.Exec(statement.query, statement.args...); err != nil {
			t.Fatalf("insert cell: %v", err)
		}
	}

	result, err := Status(store, "epic-1", "proj-1")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if result["total"] != 3 {
		t.Errorf("total = %v, want 3", result["total"])
	}

	counts := result["counts"].(map[string]int)
	if counts["open"] != 1 {
		t.Errorf("open = %d, want 1", counts["open"])
	}
	if counts["closed"] != 1 {
		t.Errorf("closed = %d, want 1", counts["closed"])
	}
	if counts["in_progress"] != 1 {
		t.Errorf("in_progress = %d, want 1", counts["in_progress"])
	}
}

func TestStatus_MissingEpicID(t *testing.T) {
	store := testStore(t)
	_, err := Status(store, "", "proj")
	if err == nil {
		t.Error("expected error for missing epic_id")
	}
}

func TestStatus_NoSubtasks(t *testing.T) {
	store := testStore(t)

	if _, err := store.DB.Exec("INSERT INTO beads (id, title, type, status) VALUES (?, ?, ?, ?)",
		"epic-empty", "Empty Epic", "epic", "open"); err != nil {
		t.Fatalf("insert epic: %v", err)
	}

	result, err := Status(store, "epic-empty", "proj")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if result["total"] != 0 {
		t.Errorf("total = %v, want 0", result["total"])
	}
}

func TestRecordOutcome(t *testing.T) {
	store := testStore(t)

	err := RecordOutcome(store, "cell-1", 5000, true, "file-based", []string{"a.go"}, 0, 0, []string{"tests_pass"})
	if err != nil {
		t.Fatalf("RecordOutcome: %v", err)
	}

	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM events WHERE type = 'forge_outcome'").Scan(&count); err != nil {
		t.Fatalf("scan outcome count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 outcome event, got %d", count)
	}
}

func TestRecordOutcome_MissingBeadID(t *testing.T) {
	store := testStore(t)
	err := RecordOutcome(store, "", 0, false, "", nil, 0, 0, nil)
	if err == nil {
		t.Error("expected error for missing bead_id")
	}
}
