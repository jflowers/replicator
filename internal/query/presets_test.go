package query

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/unbound-force/replicator/internal/db"
)

type failingQueryRows struct {
	scanErr    error
	closeErr   error
	nextCalls  int
	closeCalls int
}

func (r *failingQueryRows) Next() bool {
	r.nextCalls++
	return r.nextCalls == 1
}

func (r *failingQueryRows) Scan(...any) error { return r.scanErr }
func (r *failingQueryRows) Err() error        { return nil }
func (r *failingQueryRows) Close() error {
	r.closeCalls++
	return r.closeErr
}

type failingCountRow struct {
	err   error
	value int
	calls int
}

type queryFailingWriter struct{ err error }

func (w queryFailingWriter) Write([]byte) (int, error) { return 0, w.err }

func (r *failingCountRow) Scan(dest ...any) error {
	r.calls++
	if r.err != nil {
		return r.err
	}
	*(dest[0].(*int)) = r.value
	return nil
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

func TestListPresets(t *testing.T) {
	presets := ListPresets()
	if len(presets) != 4 {
		t.Fatalf("expected 4 presets, got %d", len(presets))
	}

	expected := map[string]bool{
		AgentActivity24h:    true,
		CellsByStatus:       true,
		ForgeCompletionRate: true,
		RecentEvents:        true,
	}
	for _, p := range presets {
		if !expected[p] {
			t.Errorf("unexpected preset: %q", p)
		}
	}
}

func TestRun_UnknownPreset(t *testing.T) {
	store := testStore(t)
	var buf bytes.Buffer

	err := Run(store, "nonexistent", &buf)
	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
	if !strings.Contains(err.Error(), "unknown preset") {
		t.Errorf("error = %q, want 'unknown preset'", err.Error())
	}
}

func TestRun_AgentActivity_Empty(t *testing.T) {
	store := testStore(t)
	var buf bytes.Buffer

	err := Run(store, AgentActivity24h, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "no activity") {
		t.Error("expected empty message")
	}
}

func TestRun_AgentActivity_WithData(t *testing.T) {
	store := testStore(t)

	// Insert events with agent_name in payload.
	for _, payload := range []string{`{"agent_name":"worker-1"}`, `{"agent_name":"worker-1"}`, `{"agent_name":"worker-2"}`} {
		if _, err := store.DB.Exec(`INSERT INTO events (type, payload, project_key) VALUES ('forge_init', ?, 'test')`, payload); err != nil {
			t.Fatalf("insert event: %v", err)
		}
	}

	var buf bytes.Buffer
	err := Run(store, AgentActivity24h, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "worker-1") {
		t.Error("expected worker-1 in output")
	}
}

func TestRun_CellsByStatus_Empty(t *testing.T) {
	store := testStore(t)
	var buf bytes.Buffer

	err := Run(store, CellsByStatus, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "no cells") {
		t.Error("expected empty message")
	}
}

func TestRun_CellsByStatus_WithData(t *testing.T) {
	store := testStore(t)

	for _, query := range []string{
		`INSERT INTO beads (id, title, status, type) VALUES ('c1', 'Task 1', 'open', 'task')`,
		`INSERT INTO beads (id, title, status, type) VALUES ('c2', 'Task 2', 'closed', 'task')`,
		`INSERT INTO beads (id, title, status, type) VALUES ('c3', 'Bug 1', 'open', 'bug')`,
	} {
		if _, err := store.DB.Exec(query); err != nil {
			t.Fatalf("insert cell: %v", err)
		}
	}

	var buf bytes.Buffer
	err := Run(store, CellsByStatus, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "open") {
		t.Error("expected 'open' in output")
	}
	if !strings.Contains(output, "closed") {
		t.Error("expected 'closed' in output")
	}
}

func TestRun_ForgeCompletionRate_Empty(t *testing.T) {
	store := testStore(t)
	var buf bytes.Buffer

	err := Run(store, ForgeCompletionRate, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Forge Completion Rate") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "N/A") {
		t.Error("expected N/A for empty data")
	}
}

func TestRun_ForgeCompletionRate_WithData(t *testing.T) {
	store := testStore(t)

	for _, eventType := range []string{"forge_init", "forge_progress", "forge_complete"} {
		if _, err := store.DB.Exec(`INSERT INTO events (type, payload, project_key) VALUES (?, '{}', 'test')`, eventType); err != nil {
			t.Fatalf("insert event: %v", err)
		}
	}

	var buf bytes.Buffer
	err := Run(store, ForgeCompletionRate, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Total forge events") {
		t.Error("expected total count")
	}
	if !strings.Contains(output, "Completed") {
		t.Error("expected completed count")
	}
}

func TestRun_RecentEvents_Empty(t *testing.T) {
	store := testStore(t)
	var buf bytes.Buffer

	err := Run(store, RecentEvents, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "no events") {
		t.Error("expected empty message")
	}
}

func TestRun_RecentEvents_WithData(t *testing.T) {
	store := testStore(t)

	if _, err := store.DB.Exec(`INSERT INTO events (type, payload, project_key) VALUES ('test_event', '{}', 'my-project')`); err != nil {
		t.Fatalf("insert event: %v", err)
	}

	var buf bytes.Buffer
	err := Run(store, RecentEvents, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test_event") {
		t.Error("expected test_event in output")
	}
	if !strings.Contains(output, "my-project") {
		t.Error("expected my-project in output")
	}
}

func TestRun_AllPresets(t *testing.T) {
	store := testStore(t)

	// Verify all presets run without error on an empty database.
	for _, preset := range ListPresets() {
		t.Run(preset, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Run(store, preset, &buf); err != nil {
				t.Fatalf("Run(%q): %v", preset, err)
			}
			if buf.Len() == 0 {
				t.Errorf("preset %q produced no output", preset)
			}
		})
	}
}

