package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/tools/org"
	"github.com/unbound-force/replicator/internal/tools/registry"
)

// testLogEntry records a single log call for test assertions.
type testLogEntry struct {
	Level   string
	Msg     any
	Keyvals []any
}

// testLogger implements the Logger interface for tests.
// It captures all log calls so tests can assert on them.
type testLogger struct {
	mu      sync.Mutex
	entries []testLogEntry
}

func (l *testLogger) Info(msg any, keyvals ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, testLogEntry{Level: "info", Msg: msg, Keyvals: keyvals})
}

func (l *testLogger) Warn(msg any, keyvals ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, testLogEntry{Level: "warn", Msg: msg, Keyvals: keyvals})
}

func (l *testLogger) Entries() []testLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	cp := make([]testLogEntry, len(l.entries))
	copy(cp, l.entries)
	return cp
}

func testServer(t *testing.T) (*Server, *db.Store, *testLogger) {
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

	reg := registry.New()
	org.Register(reg, store)
	logger := &testLogger{}
	return NewServer(reg, "test", logger), store, logger
}

func call(t *testing.T, s *Server, method string, params any) json.RawMessage {
	t.Helper()
	paramsJSON, _ := json.Marshal(params)
	req := `{"jsonrpc":"2.0","id":1,"method":"` + method + `","params":` + string(paramsJSON) + "}\n"

	var buf bytes.Buffer
	err := s.Serve(strings.NewReader(req), &buf)
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}

	var resp jsonrpcResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nraw: %s", err, buf.String())
	}
	if resp.Error != nil {
		t.Fatalf("JSON-RPC error: %d %s", resp.Error.Code, resp.Error.Message)
	}

	result, _ := json.Marshal(resp.Result)
	return result
}

func TestToolsList(t *testing.T) {
	s, _, _ := testServer(t)
	result := call(t, s, "tools/list", nil)

	var list toolsListResult
	if err := json.Unmarshal(result, &list); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(list.Tools) != 11 {
		t.Errorf("expected 11 tools, got %d", len(list.Tools))
	}

	names := make(map[string]bool)
	for _, tool := range list.Tools {
		names[tool.Name] = true
	}
	for _, expected := range []string{
		"org_cells", "org_create", "org_close", "org_update",
		"org_create_epic", "org_query", "org_start", "org_ready",
		"org_sync", "org_session_start", "org_session_end",
	} {
		if !names[expected] {
			t.Errorf("missing tool: %s", expected)
		}
	}
}

func TestToolsCall_OrgCells_Empty(t *testing.T) {
	s, _, _ := testServer(t)
	result := call(t, s, "tools/call", toolsCallParams{
		Name:      "org_cells",
		Arguments: json.RawMessage(`{}`),
	})

	var callResult toolsCallResult
	if err := json.Unmarshal(result, &callResult); err != nil {
		t.Fatalf("unmarshal call result: %v", err)
	}

	if len(callResult.Content) == 0 {
		t.Fatal("expected content in response")
	}
	if callResult.Content[0].Text != "[]" {
		t.Errorf("expected empty array, got %s", callResult.Content[0].Text)
	}
}

func TestToolsCall_OrgCreate(t *testing.T) {
	s, _, _ := testServer(t)
	result := call(t, s, "tools/call", toolsCallParams{
		Name:      "org_create",
		Arguments: json.RawMessage(`{"title": "Test cell", "type": "bug"}`),
	})

	var callResult toolsCallResult
	if err := json.Unmarshal(result, &callResult); err != nil {
		t.Fatalf("unmarshal call result: %v", err)
	}

	if len(callResult.Content) == 0 {
		t.Fatal("expected content")
	}

	var cell map[string]any
	if err := json.Unmarshal([]byte(callResult.Content[0].Text), &cell); err != nil {
		t.Fatalf("unmarshal cell: %v", err)
	}

	if cell["title"] != "Test cell" {
		t.Errorf("title = %v, want %q", cell["title"], "Test cell")
	}
	if cell["type"] != "bug" {
		t.Errorf("type = %v, want %q", cell["type"], "bug")
	}
}

