package mcpclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mcpHandler is a stateful mock MCP server handler that tracks initialization
// state, validates MCP protocol compliance, and returns SSE-formatted responses.
type mcpHandler struct {
	t *testing.T

	mu            sync.Mutex
	initialized   bool
	sessionID     string
	initCount     atomic.Int64
	toolCallCount atomic.Int64

	// overrides allow tests to customize behavior.
	initStatus     int                                          // HTTP status for initialize (0 = 200)
	initResponse   string                                       // custom init response body
	noSessionID    bool                                          // omit Mcp-Session-Id header
	toolHandler    func(name string, args json.RawMessage) any   // custom tool result
	toolStatus     int                                           // HTTP status for tools/call (0 = 200)
	toolResponse   string                                        // custom tools/call response body
	plainJSON      bool                                          // respond with application/json instead of SSE
	emptyBody      bool                                          // respond with empty body
	malformedSSE   bool                                          // respond with malformed SSE
	noDataLine     bool                                          // respond with SSE but no data: line
	emptyContent   bool                                          // respond with empty content array
	rejectCount    int                                            // reject this many tools/call with 400 before succeeding
	rejectedSoFar  int
}

func (h *mcpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Validate common headers.
	if r.Header.Get("Content-Type") != "application/json" {
		h.t.Errorf("missing Content-Type: application/json")
	}
	accept := r.Header.Get("Accept")
	if !strings.Contains(accept, "application/json") || !strings.Contains(accept, "text/event-stream") {
		h.t.Errorf("Accept header missing required types: %q", accept)
	}

	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

	method, _ := req["method"].(string)

	switch method {
	case "initialize":
		h.handleInitialize(w, r, req)
	case "tools/call":
		h.handleToolsCall(w, r, req)
	default:
		http.Error(w, fmt.Sprintf("unknown method: %s", method), http.StatusBadRequest)
	}
}

func (h *mcpHandler) handleInitialize(w http.ResponseWriter, r *http.Request, req map[string]any) {
	h.initCount.Add(1)

	if h.initStatus != 0 {
		http.Error(w, "init error", h.initStatus)
		return
	}

	h.mu.Lock()
	h.initialized = true
	h.sessionID = "test-session-123"
	h.mu.Unlock()

	if !h.noSessionID {
		w.Header().Set("Mcp-Session-Id", "test-session-123")
	}

	result := map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":   map[string]any{},
		"serverInfo": map[string]any{
			"name":    "test-server",
			"version": "1.0.0",
		},
	}

	rpcResp := map[string]any{
		"jsonrpc": "2.0",
		"id":      req["id"],
		"result":  result,
	}

	if h.plainJSON {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rpcResp)
		return
	}

	data, _ := json.Marshal(rpcResp)
	w.Header().Set("Content-Type", "text/event-stream")
	fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
}

