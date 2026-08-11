// Package memory provides a Dewey HTTP proxy client for semantic memory operations.
//
// The hivemind_store and hivemind_find tools proxy to Dewey's semantic search
// endpoints via MCP Streamable HTTP transport. Six secondary tools return
// deprecation messages pointing users to native Dewey tools.
//
// On connection failure, errors include a structured "DEWEY_UNAVAILABLE" code
// so agents can degrade gracefully.
package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/unbound-force/replicator/internal/mcpclient"
)

// Client is a Dewey HTTP proxy that forwards MCP tools/call requests.
type Client struct {
	mcp *mcpclient.Client
}

// NewClient creates a Dewey proxy client with a 10-second timeout.
func NewClient(deweyURL string) *Client {
	return &Client{
		mcp: mcpclient.New(deweyURL, mcpclient.Config{
			Name:    "replicator-memory",
			Version: "1.0.0",
			Timeout: 10 * time.Second,
		}),
	}
}

// Call sends an MCP tools/call request to the Dewey endpoint.
// The method parameter is the Dewey tool name (e.g., "dewey_health").
// Returns the tool result as raw JSON, or a structured error on failure.
func (c *Client) Call(method string, params any) (json.RawMessage, error) {
	result, err := c.mcp.Call(method, params)
	if err != nil {
		return nil, wrapError(err)
	}
	return result, nil
}

// Health pings the Dewey endpoint to verify connectivity.
func (c *Client) Health() error {
	_, err := c.Call("dewey_health", map[string]any{})
	return err
}

// Store proxies to store_learning with a deprecation warning.
func (c *Client) Store(information, tags string) (map[string]any, error) {
	params := map[string]any{
		"information": information,
	}
	if tags != "" {
		params["tags"] = tags
	}

	result, err := c.Call("store_learning", params)
	if err != nil {
		return nil, err
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		// If result isn't a map, wrap it.
		parsed = map[string]any{"result": string(result)}
	}

	// Add deprecation warning to response.
	parsed["_warning"] = "hivemind_store is deprecated. Use dewey_store_learning directly."

	return parsed, nil
}

// Find proxies to semantic_search with a deprecation warning.
func (c *Client) Find(query, collection string, limit int) (map[string]any, error) {
	params := map[string]any{
		"query": query,
	}
	if limit > 0 {
		params["limit"] = limit
	}
	if collection != "" {
		// Collection maps to source_type filter in Dewey.
		params["source_type"] = collection
	}

	result, err := c.Call("semantic_search", params)
	if err != nil {
		return nil, err
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		parsed = map[string]any{"result": string(result)}
	}

	// Add deprecation warning to response.
	parsed["_warning"] = "hivemind_find is deprecated. Use dewey_semantic_search directly."

	return parsed, nil
}

// UnavailableError indicates Dewey is not reachable.
type UnavailableError struct {
	Cause error
}

func (e *UnavailableError) Error() string {
	return fmt.Sprintf("dewey unavailable: %v", e.Cause)
}

func (e *UnavailableError) Unwrap() error {
	return e.Cause
}

// UnavailableResponse returns a structured JSON error for agents to parse.
func UnavailableResponse(err error) string {
	resp := map[string]any{
		"error":   err.Error(),
		"code":    "DEWEY_UNAVAILABLE",
		"message": "Dewey semantic search is not available. Memory operations require a running Dewey instance.",
	}
	out, _ := json.MarshalIndent(resp, "", "  ")
	return string(out)
}

// wrapError converts mcpclient.UnavailableError into memory.UnavailableError
// for backward compatibility with callers that check for *memory.UnavailableError.
func wrapError(err error) error {
	var mcpUnavail *mcpclient.UnavailableError
	if errors.As(err, &mcpUnavail) {
		return &UnavailableError{Cause: mcpUnavail.Cause}
	}
	return err
}
