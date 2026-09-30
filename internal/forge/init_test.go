package forge

import (
	"testing"

	"github.com/unbound-force/replicator/internal/db"
)

func testStore(t *testing.T) *db.Store {
	t.Helper()
	store, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	return store
}

func TestInit(t *testing.T) {
	store := testStore(t)

	result, err := Init(store, "/tmp/project", "worktree")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if result["status"] != "initialized" {
		t.Errorf("status = %v, want %q", result["status"], "initialized")
	}
	if result["project_path"] != "/tmp/project" {
		t.Errorf("project_path = %v, want %q", result["project_path"], "/tmp/project")
	}
	if result["isolation"] != "worktree" {
		t.Errorf("isolation = %v, want %q", result["isolation"], "worktree")
	}
}

func TestInit_DefaultIsolation(t *testing.T) {
	store := testStore(t)

	result, err := Init(store, "/tmp/project", "")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if result["isolation"] != "reservation" {
		t.Errorf("isolation = %v, want %q", result["isolation"], "reservation")
	}
}

func TestInit_MissingProjectPath(t *testing.T) {
	store := testStore(t)

	_, err := Init(store, "", "worktree")
	if err == nil {
		t.Error("expected error for empty project_path")
	}
}

func TestInit_RecordsEvent(t *testing.T) {
	store := testStore(t)

	if _, err := Init(store, "/tmp/project", "worktree"); err != nil {
		t.Fatalf("Init: %v", err)
	}

	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM events WHERE type = 'forge_init'").Scan(&count); err != nil {
		t.Fatalf("scan event count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 forge_init event, got %d", count)
	}
}