func (h *mcpHandler) handleToolsCall(w http.ResponseWriter, r *http.Request, req map[string]any) {
	h.toolCallCount.Add(1)

	// Validate session ID header matches issued session (only when the
	// handler actually sent a session ID in the initialize response).
	h.mu.Lock()
	expectedSID := h.sessionID
	issuedSID := !h.noSessionID && expectedSID != ""
	h.mu.Unlock()
	if issuedSID {
		gotSID := r.Header.Get("Mcp-Session-Id")
		if gotSID != expectedSID {
			h.t.Errorf("Mcp-Session-Id = %q, want %q", gotSID, expectedSID)
		}
	}

	// Session rejection simulation.
	h.mu.Lock()
	if h.rejectCount > 0 && h.rejectedSoFar < h.rejectCount {
		h.rejectedSoFar++
		h.initialized = false
		h.mu.Unlock()
		http.Error(w, "session expired", http.StatusBadRequest)
		return
	}
	h.mu.Unlock()

	if h.toolStatus != 0 {
		http.Error(w, "tool error", h.toolStatus)
		return
	}

	if h.emptyBody {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		return
	}

	if h.malformedSSE {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {not-valid-json}\n\n")
		return
	}

	if h.noDataLine {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: message\n\n")
		return
	}

	// Extract tool name and arguments from params.
	params, _ := req["params"].(map[string]any)
	toolName, _ := params["name"].(string)
	argsRaw, _ := json.Marshal(params["arguments"])

	var toolResult any
	if h.toolHandler != nil {
		toolResult = h.toolHandler(toolName, argsRaw)
	} else {
		toolResult = map[string]string{"status": "ok"}
	}

	if h.emptyContent {
		// Return valid JSON-RPC but empty content array.
		rpcResp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result": map[string]any{
				"content": []any{},
			},
		}
		data, _ := json.Marshal(rpcResp)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
		return
	}

	resultJSON, _ := json.Marshal(toolResult)
	rpcResp := map[string]any{
		"jsonrpc": "2.0",
		"id":      req["id"],
		"result": map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": string(resultJSON),
				},
			},
		},
	}

	data, _ := json.Marshal(rpcResp)

	if h.plainJSON {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
}

// newMCPServer creates a test server with the stateful MCP handler.
func newMCPServer(t *testing.T) (*httptest.Server, *mcpHandler) {
	t.Helper()
	h := &mcpHandler{t: t}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv, h
}

// --- Happy Path Tests ---

func TestCall_SuccessfulInitializeAndToolsCall(t *testing.T) {
	srv, h := newMCPServer(t)

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	result, err := client.Call("dewey_health", map[string]any{})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if parsed["status"] != "ok" {
		t.Errorf("status = %q, want %q", parsed["status"], "ok")
	}

	if h.initCount.Load() != 1 {
		t.Errorf("init count = %d, want 1", h.initCount.Load())
	}
	if h.toolCallCount.Load() != 1 {
		t.Errorf("tool call count = %d, want 1", h.toolCallCount.Load())
	}
}

func TestCall_SessionReuse(t *testing.T) {
	srv, h := newMCPServer(t)

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})

	// First call triggers init.
	if _, err := client.Call("dewey_health", map[string]any{}); err != nil {
		t.Fatalf("first Call: %v", err)
	}

	// Second call should reuse session.
	if _, err := client.Call("store_learning", map[string]any{"info": "test"}); err != nil {
		t.Fatalf("second Call: %v", err)
	}

	if h.initCount.Load() != 1 {
		t.Errorf("init count = %d, want 1 (should not re-init)", h.initCount.Load())
	}
	if h.toolCallCount.Load() != 2 {
		t.Errorf("tool call count = %d, want 2", h.toolCallCount.Load())
	}
}

// --- Error Path Tests ---

func TestCall_InitializeHTTPError(t *testing.T) {
	srv, h := newMCPServer(t)
	h.initStatus = http.StatusInternalServerError

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("dewey_health", map[string]any{})
	if err == nil {
		t.Fatal("expected error for init failure")
	}

	var unavail *UnavailableError
	if !errors.As(err, &unavail) {
		t.Errorf("expected UnavailableError, got %T: %v", err, err)
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention HTTP status: %v", err)
	}
}

func TestCall_SSEResponseWithJSONRPCError(t *testing.T) {
	// Custom server that returns a JSON-RPC error for tools/call.
	errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		method, _ := req["method"].(string)

		if method == "initialize" {
			rpcResp := map[string]any{
				"jsonrpc": "2.0",
				"id":      req["id"],
				"result": map[string]any{
					"protocolVersion": "2025-03-26",
					"capabilities":   map[string]any{},
				},
			}
			w.Header().Set("Mcp-Session-Id", "err-session")
			data, _ := json.Marshal(rpcResp)
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			return
		}

		// Return JSON-RPC error for tools/call.
		rpcResp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"error": map[string]any{
				"code":    -32600,
				"message": "invalid request",
			},
		}
		data, _ := json.Marshal(rpcResp)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
	}))
	defer errorSrv.Close()

	client := New(errorSrv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("test", map[string]any{})
	if err == nil {
		t.Fatal("expected error for JSON-RPC error response")
	}
	if !strings.Contains(err.Error(), "invalid request") {
		t.Errorf("error should contain message: %v", err)
	}
}

