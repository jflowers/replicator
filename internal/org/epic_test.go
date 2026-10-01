package org

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCreateEpic(t *testing.T) {
	store := testStore(t)

	epic, subtasks, err := CreateEpic(store, CreateEpicInput{
		EpicTitle:       "Build the thing",
		EpicDescription: "A big feature",
		Subtasks: []SubtaskInput{
			{Title: "Step 1", Priority: 2},
			{Title: "Step 2", Priority: 1, Files: []string{"foo.go", "bar.go"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateEpic: %v", err)
	}

	if epic.Title != "Build the thing" {
		t.Errorf("epic title = %q, want %q", epic.Title, "Build the thing")
	}
	if epic.Type != "epic" {
		t.Errorf("epic type = %q, want %q", epic.Type, "epic")
	}
	if epic.Status != "open" {
		t.Errorf("epic status = %q, want %q", epic.Status, "open")
	}

	if len(subtasks) != 2 {
		t.Fatalf("expected 2 subtasks, got %d", len(subtasks))
	}
	if subtasks[0].Title != "Step 1" {
		t.Errorf("subtask[0] title = %q, want %q", subtasks[0].Title, "Step 1")
	}
	if subtasks[0].Priority != 2 {
		t.Errorf("subtask[0] priority = %d, want %d", subtasks[0].Priority, 2)
	}
	if subtasks[0].ParentID == nil || *subtasks[0].ParentID != epic.ID {
		t.Errorf("subtask[0] parent_id = %v, want %q", subtasks[0].ParentID, epic.ID)
	}
	if subtasks[1].Title != "Step 2" {
		t.Errorf("subtask[1] title = %q, want %q", subtasks[1].Title, "Step 2")
	}
}

func TestCreateEpic_NoSubtasks(t *testing.T) {
	store := testStore(t)

	epic, subtasks, err := CreateEpic(store, CreateEpicInput{
		EpicTitle: "Empty epic",
	})
	if err != nil {
		t.Fatalf("CreateEpic: %v", err)
	}
	if epic.Title != "Empty epic" {
		t.Errorf("epic title = %q, want %q", epic.Title, "Empty epic")
	}
	if len(subtasks) != 0 {
		t.Errorf("expected 0 subtasks, got %d", len(subtasks))
	}
}

func TestCreateEpic_SubtasksQueryable(t *testing.T) {
	store := testStore(t)

	epic, _, err := CreateEpic(store, CreateEpicInput{
		EpicTitle: "Queryable epic",
		Subtasks: []SubtaskInput{
			{Title: "Sub A"},
			{Title: "Sub B"},
		},
	})
	if err != nil {
		t.Fatalf("CreateEpic: %v", err)
	}

	// Query all cells -- should find epic + 2 subtasks.
	cells, err := QueryCells(store, CellQuery{})
	if err != nil {
		t.Fatalf("QueryCells: %v", err)
	}
	if len(cells) != 3 {
		t.Errorf("expected 3 cells (1 epic + 2 subtasks), got %d", len(cells))
	}

	// Query by epic type.
	epics, err := QueryCells(store, CellQuery{Type: "epic"})
	if err != nil {
		t.Fatalf("QueryCells: %v", err)
	}
	if len(epics) != 1 {
		t.Errorf("expected 1 epic, got %d", len(epics))
	}
	if epics[0].ID != epic.ID {
		t.Errorf("epic ID = %q, want %q", epics[0].ID, epic.ID)
	}
}

func TestCreateEpic_DefaultPriority(t *testing.T) {
	store := testStore(t)

	_, subtasks, err := CreateEpic(store, CreateEpicInput{
		EpicTitle: "Priority test",
		Subtasks: []SubtaskInput{
			{Title: "No priority set"},
		},
	})
	if err != nil {
		t.Fatalf("CreateEpic: %v", err)
	}
	if subtasks[0].Priority != 1 {
		t.Errorf("default priority = %d, want 1", subtasks[0].Priority)
	}
}

func TestRunEpicTransaction_FinalizesExactlyOnce(t *testing.T) {
	errWork := errors.New("work sentinel")
	errCommit := errors.New("commit sentinel")
	errRollback := errors.New("rollback sentinel")

	for _, test := range []struct {
		name        string
		workErr     error
		commitErr   error
		rollbackErr error
		wantOrder   []string
		want        []error
	}{
		{name: "success", rollbackErr: sql.ErrTxDone, wantOrder: []string{"work", "commit", "rollback"}},
		{name: "work", workErr: errWork, wantOrder: []string{"work", "rollback"}, want: []error{errWork}},
		{name: "work and rollback", workErr: errWork, rollbackErr: errRollback, wantOrder: []string{"work", "rollback"}, want: []error{errWork, errRollback}},
		{name: "commit", commitErr: errCommit, wantOrder: []string{"work", "commit", "rollback"}, want: []error{errCommit}},
		{name: "commit and rollback", commitErr: errCommit, rollbackErr: errRollback, wantOrder: []string{"work", "commit", "rollback"}, want: []error{errCommit, errRollback}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var order []string
			err := runEpicTransaction(
				func() error { order = append(order, "work"); return test.workErr },
				func() error { order = append(order, "commit"); return test.commitErr },
				func() error { order = append(order, "rollback"); return test.rollbackErr },
			)
			if !reflect.DeepEqual(order, test.wantOrder) {
				t.Errorf("order = %v, want %v", order, test.wantOrder)
			}
			for _, want := range test.want {
				if !errors.Is(err, want) {
					t.Errorf("error %v does not preserve %v", err, want)
				}
			}
			if len(test.want) == 0 && err != nil {
				t.Errorf("runEpicTransaction: %v", err)
			}
		})
	}
}

func TestCreateEpic_RollsBackPartialInserts(t *testing.T) {
	store := testStore(t)
	if _, err := store.DB.Exec(`
		CREATE TRIGGER fail_second_subtask
		BEFORE INSERT ON beads
		WHEN NEW.title = 'Fail subtask'
		BEGIN
			SELECT RAISE(ABORT, 'subtask sentinel');
		END`); err != nil {
		t.Fatalf("create failure trigger: %v", err)
	}

	epic, subtasks, err := CreateEpic(store, CreateEpicInput{
		EpicTitle: "Rollback epic",
		Subtasks: []SubtaskInput{
			{Title: "First subtask"},
			{Title: "Fail subtask"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "insert subtask") {
		t.Fatalf("CreateEpic error = %v, want contextual insert failure", err)
	}
	if epic != nil || subtasks != nil {
		t.Errorf("CreateEpic result = %#v, %#v; want nil results", epic, subtasks)
	}

	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM beads").Scan(&count); err != nil {
		t.Fatalf("count cells: %v", err)
	}
	if count != 0 {
		t.Errorf("persisted cells = %d, want 0 after rollback", count)
	}
}
