package stats

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/unbound-force/replicator/internal/db"
)

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestCombineRowsCloseError_PreservesBothErrors(t *testing.T) {
	errScan := errors.New("scan sentinel")
	errClose := errors.New("close sentinel")
	err := combineRowsCloseError(errScan, errClose)
	if !errors.Is(err, errScan) || !errors.Is(err, errClose) {
		t.Fatalf("combined error %v does not preserve both sentinels", err)
	}
}

func TestCombineRowsCloseError_NilCloseReturnsPrimary(t *testing.T) {
	errPrimary := errors.New("primary sentinel")
	if err := combineRowsCloseError(errPrimary, nil); !errors.Is(err, errPrimary) {
		t.Fatalf("combineRowsCloseError() = %v, want primary error", err)
	}
}

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

func TestRun_EmptyDatabase(t *testing.T) {
	store := testStore(t)
	var buf bytes.Buffer

	err := Run(store, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Replicator Stats") {
		t.Error("expected header in output")
	}
	if !strings.Contains(output, "(no events)") {
		t.Error("expected '(no events)' for empty database")
	}
	if !strings.Contains(output, "(no cells)") {
		t.Error("expected '(no cells)' for empty database")
	}
	if !strings.Contains(output, "Recent Activity (24h): 0 events") {
		t.Error("expected 0 recent events")
	}
}

func TestRun_WithEvents(t *testing.T) {
	store := testStore(t)

	// Insert some events.
	for i := 0; i < 3; i++ {
		if _, err := store.DB.Exec(`INSERT INTO events (type, payload, project_key) VALUES (?, '{}', 'test')`, "forge_init"); err != nil {
			t.Fatalf("insert event: %v", err)
		}
	}
	if _, err := store.DB.Exec(`INSERT INTO events (type, payload, project_key) VALUES (?, '{}', 'test')`, "forge_complete"); err != nil {
		t.Fatalf("insert event: %v", err)
	}

	var buf bytes.Buffer
	err := Run(store, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "forge_init") {
		t.Error("expected forge_init in output")
	}
	if !strings.Contains(output, "forge_complete") {
		t.Error("expected forge_complete in output")
	}
}

func TestRun_WithCells(t *testing.T) {
	store := testStore(t)

	// Insert cells with different statuses.
	if _, err := store.DB.Exec(`INSERT INTO beads (id, title, status) VALUES ('c1', 'Task 1', 'open')`); err != nil {
		t.Fatalf("insert cell: %v", err)
	}
	for _, query := range []string{
		`INSERT INTO beads (id, title, status) VALUES ('c2', 'Task 2', 'open')`,
		`INSERT INTO beads (id, title, status) VALUES ('c3', 'Task 3', 'closed')`,
	} {
		if _, err := store.DB.Exec(query); err != nil {
			t.Fatalf("insert cell: %v", err)
		}
	}

	var buf bytes.Buffer
	err := Run(store, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "3 total") {
		t.Errorf("expected '3 total' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "open") {
		t.Error("expected 'open' status in output")
	}
	if !strings.Contains(output, "closed") {
		t.Error("expected 'closed' status in output")
	}
}

func TestRun_RecentActivity(t *testing.T) {
	store := testStore(t)

	// Insert events with current timestamp (default).
	for i := 0; i < 5; i++ {
		if _, err := store.DB.Exec(`INSERT INTO events (type, payload, project_key) VALUES ('test', '{}', 'test')`); err != nil {
			t.Fatalf("insert event: %v", err)
		}
	}

	var buf bytes.Buffer
	err := Run(store, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Recent Activity (24h): 5 events") {
		t.Errorf("expected 5 recent events, got:\n%s", output)
	}
}

func TestRun_WritesToWriter(t *testing.T) {
	store := testStore(t)
	var buf bytes.Buffer

	err := Run(store, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestRun_WriterFailure(t *testing.T) {
	store := testStore(t)
	errWrite := errors.New("write sentinel")

	err := Run(store, failingWriter{err: errWrite})
	if !errors.Is(err, errWrite) {
		t.Fatalf("Run() error = %v, want write sentinel", err)
	}
}

func TestQueries_ClosedDatabase(t *testing.T) {
	queries := []struct {
		name string
		run  func(*db.Store) error
	}{
		{
			name: "event counts",
			run: func(store *db.Store) error {
				_, err := queryEventCounts(store)
				return err
			},
		},
		{
			name: "recent events",
			run: func(store *db.Store) error {
				_, err := queryRecentEvents(store)
				return err
			},
		},
		{
			name: "cell counts",
			run: func(store *db.Store) error {
				_, err := queryCellCounts(store)
				return err
			},
		},
	}

	for _, query := range queries {
		t.Run(query.name, func(t *testing.T) {
			store, err := db.OpenMemory()
			if err != nil {
				t.Fatalf("OpenMemory: %v", err)
			}
			if err := store.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}

			if err := query.run(store); err == nil {
				t.Fatal("query on closed database returned nil error")
			}
		})
	}
}