func TestCall_ConnectionRefused(t *testing.T) {
	client := New("http://127.0.0.1:1", Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("test", map[string]any{})
	if err == nil {
		t.Fatal("expected error for connection refused")
	}

	var unavail *UnavailableError
	if !errors.As(err, &unavail) {
		t.Errorf("expected UnavailableError, got %T: %v", err, err)
	}
}

func TestCall_MalformedJSONInSSE(t *testing.T) {
	srv, h := newMCPServer(t)
	h.malformedSSE = true

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("test", map[string]any{})
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("error should mention unmarshal: %v", err)
	}
}

func TestCall_TimeoutOnRequest(t *testing.T) {
	// Slow server that takes too long.
	slowSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer slowSrv.Close()

	client := New(slowSrv.URL, Config{
		Name:    "test-client",
		Version: "1.0.0",
		Timeout: 100 * time.Millisecond,
	})
	_, err := client.Call("test", map[string]any{})
	if err == nil {
		t.Fatal("expected timeout error")
	}

	var unavail *UnavailableError
	if !errors.As(err, &unavail) {
		t.Errorf("expected UnavailableError, got %T: %v", err, err)
	}
}

// --- Edge Case Tests ---

func TestCall_InitializeMissingSessionHeader(t *testing.T) {
	srv, h := newMCPServer(t)
	h.noSessionID = true

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	result, err := client.Call("dewey_health", map[string]any{})
	if err != nil {
		t.Fatalf("Call should succeed without session header: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed["status"] != "ok" {
		t.Errorf("status = %q, want %q", parsed["status"], "ok")
	}
}

func TestCall_PlainJSONResponse(t *testing.T) {
	srv, h := newMCPServer(t)
	h.plainJSON = true

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	result, err := client.Call("dewey_health", map[string]any{})
	if err != nil {
		t.Fatalf("Call with plain JSON response: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed["status"] != "ok" {
		t.Errorf("status = %q, want %q", parsed["status"], "ok")
	}
}

func TestCall_EmptyResponseBody(t *testing.T) {
	srv, h := newMCPServer(t)
	h.emptyBody = true

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("test", map[string]any{})
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	if !strings.Contains(err.Error(), "no valid response") {
		t.Errorf("error should mention no valid response: %v", err)
	}
}

func TestCall_SSENoDataLine(t *testing.T) {
	srv, h := newMCPServer(t)
	h.noDataLine = true

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("test", map[string]any{})
	if err == nil {
		t.Fatal("expected error for no data line")
	}
	if !strings.Contains(err.Error(), "no valid response") {
		t.Errorf("error should mention no valid response: %v", err)
	}
}

func TestCall_EmptyContentArray(t *testing.T) {
	srv, h := newMCPServer(t)
	h.emptyContent = true

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("test", map[string]any{})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	if !strings.Contains(err.Error(), "empty content") {
		t.Errorf("error should mention empty content: %v", err)
	}
}

// --- Contract Tests ---

func TestCall_ToolsCallEnvelopeCorrectness(t *testing.T) {
	var receivedReq map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		method, _ := req["method"].(string)

		if method == "initialize" {
			rpcResp := map[string]any{
				"jsonrpc": "2.0",
				"id":      req["id"],
				"result": map[string]any{
					"protocolVersion": "2025-03-26",
					"capabilities":   map[string]any{},
				},
			}
			w.Header().Set("Mcp-Session-Id", "envelope-session")
			data, _ := json.Marshal(rpcResp)
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			return
		}

		receivedReq = req

		toolResult := map[string]string{"status": "ok"}
		resultJSON, _ := json.Marshal(toolResult)
		rpcResp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result": map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": string(resultJSON)},
				},
			},
		}
		data, _ := json.Marshal(rpcResp)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
	}))
	defer srv.Close()

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("dewey_health", map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}

	// Verify the tools/call envelope.
	if receivedReq["method"] != "tools/call" {
		t.Errorf("method = %v, want %q", receivedReq["method"], "tools/call")
	}

	params, ok := receivedReq["params"].(map[string]any)
	if !ok {
		t.Fatalf("params is not a map: %T", receivedReq["params"])
	}
	if params["name"] != "dewey_health" {
		t.Errorf("params.name = %v, want %q", params["name"], "dewey_health")
	}

	args, ok := params["arguments"].(map[string]any)
	if !ok {
		t.Fatalf("params.arguments is not a map: %T", params["arguments"])
	}
	if args["key"] != "value" {
		t.Errorf("params.arguments.key = %v, want %q", args["key"], "value")
	}
}