func TestToolsCall_CreateThenQuery(t *testing.T) {
	s, _, _ := testServer(t)

	// Create a cell.
	call(t, s, "tools/call", toolsCallParams{
		Name:      "org_create",
		Arguments: json.RawMessage(`{"title": "My task"}`),
	})

	// Query cells.
	result := call(t, s, "tools/call", toolsCallParams{
		Name:      "org_cells",
		Arguments: json.RawMessage(`{}`),
	})

	var callResult toolsCallResult
	if err := json.Unmarshal(result, &callResult); err != nil {
		t.Fatalf("unmarshal call result: %v", err)
	}

	var cells []map[string]any
	if err := json.Unmarshal([]byte(callResult.Content[0].Text), &cells); err != nil {
		t.Fatalf("unmarshal cells: %v", err)
	}

	if len(cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(cells))
	}
	if cells[0]["title"] != "My task" {
		t.Errorf("title = %v, want %q", cells[0]["title"], "My task")
	}
}

func TestToolsCall_UnknownTool(t *testing.T) {
	s, _, _ := testServer(t)

	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"nonexistent","arguments":{}}}` + "\n"
	var buf bytes.Buffer
	if err := s.Serve(strings.NewReader(req), &buf); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	var resp jsonrpcResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestInitialize(t *testing.T) {
	s, _, _ := testServer(t)
	result := call(t, s, "initialize", map[string]any{})

	var initResult map[string]any
	if err := json.Unmarshal(result, &initResult); err != nil {
		t.Fatalf("unmarshal initialize result: %v", err)
	}

	if initResult["protocolVersion"] != "2024-11-05" {
		t.Errorf("protocolVersion = %v", initResult["protocolVersion"])
	}
}

func TestToolsCall_LogsToolName(t *testing.T) {
	s, _, logger := testServer(t)

	call(t, s, "tools/call", toolsCallParams{
		Name:      "org_cells",
		Arguments: json.RawMessage(`{}`),
	})

	entries := logger.Entries()
	if len(entries) == 0 {
		t.Fatal("expected at least one log entry after tool call")
	}

	entry := entries[0]
	if entry.Level != "info" {
		t.Errorf("log level = %q, want %q", entry.Level, "info")
	}
	if entry.Msg != "tool call" {
		t.Errorf("log msg = %v, want %q", entry.Msg, "tool call")
	}

	// Verify keyvals contain "tool" and "duration".
	kvMap := keyvalMap(entry.Keyvals)
	if kvMap["tool"] != "org_cells" {
		t.Errorf("log tool = %v, want %q", kvMap["tool"], "org_cells")
	}
	if _, ok := kvMap["duration"]; !ok {
		t.Error("log entry missing 'duration' key")
	}
	if kvMap["success"] != true {
		t.Errorf("log success = %v, want true", kvMap["success"])
	}
}

func TestToolsCall_LogsMultipleCalls(t *testing.T) {
	s, _, logger := testServer(t)

	// Two tool calls should produce two log entries.
	call(t, s, "tools/call", toolsCallParams{
		Name:      "org_cells",
		Arguments: json.RawMessage(`{}`),
	})
	call(t, s, "tools/call", toolsCallParams{
		Name:      "org_create",
		Arguments: json.RawMessage(`{"title":"logged"}`),
	})

	entries := logger.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(entries))
	}

	kv0 := keyvalMap(entries[0].Keyvals)
	kv1 := keyvalMap(entries[1].Keyvals)
	if kv0["tool"] != "org_cells" {
		t.Errorf("first call tool = %v, want org_cells", kv0["tool"])
	}
	if kv1["tool"] != "org_create" {
		t.Errorf("second call tool = %v, want org_create", kv1["tool"])
	}
}

func TestNewServer_NilLogger(t *testing.T) {
	// A nil logger must not panic during tool calls.
	store, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	reg := registry.New()
	org.Register(reg, store)
	s := NewServer(reg, "test", nil)

	// Should not panic.
	call(t, s, "tools/call", toolsCallParams{
		Name:      "org_cells",
		Arguments: json.RawMessage(`{}`),
	})
}

// keyvalMap converts a flat keyval slice (key, value, key, value, ...)
// into a map for easier test assertions.
func keyvalMap(kvs []any) map[string]any {
	m := make(map[string]any)
	for i := 0; i+1 < len(kvs); i += 2 {
		if k, ok := kvs[i].(string); ok {
			m[k] = kvs[i+1]
		}
	}
	return m
}