func TestRun_ReturnsWriterError(t *testing.T) {
	errWrite := errors.New("write sentinel")
	err := Run(testStore(t), AgentActivity24h, queryFailingWriter{err: errWrite})
	if !errors.Is(err, errWrite) {
		t.Fatalf("error %v does not preserve %v", err, errWrite)
	}
}

func TestReadTableRows_PreservesScanAndCloseErrors(t *testing.T) {
	errScan := errors.New("scan sentinel")
	errClose := errors.New("close sentinel")

	for _, test := range []struct {
		name     string
		scanErr  error
		closeErr error
		want     []error
	}{
		{name: "success"},
		{name: "scan", scanErr: errScan, want: []error{errScan}},
		{name: "close", closeErr: errClose, want: []error{errClose}},
		{name: "scan and close", scanErr: errScan, closeErr: errClose, want: []error{errScan, errClose}},
	} {
		t.Run(test.name, func(t *testing.T) {
			rows := &failingQueryRows{scanErr: test.scanErr, closeErr: test.closeErr}
			got, err := readTableRows(rows, func(rows queryRows) ([]string, error) {
				if err := rows.Scan(new(string)); err != nil {
					return nil, err
				}
				return []string{"row"}, nil
			})
			if rows.closeCalls != 1 {
				t.Errorf("close calls = %d, want 1", rows.closeCalls)
			}
			if len(test.want) == 0 && (err != nil || len(got) != 1 || got[0][0] != "row") {
				t.Fatalf("readTableRows = %v, %v; want one row and nil", got, err)
			}
			if len(test.want) > 0 && got != nil {
				t.Errorf("rows = %v, want nil on failure", got)
			}
			for _, want := range test.want {
				if !errors.Is(err, want) {
					t.Errorf("error %v does not preserve %v", err, want)
				}
			}
			if test.scanErr != nil && !strings.Contains(err.Error(), "scan rows") {
				t.Errorf("error %q missing scan context", err)
			}
			if test.closeErr != nil && !strings.Contains(err.Error(), "close rows") {
				t.Errorf("error %q missing close context", err)
			}
		})
	}
}

func TestScanForgeCounts_ReturnsEachScanError(t *testing.T) {
	errTotal := errors.New("total sentinel")
	errCompleted := errors.New("completed sentinel")

	first := &failingCountRow{err: errTotal}
	second := &failingCountRow{value: 2}
	_, _, err := scanForgeCounts(first, func() countRow { return second })
	if !errors.Is(err, errTotal) || first.calls != 1 || second.calls != 0 {
		t.Fatalf("first failure: err=%v calls=%d/%d", err, first.calls, second.calls)
	}

	first = &failingCountRow{value: 3}
	second = &failingCountRow{err: errCompleted}
	_, _, err = scanForgeCounts(first, func() countRow { return second })
	if !errors.Is(err, errCompleted) || first.calls != 1 || second.calls != 1 {
		t.Fatalf("second failure: err=%v calls=%d/%d", err, first.calls, second.calls)
	}

	first = &failingCountRow{value: 3}
	second = &failingCountRow{value: 2}
	total, completed, err := scanForgeCounts(first, func() countRow { return second })
	if err != nil || total != 3 || completed != 2 {
		t.Fatalf("success = %d/%d, %v; want 3/2, nil", total, completed, err)
	}
}
