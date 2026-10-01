package org

import "testing"

func TestStartCell(t *testing.T) {
	store := testStore(t)

	cell, _ := CreateCell(store, CreateCellInput{Title: "Start me"})
	if err := StartCell(store, cell.ID); err != nil {
		t.Fatalf("StartCell: %v", err)
	}

	cells, _ := QueryCells(store, CellQuery{ID: cell.ID})
	if len(cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(cells))
	}
	if cells[0].Status != "in_progress" {
		t.Errorf("status = %q, want %q", cells[0].Status, "in_progress")
	}
}

func TestStartCell_NotFound(t *testing.T) {
	store := testStore(t)
	err := StartCell(store, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent cell")
	}
}

func TestReadyCell_NoParent(t *testing.T) {
	store := testStore(t)

	// Create two open cells with different priorities.
	if _, err := CreateCell(store, CreateCellInput{Title: "Low priority", Priority: 1}); err != nil {
		t.Fatalf("CreateCell low priority: %v", err)
	}
	if _, err := CreateCell(store, CreateCellInput{Title: "High priority", Priority: 3}); err != nil {
		t.Fatalf("CreateCell high priority: %v", err)
	}

	cell, err := ReadyCell(store)
	if err != nil {
		t.Fatalf("ReadyCell: %v", err)
	}
	if cell == nil {
		t.Fatal("expected a cell, got nil")
	}
	if cell.Title != "High priority" {
		t.Errorf("title = %q, want %q (highest priority)", cell.Title, "High priority")
	}
}

func TestReadyCell_Empty(t *testing.T) {
	store := testStore(t)

	cell, err := ReadyCell(store)
	if err != nil {
		t.Fatalf("ReadyCell: %v", err)
	}
	if cell != nil {
		t.Errorf("expected nil, got %+v", cell)
	}
}

func TestReadyCell_BlockedByParent(t *testing.T) {
	store := testStore(t)

	// Create an epic (open) with a subtask.
	epic, _, err := CreateEpic(store, CreateEpicInput{
		EpicTitle: "Blocking epic",
		Subtasks:  []SubtaskInput{{Title: "Blocked subtask", Priority: 3}},
	})
	if err != nil {
		t.Fatalf("CreateEpic: %v", err)
	}

	// The subtask should be blocked because its parent epic is open.
	// Create a standalone cell with lower priority.
	if _, err := CreateCell(store, CreateCellInput{Title: "Standalone", Priority: 1}); err != nil {
		t.Fatalf("CreateCell standalone: %v", err)
	}

	cell, err := ReadyCell(store)
	if err != nil {
		t.Fatalf("ReadyCell: %v", err)
	}
	if cell == nil {
		t.Fatal("expected a cell, got nil")
	}
	// Should return the standalone cell, not the blocked subtask.
	if cell.Title != "Standalone" {
		t.Errorf("title = %q, want %q", cell.Title, "Standalone")
	}

	// Close the epic -- now the subtask should become ready.
	if err := CloseCell(store, epic.ID, "done"); err != nil {
		t.Fatalf("CloseCell epic: %v", err)
	}

	cell, err = ReadyCell(store)
	if err != nil {
		t.Fatalf("ReadyCell after close: %v", err)
	}
	if cell == nil {
		t.Fatal("expected a cell after closing epic, got nil")
	}
	// Subtask has priority 3, standalone has priority 1.
	if cell.Title != "Blocked subtask" {
		t.Errorf("title = %q, want %q", cell.Title, "Blocked subtask")
	}
}

func TestReadyCell_AllClosed(t *testing.T) {
	store := testStore(t)

	cell, _ := CreateCell(store, CreateCellInput{Title: "Close me"})
	if err := CloseCell(store, cell.ID, "done"); err != nil {
		t.Fatalf("CloseCell: %v", err)
	}

	ready, err := ReadyCell(store)
	if err != nil {
		t.Fatalf("ReadyCell: %v", err)
	}
	if ready != nil {
		t.Errorf("expected nil, got %+v", ready)
	}
}
