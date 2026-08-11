// Package mcpclient provides an MCP Streamable HTTP transport client.
//
// The Client type handles the MCP session lifecycle: initialize handshake,
// Mcp-Session-Id management, tools/call envelope wrapping, dual-format
// response parsing (SSE + plain JSON), session recovery on HTTP 400/404,
// and concurrency safety for shared usage across goroutines.
package mcpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// maxResponseBytes limits the response body read to prevent unbounded memory
// consumption on malformed responses.
const maxResponseBytes = 10 * 1024 * 1024 // 10MB

// Config configures the MCP client.
type Config struct {
	// Name is the client identity sent in clientInfo.name during initialize.
	Name string

	// Version is the client version sent in clientInfo.version during initialize.
	Version string

	// Timeout is the per-request HTTP timeout. Default: 10s.
	Timeout time.Duration

	// Logger is an optional logger for session lifecycle events.
	// If nil, the client operates silently.
	Logger Logger
}

// Logger is an optional interface for structured logging of session lifecycle events.
type Logger interface {
	Info(msg string, keyvals ...any)
	Warn(msg string, keyvals ...any)
}

// Client is an MCP Streamable HTTP transport client that speaks to a remote
// MCP server (e.g., Dewey). It handles session lifecycle, envelope wrapping,
// and response parsing. It is safe for concurrent use.
type Client struct {
	url    string
	config Config
	http   *http.Client

	mu        sync.Mutex
	sessionID string
	inited    bool
	nextID    atomic.Int64
}

// New creates an MCP client for the given endpoint URL.
func New(url string, cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	c := &Client{
		url:    url,
		config: cfg,
		http: &http.Client{
			Timeout: timeout,
		},
	}
	c.nextID.Store(1)
	return c
}

