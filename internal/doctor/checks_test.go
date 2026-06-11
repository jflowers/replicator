package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/unbound-force/replicator/internal/config"
	"github.com/unbound-force/replicator/internal/db"
)

// mcpHandler returns an http.HandlerFunc that mimics the MCP Streamable HTTP
// transport. It validates POST method, Content-Type, Accept header, and
// JSON-RPC protocol fields. Responds with SSE-formatted JSON-RPC success.
func mcpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			http.Error(w, "wrong content type", http.StatusUnsupportedMediaType)
			return
		}

		accept := r.Header.Get("Accept")
		if !strings.Contains(accept, "application/json") || !strings.Contains(accept, "text/event-stream") {
			http.Error(w, "Accept must contain both 'application/json' and 'text/event-stream'", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if req["jsonrpc"] != "2.0" {
			http.Error(w, "invalid jsonrpc version", http.StatusBadRequest)
			return
		}
		if req["method"] == nil {
			http.Error(w, "missing method", http.StatusBadRequest)
			return
		}

		id, _ := req["id"].(float64)
		resp := map[string]any{
			"jsonrpc": "2.0",
			"result": map[string]any{
				"capabilities":    map[string]any{},
				"protocolVersion": "2025-03-26",
				"serverInfo":      map[string]any{"name": "dewey", "version": "test"},
			},
			"id": int(id),
		}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
			return
		}

		// Respond in SSE format, matching the MCP Streamable HTTP transport.
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", respJSON)
	}
}

func testStore(t *testing.T) *db.Store {
	t.Helper()
	store, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestRun_AllChecks(t *testing.T) {
	store := testStore(t)

	// Mock Dewey as healthy JSON-RPC endpoint.
	srv := httptest.NewServer(mcpHandler())
	defer srv.Close()

	cfg := &config.Config{
		DeweyURL: srv.URL,
	}

	results, err := Run(store, cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(results) != 4 {
		t.Fatalf("expected 4 checks, got %d", len(results))
	}

	// Verify check names.
	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}
	for _, expected := range []string{"git", "database", "dewey", "config_dir"} {
		if !names[expected] {
			t.Errorf("missing check: %s", expected)
		}
	}
}

func TestCheckGit(t *testing.T) {
	if testing.Short() {
		t.Skip("requires git")
	}

	result := checkGit()
	if result.Name != "git" {
		t.Errorf("name = %q, want %q", result.Name, "git")
	}
	if result.Status != "pass" {
		t.Errorf("status = %q, want %q (message: %s)", result.Status, "pass", result.Message)
	}
	if result.Duration <= 0 {
		t.Error("duration should be positive")
	}
}

func TestCheckDatabase_Healthy(t *testing.T) {
	store := testStore(t)

	result := checkDatabase(store)
	if result.Status != "pass" {
		t.Errorf("status = %q, want %q (message: %s)", result.Status, "pass", result.Message)
	}
}

func TestCheckDatabase_Closed(t *testing.T) {
	store := testStore(t)
	store.Close()

	result := checkDatabase(store)
	if result.Status != "fail" {
		t.Errorf("status = %q, want %q for closed database", result.Status, "fail")
	}
}

func TestCheckDewey_Healthy(t *testing.T) {
	srv := httptest.NewServer(mcpHandler())
	defer srv.Close()

	result := checkDewey(srv.URL)
	if result.Name != "dewey" {
		t.Errorf("name = %q, want %q", result.Name, "dewey")
	}
	if result.Status != "pass" {
		t.Errorf("status = %q, want %q (message: %s)", result.Status, "pass", result.Message)
	}
	if !strings.Contains(result.Message, "Dewey is reachable") {
		t.Errorf("message = %q, want it to contain %q", result.Message, "Dewey is reachable")
	}
	if result.Duration <= 0 {
		t.Error("duration should be positive")
	}
}

func TestCheckDewey_Unreachable(t *testing.T) {
	result := checkDewey("http://127.0.0.1:1")
	if result.Name != "dewey" {
		t.Errorf("name = %q, want %q", result.Name, "dewey")
	}
	if result.Status != "warn" {
		t.Errorf("status = %q, want %q for unreachable Dewey", result.Status, "warn")
	}
	if !strings.Contains(result.Message, "not reachable") {
		t.Errorf("message = %q, want it to contain %q", result.Message, "not reachable")
	}
}

func TestCheckDewey_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	result := checkDewey(srv.URL)
	if result.Name != "dewey" {
		t.Errorf("name = %q, want %q", result.Name, "dewey")
	}
	if result.Status != "warn" {
		t.Errorf("status = %q, want %q for HTTP 500", result.Status, "warn")
	}
	if !strings.Contains(result.Message, "not reachable") {
		t.Errorf("message = %q, want it to contain %q", result.Message, "not reachable")
	}
}

func TestCheckDewey_RPCError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32601,\"message\":\"method not found\"},\"id\":1}\n\n")
	}))
	defer srv.Close()

	result := checkDewey(srv.URL)
	if result.Name != "dewey" {
		t.Errorf("name = %q, want %q", result.Name, "dewey")
	}
	if result.Status != "warn" {
		t.Errorf("status = %q, want %q for JSON-RPC error", result.Status, "warn")
	}
	if !strings.Contains(result.Message, "not reachable") {
		t.Errorf("message = %q, want it to contain %q", result.Message, "not reachable")
	}
}

func TestCheckConfigDir(t *testing.T) {
	result := checkConfigDir()
	// The config dir may or may not exist in CI, but the check should not panic.
	if result.Name != "config_dir" {
		t.Errorf("name = %q, want %q", result.Name, "config_dir")
	}
	if result.Status != "pass" && result.Status != "fail" {
		t.Errorf("status = %q, want pass or fail", result.Status)
	}
	if result.Duration <= 0 {
		t.Error("duration should be positive")
	}
}

func TestCheckResult_StatusValues(t *testing.T) {
	// Verify that all results use valid status values.
	store := testStore(t)
	srv := httptest.NewServer(mcpHandler())
	defer srv.Close()

	cfg := &config.Config{DeweyURL: srv.URL}
	results, _ := Run(store, cfg)

	validStatuses := map[string]bool{"pass": true, "fail": true, "warn": true}
	for _, r := range results {
		if !validStatuses[r.Status] {
			t.Errorf("check %q has invalid status %q", r.Name, r.Status)
		}
	}
}
