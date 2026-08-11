package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// mcpMockHandler is a stateful mock MCP server for testing the memory proxy.
// It simulates MCP Streamable HTTP transport: initialize handshake, session ID,
// tools/call envelope validation, and SSE response format.
type mcpMockHandler struct {
	t *testing.T

	mu          sync.Mutex
	initialized bool

	// toolHandler receives the tool name and raw arguments, returns a result.
	toolHandler func(name string, args json.RawMessage) (any, error)
}

func (h *mcpMockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

	method, _ := req["method"].(string)

	switch method {
	case "initialize":
		h.handleInitialize(w, req)
	case "tools/call":
		h.handleToolsCall(w, req)
	default:
		// Reject bare JSON-RPC methods (non-MCP) with 400.
		http.Error(w, fmt.Sprintf("unknown method %q: MCP requires initialize + tools/call", method), http.StatusBadRequest)
	}
}

func (h *mcpMockHandler) handleInitialize(w http.ResponseWriter, req map[string]any) {
	h.mu.Lock()
	h.initialized = true
	h.mu.Unlock()

	result := map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":   map[string]any{},
		"serverInfo": map[string]any{
			"name":    "mock-dewey",
			"version": "1.0.0",
		},
	}

	rpcResp := map[string]any{
		"jsonrpc": "2.0",
		"id":      req["id"],
		"result":  result,
	}

	w.Header().Set("Mcp-Session-Id", "mock-session-id")
	data, _ := json.Marshal(rpcResp)
	w.Header().Set("Content-Type", "text/event-stream")
	fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
}

func (h *mcpMockHandler) handleToolsCall(w http.ResponseWriter, req map[string]any) {
	params, _ := req["params"].(map[string]any)
	toolName, _ := params["name"].(string)
	argsRaw, _ := json.Marshal(params["arguments"])

	var toolResult any
	if h.toolHandler != nil {
		var err error
		toolResult, err = h.toolHandler(toolName, argsRaw)
		if err != nil {
			rpcResp := map[string]any{
				"jsonrpc": "2.0",
				"id":      req["id"],
				"error": map[string]any{
					"code":    -32600,
					"message": err.Error(),
				},
			}
			data, _ := json.Marshal(rpcResp)
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			return
		}
	} else {
		toolResult = map[string]string{"status": "ok"}
	}

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
}

// newMCPTestServer creates a test server with the MCP mock handler.
func newMCPTestServer(t *testing.T, handler func(name string, args json.RawMessage) (any, error)) *httptest.Server {
	t.Helper()
	h := &mcpMockHandler{t: t, toolHandler: handler}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func TestCall_Success(t *testing.T) {
	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		return map[string]string{"status": "ok"}, nil
	})

	client := NewClient(srv.URL)
	result, err := client.Call("test_method", map[string]string{"key": "value"})
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
}

func TestCall_RPCError(t *testing.T) {
	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		return nil, fmt.Errorf("invalid request")
	})

	client := NewClient(srv.URL)
	_, err := client.Call("test_method", nil)
	if err == nil {
		t.Fatal("expected error for RPC error response")
	}
	if !strings.Contains(err.Error(), "invalid request") {
		t.Errorf("error = %q, should mention invalid request", err.Error())
	}
}

func TestCall_ConnectionRefused(t *testing.T) {
	// Use a URL that will refuse connections.
	client := NewClient("http://127.0.0.1:1")
	_, err := client.Call("test_method", nil)
	if err == nil {
		t.Fatal("expected error for connection refused")
	}

	var unavail *UnavailableError
	if !errors.As(err, &unavail) {
		t.Errorf("expected UnavailableError, got %T: %v", err, err)
	}
}

func TestCall_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.Call("test_method", nil)
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}

	var unavail *UnavailableError
	if !errors.As(err, &unavail) {
		t.Errorf("expected UnavailableError, got %T: %v", err, err)
	}
}

func TestHealth_Success(t *testing.T) {
	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		if name != "dewey_health" {
			t.Errorf("tool name = %q, want %q", name, "dewey_health")
		}
		return map[string]string{"status": "healthy"}, nil
	})

	client := NewClient(srv.URL)
	if err := client.Health(); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

func TestHealth_Failure(t *testing.T) {
	client := NewClient("http://127.0.0.1:1")
	err := client.Health()
	if err == nil {
		t.Fatal("expected error for unreachable Dewey")
	}
}

func TestStore_Success(t *testing.T) {
	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		if name != "store_learning" {
			t.Errorf("tool name = %q, want %q", name, "store_learning")
		}

		var p map[string]string
		json.Unmarshal(args, &p)
		if p["information"] != "test learning" {
			t.Errorf("information = %q, want %q", p["information"], "test learning")
		}
		if p["tags"] != "go,testing" {
			t.Errorf("tags = %q, want %q", p["tags"], "go,testing")
		}

		return map[string]any{"id": "mem-123", "stored": true}, nil
	})

	client := NewClient(srv.URL)
	result, err := client.Store("test learning", "go,testing")
	if err != nil {
		t.Fatalf("Store: %v", err)
	}

	if result["_warning"] == nil {
		t.Error("expected deprecation warning in response")
	}
	warning, ok := result["_warning"].(string)
	if !ok || warning == "" {
		t.Error("expected non-empty deprecation warning string")
	}
}