func TestCall_CorrectHeadersOnInitialize(t *testing.T) {
	var initHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		method, _ := req["method"].(string)

		if method == "initialize" {
			initHeaders = r.Header.Clone()
			rpcResp := map[string]any{
				"jsonrpc": "2.0",
				"id":      req["id"],
				"result": map[string]any{
					"protocolVersion": "2025-03-26",
				},
			}
			data, _ := json.Marshal(rpcResp)
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Mcp-Session-Id", "header-session")
			fmt.Fprintf(w, "data: %s\n\n", data)
			return
		}

		// tools/call
		toolResult := map[string]string{"ok": "true"}
		resultJSON, _ := json.Marshal(toolResult)
		rpcResp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result": map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": string(resultJSON)},
				},
			},
		}
		data, _ := json.Marshal(rpcResp)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\n", data)
	}))
	defer srv.Close()

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("test", map[string]any{})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}

	if initHeaders.Get("Accept") != "application/json, text/event-stream" {
		t.Errorf("Accept = %q, want %q", initHeaders.Get("Accept"), "application/json, text/event-stream")
	}
	if initHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", initHeaders.Get("Content-Type"), "application/json")
	}
}

// --- Recovery Tests ---

func TestCall_SessionRecoveryOn400(t *testing.T) {
	srv, h := newMCPServer(t)
	h.rejectCount = 1 // Reject first tools/call, accept after re-init.

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	result, err := client.Call("dewey_health", map[string]any{})
	if err != nil {
		t.Fatalf("Call should succeed after recovery: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed["status"] != "ok" {
		t.Errorf("status = %q, want %q", parsed["status"], "ok")
	}

	// Should have initialized twice (once originally, once for recovery).
	if h.initCount.Load() != 2 {
		t.Errorf("init count = %d, want 2", h.initCount.Load())
	}
}

func TestCall_SessionRecoveryOn404(t *testing.T) {
	// Use a custom server that returns 404 for the first tools/call.
	var initCount atomic.Int64
	var toolCallCount atomic.Int64
	rejected := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		method, _ := req["method"].(string)

		if method == "initialize" {
			initCount.Add(1)
			rpcResp := map[string]any{
				"jsonrpc": "2.0",
				"id":      req["id"],
				"result": map[string]any{
					"protocolVersion": "2025-03-26",
					"capabilities":   map[string]any{},
				},
			}
			w.Header().Set("Mcp-Session-Id", "recovery-session")
			data, _ := json.Marshal(rpcResp)
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			return
		}

		toolCallCount.Add(1)
		if !rejected {
			rejected = true
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		toolResult := map[string]string{"status": "ok"}
		resultJSON, _ := json.Marshal(toolResult)
		rpcResp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result": map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": string(resultJSON)},
				},
			},
		}
		data, _ := json.Marshal(rpcResp)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
	}))
	defer srv.Close()

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	result, err := client.Call("dewey_health", map[string]any{})
	if err != nil {
		t.Fatalf("Call should succeed after 404 recovery: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed["status"] != "ok" {
		t.Errorf("status = %q, want %q", parsed["status"], "ok")
	}

	if initCount.Load() != 2 {
		t.Errorf("init count = %d, want 2", initCount.Load())
	}
}

