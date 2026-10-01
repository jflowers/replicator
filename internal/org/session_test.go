package org

import "testing"

func TestSessionStart_NoHistory(t *testing.T) {
	store := testStore(t)

	notes, err := SessionStart(store, "")
	if err != nil {
		t.Fatalf("SessionStart: %v", err)
	}
	if notes != "" {
		t.Errorf("expected empty handoff notes, got %q", notes)
	}
}

func TestSessionStart_WithActiveCellID(t *testing.T) {
	store := testStore(t)

	cell, _ := CreateCell(store, CreateCellInput{Title: "Active task"})
	notes, err := SessionStart(store, cell.ID)
	if err != nil {
		t.Fatalf("SessionStart: %v", err)
	}
	if notes != "" {
		t.Errorf("expected empty handoff notes, got %q", notes)
	}

	// Verify session was created.
	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count); err != nil {
		t.Fatalf("scan session count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 session, got %d", count)
	}
}

func TestSessionEnd(t *testing.T) {
	store := testStore(t)

	// Start a session first.
	_, err := SessionStart(store, "")
	if err != nil {
		t.Fatalf("SessionStart: %v", err)
	}

	// End it with handoff notes.
	if err := SessionEnd(store, "Left off at task 3"); err != nil {
		t.Fatalf("SessionEnd: %v", err)
	}

	// Verify ended_at and handoff_notes are set.
	var endedAt, handoff string
	err = store.DB.QueryRow(
		"SELECT ended_at, handoff_notes FROM sessions LIMIT 1",
	).Scan(&endedAt, &handoff)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if endedAt == "" {
		t.Error("ended_at should not be empty")
	}
	if handoff != "Left off at task 3" {
		t.Errorf("handoff_notes = %q, want %q", handoff, "Left off at task 3")
	}
}

func TestSessionEnd_NoActiveSession(t *testing.T) {
	store := testStore(t)

	err := SessionEnd(store, "notes")
	if err == nil {
		t.Error("expected error when no active session exists")
	}
}

func TestSessionHandoff(t *testing.T) {
	store := testStore(t)

	// Start and end a session with handoff notes.
	if _, err := SessionStart(store, ""); err != nil {
		t.Fatalf("SessionStart: %v", err)
	}
	if err := SessionEnd(store, "Continue with task 5"); err != nil {
		t.Fatalf("SessionEnd: %v", err)
	}

	// Start a new session -- should get previous handoff notes.
	notes, err := SessionStart(store, "")
	if err != nil {
		t.Fatalf("SessionStart: %v", err)
	}
	if notes != "Continue with task 5" {
		t.Errorf("handoff notes = %q, want %q", notes, "Continue with task 5")
	}
}

func TestSessionMultipleHandoffs(t *testing.T) {
	store := testStore(t)

	// Session 1.
	if _, err := SessionStart(store, ""); err != nil {
		t.Fatalf("SessionStart: %v", err)
	}
	if err := SessionEnd(store, "First handoff"); err != nil {
		t.Fatalf("SessionEnd: %v", err)
	}

	// Session 2.
	if _, err := SessionStart(store, ""); err != nil {
		t.Fatalf("SessionStart: %v", err)
	}
	if err := SessionEnd(store, "Second handoff"); err != nil {
		t.Fatalf("SessionEnd: %v", err)
	}

	// Session 3 should get the most recent handoff.
	notes, err := SessionStart(store, "")
	if err != nil {
		t.Fatalf("SessionStart: %v", err)
	}
	if notes != "Second handoff" {
		t.Errorf("handoff notes = %q, want %q", notes, "Second handoff")
	}
}
