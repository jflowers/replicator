package comms

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestReserve(t *testing.T) {
	store := testStore(t)

	reservations, err := Reserve(store, "worker-1", []string{"foo.go", "bar.go"}, true, "implementing feature", 300)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if len(reservations) != 2 {
		t.Fatalf("expected 2 reservations, got %d", len(reservations))
	}
	if reservations[0].Path != "foo.go" {
		t.Errorf("path = %q, want %q", reservations[0].Path, "foo.go")
	}
	if reservations[0].AgentName != "worker-1" {
		t.Errorf("agent_name = %q, want %q", reservations[0].AgentName, "worker-1")
	}
	if !reservations[0].Exclusive {
		t.Error("exclusive should be true")
	}
	if reservations[0].TTLSeconds != 300 {
		t.Errorf("ttl_seconds = %d, want 300", reservations[0].TTLSeconds)
	}
}

func TestReserve_DefaultTTL(t *testing.T) {
	store := testStore(t)

	reservations, err := Reserve(store, "worker-1", []string{"foo.go"}, true, "test", 0)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if reservations[0].TTLSeconds != 300 {
		t.Errorf("default ttl = %d, want 300", reservations[0].TTLSeconds)
	}
}

func TestReserve_ExclusiveConflict(t *testing.T) {
	store := testStore(t)

	// Worker 1 reserves foo.go exclusively.
	_, err := Reserve(store, "worker-1", []string{"foo.go"}, true, "first", 300)
	if err != nil {
		t.Fatalf("first Reserve: %v", err)
	}

	// Worker 2 tries to reserve foo.go -- should fail.
	_, err = Reserve(store, "worker-2", []string{"foo.go"}, true, "second", 300)
	if err == nil {
		t.Error("expected conflict error for exclusive reservation")
	}
}

func TestReserve_SameAgentNoConflict(t *testing.T) {
	store := testStore(t)

	// Same agent can reserve the same path again.
	_, err := Reserve(store, "worker-1", []string{"foo.go"}, true, "first", 300)
	if err != nil {
		t.Fatalf("first Reserve: %v", err)
	}

	_, err = Reserve(store, "worker-1", []string{"foo.go"}, true, "second", 300)
	if err != nil {
		t.Fatalf("same agent should not conflict: %v", err)
	}
}

func TestReserve_NonExclusiveNoConflict(t *testing.T) {
	store := testStore(t)

	// Non-exclusive reservation should not block others.
	_, err := Reserve(store, "worker-1", []string{"foo.go"}, false, "reading", 300)
	if err != nil {
		t.Fatalf("first Reserve: %v", err)
	}

	_, err = Reserve(store, "worker-2", []string{"foo.go"}, false, "also reading", 300)
	if err != nil {
		t.Fatalf("non-exclusive should not conflict: %v", err)
	}
}

func TestRelease_ByPath(t *testing.T) {
	store := testStore(t)

	if _, err := Reserve(store, "worker-1", []string{"foo.go"}, true, "test", 300); err != nil {
		t.Fatalf("Reserve: %v", err)
	}

	if err := Release(store, []string{"foo.go"}, nil); err != nil {
		t.Fatalf("Release: %v", err)
	}

	// Should be able to reserve again by another agent.
	_, err := Reserve(store, "worker-2", []string{"foo.go"}, true, "after release", 300)
	if err != nil {
		t.Fatalf("reserve after release: %v", err)
	}
}

func TestRelease_ByID(t *testing.T) {
	store := testStore(t)

	reservations, _ := Reserve(store, "worker-1", []string{"foo.go"}, true, "test", 300)

	if err := Release(store, nil, []int{reservations[0].ID}); err != nil {
		t.Fatalf("Release by ID: %v", err)
	}

	// Should be able to reserve again.
	_, err := Reserve(store, "worker-2", []string{"foo.go"}, true, "after release", 300)
	if err != nil {
		t.Fatalf("reserve after release by ID: %v", err)
	}
}

func TestRelease_NothingSpecified(t *testing.T) {
	store := testStore(t)

	err := Release(store, nil, nil)
	if err == nil {
		t.Error("expected error when nothing specified")
	}
}

func TestReleaseAll(t *testing.T) {
	store := testStore(t)

	if _, err := Reserve(store, "worker-1", []string{"a.go"}, true, "test", 300); err != nil {
		t.Fatalf("Reserve worker-1: %v", err)
	}
	if _, err := Reserve(store, "worker-2", []string{"b.go"}, true, "test", 300); err != nil {
		t.Fatalf("Reserve worker-2: %v", err)
	}

	if err := ReleaseAll(store, ""); err != nil {
		t.Fatalf("ReleaseAll: %v", err)
	}

	// Both should be available now.
	_, err := Reserve(store, "worker-3", []string{"a.go", "b.go"}, true, "after release all", 300)
	if err != nil {
		t.Fatalf("reserve after release all: %v", err)
	}
}

func TestReleaseAgent(t *testing.T) {
	store := testStore(t)

	if _, err := Reserve(store, "worker-1", []string{"a.go", "b.go"}, true, "test", 300); err != nil {
		t.Fatalf("Reserve worker-1: %v", err)
	}
	if _, err := Reserve(store, "worker-2", []string{"c.go"}, true, "test", 300); err != nil {
		t.Fatalf("Reserve worker-2: %v", err)
	}

	if err := ReleaseAgent(store, "worker-1"); err != nil {
		t.Fatalf("ReleaseAgent: %v", err)
	}

	// worker-1's paths should be available.
	_, err := Reserve(store, "worker-3", []string{"a.go"}, true, "after agent release", 300)
	if err != nil {
		t.Fatalf("reserve a.go after agent release: %v", err)
	}

	// worker-2's path should still be reserved.
	_, err = Reserve(store, "worker-3", []string{"c.go"}, true, "conflict", 300)
	if err == nil {
		t.Error("expected conflict -- worker-2's reservation should still exist")
	}
}

func TestRunReservationTransaction_FinalizesExactlyOnce(t *testing.T) {
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
			err := runReservationTransaction(
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
				t.Errorf("runReservationTransaction: %v", err)
			}
		})
	}
}

func TestReserve_RollsBackPartialInserts(t *testing.T) {
	store := testStore(t)
	if _, err := store.DB.Exec(`
		CREATE TRIGGER fail_second_reservation
		BEFORE INSERT ON reservations
		WHEN NEW.path = 'fail.go'
		BEGIN
			SELECT RAISE(ABORT, 'reservation sentinel');
		END`); err != nil {
		t.Fatalf("create failure trigger: %v", err)
	}

	reservations, err := Reserve(store, "worker-1", []string{"first.go", "fail.go"}, true, "test rollback", 300)
	if err == nil || !strings.Contains(err.Error(), "insert reservation") {
		t.Fatalf("Reserve error = %v, want contextual insert failure", err)
	}
	if reservations != nil {
		t.Errorf("reservations = %#v, want nil", reservations)
	}

	var count int
	if err := store.DB.QueryRow("SELECT COUNT(*) FROM reservations").Scan(&count); err != nil {
		t.Fatalf("count reservations: %v", err)
	}
	if count != 0 {
		t.Errorf("persisted reservations = %d, want 0 after rollback", count)
	}
}