func TestCall_RetryFailure(t *testing.T) {
	// Server that always returns 400 for tools/call.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		method, _ := req["method"].(string)

		if method == "initialize" {
			rpcResp := map[string]any{
				"jsonrpc": "2.0",
				"id":      req["id"],
				"result": map[string]any{
					"protocolVersion": "2025-03-26",
				},
			}
			w.Header().Set("Mcp-Session-Id", "retry-session")
			data, _ := json.Marshal(rpcResp)
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			return
		}

		http.Error(w, "always failing", http.StatusBadRequest)
	}))
	defer srv.Close()

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("test", map[string]any{})
	if err == nil {
		t.Fatal("expected error when retry also fails")
	}

	var unavail *UnavailableError
	if !errors.As(err, &unavail) {
		t.Errorf("expected UnavailableError, got %T: %v", err, err)
	}
}

// --- Concurrency Tests ---

func TestCall_ConcurrentInitialization(t *testing.T) {
	srv, h := newMCPServer(t)

	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})

	const goroutines = 10
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	wg.Add(goroutines)
	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = client.Call("dewey_health", map[string]any{})
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i, err)
		}
	}

	// With the mutex, only one goroutine should initialize.
	if h.initCount.Load() != 1 {
		t.Errorf("init count = %d, want 1 (concurrent init should be serialized)", h.initCount.Load())
	}
}

// --- Logger Tests ---

func TestCall_LoggerCalledOnSessionEvents(t *testing.T) {
	srv, h := newMCPServer(t)
	h.rejectCount = 1

	logger := &testLogger{}
	client := New(srv.URL, Config{
		Name:    "test-client",
		Version: "1.0.0",
		Logger:  logger,
	})

	_, err := client.Call("dewey_health", map[string]any{})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}

	if logger.infoCount < 1 {
		t.Errorf("expected at least 1 Info log, got %d", logger.infoCount)
	}
	if logger.warnCount < 1 {
		t.Errorf("expected at least 1 Warn log (recovery), got %d", logger.warnCount)
	}
}

func TestCall_NoLoggerOperatesSilently(t *testing.T) {
	srv, _ := newMCPServer(t)

	// No Logger configured — should not panic.
	client := New(srv.URL, Config{Name: "test-client", Version: "1.0.0"})
	_, err := client.Call("dewey_health", map[string]any{})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
}

// --- Default Config Tests ---

func TestNew_DefaultTimeout(t *testing.T) {
	client := New("http://example.com", Config{Name: "test", Version: "1.0.0"})
	if client.http.Timeout != 10*time.Second {
		t.Errorf("timeout = %v, want 10s", client.http.Timeout)
	}
}

func TestNew_CustomTimeout(t *testing.T) {
	client := New("http://example.com", Config{
		Name:    "test",
		Version: "1.0.0",
		Timeout: 5 * time.Second,
	})
	if client.http.Timeout != 5*time.Second {
		t.Errorf("timeout = %v, want 5s", client.http.Timeout)
	}
}

// --- Helpers ---

type testLogger struct {
	mu        sync.Mutex
	infoCount int
	warnCount int
}

func (l *testLogger) Info(msg string, keyvals ...any) {
	l.mu.Lock()
	l.infoCount++
	l.mu.Unlock()
}

func (l *testLogger) Warn(msg string, keyvals ...any) {
	l.mu.Lock()
	l.warnCount++
	l.mu.Unlock()
}
