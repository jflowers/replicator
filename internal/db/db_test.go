package db

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type failingCloser struct {
	err   error
	calls int
}

func (c *failingCloser) Close() error {
	c.calls++
	return c.err
}

func TestOpenMemory(t *testing.T) {
	store, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	// Verify tables exist.
	tables := []string{"events", "agents", "beads", "cell_events", "sessions", "messages", "reservations"}
	for _, table := range tables {
		var name string
		err := store.DB.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestOpenMemory_Idempotent(t *testing.T) {
	store, err := OpenMemory()
	if err != nil {
		t.Fatalf("first open: %v", err)
	}

	// Running migrate again should not fail.
	if err := store.migrate(); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestCloseOnOpenError_PreservesPrimaryAndCloseErrors(t *testing.T) {
	errPrimary := errors.New("primary sentinel")
	errClose := errors.New("close sentinel")

	for _, test := range []struct {
		name      string
		primary   error
		closeErr  error
		wantCalls int
		want      []error
		contexts  []string
	}{
		{name: "no primary", wantCalls: 0},
		{name: "ping", primary: fmt.Errorf("ping sqlite: %w", errPrimary), wantCalls: 1, want: []error{errPrimary}, contexts: []string{"ping sqlite"}},
		{name: "ping and close", primary: fmt.Errorf("ping sqlite: %w", errPrimary), closeErr: errClose, wantCalls: 1, want: []error{errPrimary, errClose}, contexts: []string{"ping sqlite", "close database after open failure"}},
		{name: "disk migration", primary: fmt.Errorf("migrate: %w", errPrimary), wantCalls: 1, want: []error{errPrimary}, contexts: []string{"migrate"}},
		{name: "disk migration and close", primary: fmt.Errorf("migrate: %w", errPrimary), closeErr: errClose, wantCalls: 1, want: []error{errPrimary, errClose}, contexts: []string{"migrate", "close database after open failure"}},
		{name: "memory migration", primary: fmt.Errorf("migrate memory database: %w", errPrimary), wantCalls: 1, want: []error{errPrimary}, contexts: []string{"migrate memory database"}},
		{name: "memory migration and close", primary: fmt.Errorf("migrate memory database: %w", errPrimary), closeErr: errClose, wantCalls: 1, want: []error{errPrimary, errClose}, contexts: []string{"migrate memory database", "close database after open failure"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			closer := &failingCloser{err: test.closeErr}
			err := closeOnOpenError(test.primary, closer)
			if closer.calls != test.wantCalls {
				t.Fatalf("close calls = %d, want %d", closer.calls, test.wantCalls)
			}
			for _, want := range test.want {
				if !errors.Is(err, want) {
					t.Errorf("error %v does not preserve %v", err, want)
				}
			}
			for _, context := range test.contexts {
				if !strings.Contains(err.Error(), context) {
					t.Errorf("error %q does not contain %q", err, context)
				}
			}
		})
	}
}