func TestStore_NoTags(t *testing.T) {
	var receivedArgs json.RawMessage

	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		receivedArgs = args
		return map[string]any{"stored": true}, nil
	})

	client := NewClient(srv.URL)
	_, err := client.Store("info only", "")
	if err != nil {
		t.Fatalf("Store: %v", err)
	}

	var params map[string]any
	json.Unmarshal(receivedArgs, &params)
	if _, hasTags := params["tags"]; hasTags {
		t.Error("tags should not be sent when empty")
	}
}

func TestStore_DeweyUnavailable(t *testing.T) {
	client := NewClient("http://127.0.0.1:1")
	_, err := client.Store("test", "")
	if err == nil {
		t.Fatal("expected error for unreachable Dewey")
	}

	var unavail *UnavailableError
	if !errors.As(err, &unavail) {
		t.Errorf("expected UnavailableError, got %T", err)
	}
}

func TestFind_Success(t *testing.T) {
	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		if name != "semantic_search" {
			t.Errorf("tool name = %q, want %q", name, "semantic_search")
		}

		var p map[string]any
		json.Unmarshal(args, &p)
		if p["query"] != "test query" {
			t.Errorf("query = %v, want %q", p["query"], "test query")
		}

		return map[string]any{
			"results": []map[string]string{
				{"page": "test-page", "score": "0.95"},
			},
		}, nil
	})

	client := NewClient(srv.URL)
	result, err := client.Find("test query", "", 5)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}

	if result["_warning"] == nil {
		t.Error("expected deprecation warning in response")
	}
}

func TestFind_WithCollection(t *testing.T) {
	var receivedArgs json.RawMessage

	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		receivedArgs = args
		return map[string]any{"results": []any{}}, nil
	})

	client := NewClient(srv.URL)
	_, err := client.Find("query", "learnings", 10)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}

	var params map[string]any
	json.Unmarshal(receivedArgs, &params)
	if params["source_type"] != "learnings" {
		t.Errorf("source_type = %v, want %q", params["source_type"], "learnings")
	}
}

func TestFind_WithLimit(t *testing.T) {
	var receivedArgs json.RawMessage

	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		receivedArgs = args
		return map[string]any{"results": []any{}}, nil
	})

	client := NewClient(srv.URL)
	_, err := client.Find("query", "", 7)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}

	var params map[string]any
	json.Unmarshal(receivedArgs, &params)
	// JSON numbers unmarshal as float64.
	if params["limit"] != float64(7) {
		t.Errorf("limit = %v, want 7", params["limit"])
	}
}

func TestFind_ZeroLimit(t *testing.T) {
	var receivedArgs json.RawMessage

	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		receivedArgs = args
		return map[string]any{"results": []any{}}, nil
	})

	client := NewClient(srv.URL)
	_, err := client.Find("query", "", 0)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}

	var params map[string]any
	json.Unmarshal(receivedArgs, &params)
	if _, hasLimit := params["limit"]; hasLimit {
		t.Error("limit should not be sent when zero")
	}
}

func TestFind_DeweyUnavailable(t *testing.T) {
	client := NewClient("http://127.0.0.1:1")
	_, err := client.Find("test", "", 5)
	if err == nil {
		t.Fatal("expected error for unreachable Dewey")
	}
}

func TestUnavailableResponse(t *testing.T) {
	err := &UnavailableError{Cause: errors.New("connection refused")}
	resp := UnavailableResponse(err)

	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(resp), &parsed); jsonErr != nil {
		t.Fatalf("unmarshal: %v", jsonErr)
	}

	if parsed["code"] != "DEWEY_UNAVAILABLE" {
		t.Errorf("code = %v, want %q", parsed["code"], "DEWEY_UNAVAILABLE")
	}
}

// TestCall_RejectsBareMethods verifies that the mock MCP server rejects bare
// JSON-RPC methods (the old behavior) and the new MCP transport succeeds.
// This is the regression test for issue #19.
func TestCall_RejectsBareMethods(t *testing.T) {
	srv := newMCPTestServer(t, func(name string, args json.RawMessage) (any, error) {
		return map[string]string{"status": "ok"}, nil
	})

	// The new client wraps in tools/call, so it should succeed.
	client := NewClient(srv.URL)
	_, err := client.Call("dewey_health", map[string]any{})
	if err != nil {
		t.Fatalf("MCP client should succeed: %v", err)
	}

	// Verify the mock rejects bare methods by sending a raw non-MCP request.
	bareReq := `{"jsonrpc":"2.0","method":"dewey_health","params":{},"id":1}`
	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(bareReq))
	if err != nil {
		t.Fatalf("bare request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bare method should be rejected with 400, got %d", resp.StatusCode)
	}
}
