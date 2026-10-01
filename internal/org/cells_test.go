package org

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

func TestCreateCell(t *testing.T) {
	store := testStore(t)

	cell, err := CreateCell(store, CreateCellInput{
		Title: "Fix the bug",
		Type:  "bug",
	})
	if err != nil {
		t.Fatalf("CreateCell: %v", err)
	}
	if cell.Title != "Fix the bug" {
		t.Errorf("title = %q, want %q", cell.Title, "Fix the bug")
	}
	if cell.Type != "bug" {
		t.Errorf("type = %q, want %q", cell.Type, "bug")
	}
	if cell.Status != "open" {
		t.Errorf("status = %q, want %q", cell.Status, "open")
	}
	if cell.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestCreateCell_Defaults(t *testing.T) {
	store := testStore(t)

	cell, err := CreateCell(store, CreateCellInput{Title: "Do something"})
	if err != nil {
		t.Fatalf("CreateCell: %v", err)
	}
	if cell.Type != "task" {
		t.Errorf("default type = %q, want %q", cell.Type, "task")
	}
	if cell.Priority != 1 {
		t.Errorf("default priority = %d, want %d", cell.Priority, 1)
	}
}

func TestQueryCells_Empty(t *testing.T) {
	store := testStore(t)

	cells, err := QueryCells(store, CellQuery{})
	if err != nil {
		t.Fatalf("QueryCells: %v", err)
	}
	if len(cells) != 0 {
		t.Errorf("expected 0 cells, got %d", len(cells))
	}
}

func TestQueryCells_ByStatus(t *testing.T) {
	store := testStore(t)

	if _, err := CreateCell(store, CreateCellInput{Title: "Open task"}); err != nil {
		t.Fatalf("CreateCell: %v", err)
	}
	cell2, _ := CreateCell(store, CreateCellInput{Title: "Done task"})
	if err := CloseCell(store, cell2.ID, "completed"); err != nil {
		t.Fatalf("CloseCell: %v", err)
	}

	open, err := QueryCells(store, CellQuery{Status: "open"})
	if err != nil {
		t.Fatalf("QueryCells: %v", err)
	}
	if len(open) != 1 {
		t.Errorf("expected 1 open cell, got %d", len(open))
	}

	closed, err := QueryCells(store, CellQuery{Status: "closed"})
	if err != nil {
		t.Fatalf("QueryCells: %v", err)
	}
	if len(closed) != 1 {
		t.Errorf("expected 1 closed cell, got %d", len(closed))
	}
}

func TestQueryCells_ByType(t *testing.T) {
	store := testStore(t)

	if _, err := CreateCell(store, CreateCellInput{Title: "Bug", Type: "bug"}); err != nil {
		t.Fatalf("CreateCell: %v", err)
	}
	if _, err := CreateCell(store, CreateCellInput{Title: "Feature", Type: "feature"}); err != nil {
		t.Fatalf("CreateCell: %v", err)
	}

	bugs, err := QueryCells(store, CellQuery{Type: "bug"})
	if err != nil {
		t.Fatalf("QueryCells: %v", err)
	}
	if len(bugs) != 1 {
		t.Errorf("expected 1 bug, got %d", len(bugs))
	}
}

func TestCloseCell(t *testing.T) {
	store := testStore(t)

	cell, _ := CreateCell(store, CreateCellInput{Title: "Close me"})
	err := CloseCell(store, cell.ID, "done")
	if err != nil {
		t.Fatalf("CloseCell: %v", err)
	}

	cells, _ := QueryCells(store, CellQuery{Status: "closed"})
	if len(cells) != 1 {
		t.Fatalf("expected 1 closed cell, got %d", len(cells))
	}
	if cells[0].CloseReason == nil || *cells[0].CloseReason != "done" {
		t.Errorf("close_reason = %v, want %q", cells[0].CloseReason, "done")
	}
}

func TestCloseCell_NotFound(t *testing.T) {
	store := testStore(t)
	err := CloseCell(store, "nonexistent", "reason")
	if err == nil {
		t.Error("expected error for nonexistent cell")
	}
}

func TestUpdateCell(t *testing.T) {
	store := testStore(t)

	cell, _ := CreateCell(store, CreateCellInput{Title: "Update me"})
	status := "in_progress"
	desc := "Working on it"
	err := UpdateCell(store, cell.ID, &status, &desc, nil)
	if err != nil {
		t.Fatalf("UpdateCell: %v", err)
	}

	cells, _ := QueryCells(store, CellQuery{ID: cell.ID})
	if len(cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(cells))
	}
	if cells[0].Status != "in_progress" {
		t.Errorf("status = %q, want %q", cells[0].Status, "in_progress")
	}
	if cells[0].Description != "Working on it" {
		t.Errorf("description = %q, want %q", cells[0].Description, "Working on it")
	}
}
