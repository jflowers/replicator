// Package mcp implements a JSON-RPC 2.0 server for the Model Context Protocol.
//
// The server communicates over stdio (stdin/stdout) using newline-delimited
// JSON-RPC messages. It handles two primary methods:
//   - tools/list: returns the list of available tools
//   - tools/call: executes a tool by name with arguments
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/unbound-force/replicator/internal/tools/registry"
)

// Logger is a minimal logging interface for the MCP server.
// It is intentionally small so callers can use charmbracelet/log,
// the standard library log/slog, or a test double.
type Logger interface {
	Info(msg any, keyvals ...any)
	Warn(msg any, keyvals ...any)
}

// Server is an MCP JSON-RPC server.
type Server struct {
	registry *registry.Registry
	version  string
	logger   Logger
}

// NewServer creates an MCP server backed by the given tool registry.
// The version string is reported in the initialize handshake.
// If logger is nil, tool call logging is silently skipped.
func NewServer(reg *registry.Registry, version string, logger Logger) *Server {
	return &Server{registry: reg, version: version, logger: logger}
}

// jsonrpcRequest is a JSON-RPC 2.0 request.
type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// jsonrpcResponse is a JSON-RPC 2.0 response.
type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// toolsListResult is the response shape for tools/list.
type toolsListResult struct {
	Tools []toolInfo `json:"tools"`
}

type toolInfo struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// toolsCallParams are the parameters for tools/call.
type toolsCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// toolsCallResult is the response shape for tools/call.
type toolsCallResult struct {
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ServeStdio reads JSON-RPC requests from stdin and writes responses to stdout.
func (s *Server) ServeStdio() error {
	return s.Serve(os.Stdin, os.Stdout)
}

// Serve reads JSON-RPC requests from r and writes responses to w.
func (s *Server) Serve(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	// Allow large messages (up to 10MB).
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			if err := s.writeError(w, nil, -32700, "Parse error"); err != nil {
				return err
			}
			continue
		}

		resp := s.handleRequest(&req)
		if err := s.writeResponse(w, resp); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req *jsonrpcRequest) *jsonrpcResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		}
	}
}

func (s *Server) handleInitialize(req *jsonrpcRequest) *jsonrpcResponse {
	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "replicator",
				"version": s.version,
			},
		},
	}
}

func (s *Server) handleToolsList(req *jsonrpcRequest) *jsonrpcResponse {
	tools := s.registry.List()
	infos := make([]toolInfo, len(tools))
	for i, t := range tools {
		infos[i] = toolInfo{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		}
	}

	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  toolsListResult{Tools: infos},
	}
}

func (s *Server) handleToolsCall(req *jsonrpcRequest) *jsonrpcResponse {
	var params toolsCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcError{Code: -32602, Message: "Invalid params"},
		}
	}

	tool := s.registry.Get(params.Name)
	if tool == nil {
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcError{Code: -32602, Message: fmt.Sprintf("Unknown tool: %s", params.Name)},
		}
	}

	start := time.Now()
	result, err := tool.Execute(params.Arguments)
	duration := time.Since(start)

	if err != nil {
		if s.logger != nil {
			s.logger.Warn("tool error", "tool", params.Name, "duration", duration, "error", err)
		}
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: toolsCallResult{
				Content: []contentBlock{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			},
		}
	}

	if s.logger != nil {
		s.logger.Info("tool call", "tool", params.Name, "duration", duration, "success", true)
	}

	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: toolsCallResult{
			Content: []contentBlock{{Type: "text", Text: result}},
		},
	}
}

func (s *Server) writeResponse(w io.Writer, resp *jsonrpcResponse) error {
	data, _ := json.Marshal(resp)
	if _, err := fmt.Fprintf(w, "%s\n", data); err != nil {
		return fmt.Errorf("write JSON-RPC response: %w", err)
	}
	return nil
}

func (s *Server) writeError(w io.Writer, id json.RawMessage, code int, msg string) error {
	resp := &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonrpcError{Code: code, Message: msg},
	}
	return s.writeResponse(w, resp)
}
