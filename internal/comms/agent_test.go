package comms

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

	agent, err := Init(store, "worker-1", "/project", "implement feature X")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if agent.Name != "worker-1" {
		t.Errorf("name = %q, want %q", agent.Name, "worker-1")
	}
	if agent.ProjectPath != "/project" {
		t.Errorf("project_path = %q, want %q", agent.ProjectPath, "/project")
	}
	if agent.TaskDescription != "implement feature X" {
		t.Errorf("task_description = %q, want %q", agent.TaskDescription, "implement feature X")
	}
	if agent.Status != "active" {
		t.Errorf("status = %q, want %q", agent.Status, "active")
	}
}

func TestInit_Upsert(t *testing.T) {
	store := testStore(t)

	// First init.
	agent1, err := Init(store, "worker-1", "/old", "old task")
	if err != nil {
		t.Fatalf("first Init: %v", err)
	}

	// Second init with same name -- should update.
	agent2, err := Init(store, "worker-1", "/new", "new task")
	if err != nil {
		t.Fatalf("second Init: %v", err)
	}

	if agent2.ID != agent1.ID {
		t.Errorf("ID changed from %d to %d on upsert", agent1.ID, agent2.ID)
	}
	if agent2.ProjectPath != "/new" {
		t.Errorf("project_path = %q, want %q", agent2.ProjectPath, "/new")
	}
	if agent2.TaskDescription != "new task" {
		t.Errorf("task_description = %q, want %q", agent2.TaskDescription, "new task")
	}
}

func TestInit_EmptyName(t *testing.T) {
	store := testStore(t)

	_, err := Init(store, "", "/project", "task")
	if err == nil {
		t.Error("expected error for empty agent name")
	}
}