// Call sends an MCP tools/call request, handling session initialization
// on first use and session recovery on HTTP 400/404.
// The method parameter is the MCP tool name (e.g., "dewey_health").
// Returns the tool result as raw JSON.
func (c *Client) Call(method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	if !c.inited {
		if err := c.initSession(); err != nil {
			c.mu.Unlock()
			return nil, err
		}
	}
	sessionID := c.sessionID
	c.mu.Unlock()

	result, statusCode, err := c.doToolsCall(method, params, sessionID)
	if err != nil && (statusCode == http.StatusBadRequest || statusCode == http.StatusNotFound) {
		// Session recovery: reset and retry once.
		c.mu.Lock()
		c.inited = false
		c.sessionID = ""
		if err := c.initSession(); err != nil {
			c.mu.Unlock()
			return nil, err
		}
		sessionID = c.sessionID
		c.mu.Unlock()

		if c.config.Logger != nil {
			c.config.Logger.Warn("session recovery triggered", "status", statusCode)
		}

		result, _, err = c.doToolsCall(method, params, sessionID)
		if err != nil {
			if c.config.Logger != nil {
				c.config.Logger.Warn("session recovery failed", "error", err)
			}
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

// initSession sends the MCP initialize handshake. Must be called with c.mu held.
func (c *Client) initSession() error {
	reqBody := map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialize",
		"id":      c.nextID.Add(1) - 1,
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":   map[string]any{},
			"clientInfo": map[string]any{
				"name":    c.config.Name,
				"version": c.config.Version,
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return &UnavailableError{Cause: fmt.Errorf("marshal initialize: %w", err)}
	}

	req, err := http.NewRequest(http.MethodPost, c.url, strings.NewReader(string(body)))
	if err != nil {
		return &UnavailableError{Cause: fmt.Errorf("create initialize request: %w", err)}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return &UnavailableError{Cause: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
		return &UnavailableError{
			Cause: fmt.Errorf("initialize failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody))),
		}
	}

	// Parse the response to verify it's a valid initialize response.
	_, err = c.parseResponse(resp)
	if err != nil {
		return &UnavailableError{Cause: fmt.Errorf("initialize response: %w", err)}
	}

	// Capture session ID if present.
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		c.sessionID = sid
	}

	c.inited = true

	if c.config.Logger != nil {
		c.config.Logger.Info("session initialized", "url", c.url)
	}

	return nil
}

// doToolsCall sends a tools/call request and returns the result.
func (c *Client) doToolsCall(method string, params any, sessionID string) (json.RawMessage, int, error) {
	reqBody := map[string]any{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"id":      c.nextID.Add(1) - 1,
		"params": map[string]any{
			"name":      method,
			"arguments": params,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal tools/call: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.url, strings.NewReader(string(body)))
	if err != nil {
		return nil, 0, fmt.Errorf("create tools/call request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, &UnavailableError{Cause: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
		return nil, resp.StatusCode, &UnavailableError{
			Cause: fmt.Errorf("tools/call failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody))),
		}
	}

	result, err := c.parseToolsCallResponse(resp)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return result, resp.StatusCode, nil
}

// jsonRPCResponse is a JSON-RPC 2.0 response envelope.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
	ID      any             `json:"id"`
}

// jsonRPCError is a JSON-RPC 2.0 error object.
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// mcpContent represents a single content block in an MCP tools/call result.
type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// mcpToolResult represents the result field of an MCP tools/call response.
type mcpToolResult struct {
	Content []mcpContent `json:"content"`
}

// parseResponse reads and parses an HTTP response as a JSON-RPC response,
// handling both SSE (text/event-stream) and plain JSON (application/json) formats.
func (c *Client) parseResponse(resp *http.Response) (*jsonRPCResponse, error) {
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	ct := resp.Header.Get("Content-Type")

	// Plain JSON response.
	if strings.HasPrefix(ct, "application/json") {
		var rpcResp jsonRPCResponse
		if err := json.Unmarshal(respBody, &rpcResp); err != nil {
			return nil, fmt.Errorf("unmarshal JSON response: %w", err)
		}
		if rpcResp.Error != nil {
			return nil, fmt.Errorf("JSON-RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
		}
		return &rpcResp, nil
	}

	// SSE response: scan for data: lines.
	for _, line := range strings.Split(string(respBody), "\n") {
		line = strings.TrimSpace(line)
		var data string
		if strings.HasPrefix(line, "data: ") {
			data = strings.TrimPrefix(line, "data: ")
		} else if strings.HasPrefix(line, "data:") {
			data = strings.TrimPrefix(line, "data:")
		} else {
			continue
		}

		var rpcResp jsonRPCResponse
		if err := json.Unmarshal([]byte(data), &rpcResp); err != nil {
			return nil, fmt.Errorf("unmarshal SSE data: %w", err)
		}
		if rpcResp.Error != nil {
			return nil, fmt.Errorf("JSON-RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
		}
		return &rpcResp, nil
	}

	return nil, fmt.Errorf("no valid response found in SSE stream")
}

// parseToolsCallResponse parses an HTTP response as an MCP tools/call result,
// extracting content[0].text as the tool result.
func (c *Client) parseToolsCallResponse(resp *http.Response) (json.RawMessage, error) {
	rpcResp, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var toolResult mcpToolResult
	if err := json.Unmarshal(rpcResp.Result, &toolResult); err != nil {
		return nil, fmt.Errorf("unmarshal tool result: %w", err)
	}

	if len(toolResult.Content) == 0 {
		return nil, fmt.Errorf("empty content array in tools/call response")
	}

	return json.RawMessage(toolResult.Content[0].Text), nil
}

// UnavailableError indicates the MCP endpoint is not reachable or returned
// a non-recoverable error.
type UnavailableError struct {
	Cause error
}

func (e *UnavailableError) Error() string {
	return fmt.Sprintf("mcp unavailable: %v", e.Cause)
}

func (e *UnavailableError) Unwrap() error {
	return e.Cause
}
