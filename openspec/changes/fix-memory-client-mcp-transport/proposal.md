## Why

`memory.Client.Call()` in `internal/memory/proxy.go` sends bare `http.Post()` to the Dewey MCP endpoint. The MCP Streamable HTTP transport requires specific headers, session lifecycle management, SSE response parsing, and a `tools/call` method envelope. Without these, Dewey returns HTTP 400 and all memory proxy tools (`hivemind_store`, `hivemind_find`, and all deprecated stubs) fail with `DEWEY_UNAVAILABLE`.

This was identified as issue [#19](https://github.com/unbound-force/replicator/issues/19). The doctor health check (`internal/doctor/checks.go`) was already fixed in a prior change (`fix-dewey-doctor-check`) and correctly speaks MCP Streamable HTTP, proving the pattern works. The proxy client was not updated at that time.

## What Changes

Rewrite `memory.Client` to speak MCP Streamable HTTP transport:

1. **Session lifecycle** — Send an `initialize` handshake on first use, capture `Mcp-Session-Id` from response headers, attach it to subsequent requests.
2. **Request formatting** — Set `Accept: application/json, text/event-stream` header. Wrap tool calls in the `tools/call` JSON-RPC method with tool name and arguments in the params envelope.
3. **SSE response parsing** — Parse `text/event-stream` responses by scanning for `data: ` prefixed lines and extracting the JSON-RPC result.
4. **Graceful degradation** — Preserve existing `DEWEY_UNAVAILABLE` error semantics. Session initialization failures degrade gracefully (no panic, no crash).
5. **Test updates** — Update `proxy_test.go` mock handlers to simulate MCP Streamable HTTP responses (SSE format, session headers).

## Capabilities

### New Capabilities
- `mcp-session-management`: Client maintains MCP session state (`Mcp-Session-Id`) across calls within a process lifetime, with concurrency-safe access
- `sse-response-parsing`: Client correctly parses both SSE-formatted and plain JSON responses from MCP Streamable HTTP endpoints
- `mcpclient-package`: Shared `internal/mcpclient/` package reusable by any component needing MCP Streamable HTTP transport

### Modified Capabilities
- `memory.Client.Call()`: Switches from bare HTTP POST to full MCP Streamable HTTP transport (initialize + tools/call + SSE parsing). Delegates to new `mcpclient.Client` internally.
- `memory.Client.Health()`: Now works correctly against MCP-speaking Dewey endpoints (benefits from `Call()` transport fix — method signature and logic unchanged)
- `memory.Client.Store()`: Now works correctly against MCP-speaking Dewey endpoints (benefits from `Call()` transport fix — method signature and logic unchanged)
- `memory.Client.Find()`: Now works correctly against MCP-speaking Dewey endpoints (benefits from `Call()` transport fix — method signature and logic unchanged)

### Removed Capabilities
- None

## Impact

- **`internal/mcpclient/`**: New shared package providing MCP Streamable HTTP client (`Client` type) with session lifecycle, SSE/JSON parsing, concurrency safety, configurable timeout/identity.
- **`internal/memory/proxy.go`**: Rewrite of `Call()` method to delegate to `mcpclient.Client`. `NewClient()` constructs an `mcpclient.Client` internally. Method signatures unchanged.
- **`internal/memory/proxy_test.go`**: Test mock handlers must simulate MCP Streamable HTTP (SSE responses, session headers, initialize handshake).
- **`internal/doctor/checks.go`**: Migrate `deweyHealthProbe()` to use `mcpclient.Client` instead of inline MCP implementation (removes duplication).
- **`cmd/replicator/serve.go`**: No changes expected — `NewClient()` signature stays the same.
- **`AGENTS.md`**: Update project structure to include `internal/mcpclient/` package.
- **Downstream tools**: All memory tools (`hivemind_store`, `hivemind_find`, deprecated stubs) benefit automatically since they proxy through `memory.Client`.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The memory proxy tools remain independently callable via MCP. The `Client` produces self-describing JSON responses. Inter-agent communication patterns are unaffected — only the transport between replicator and Dewey changes.

### II. Composability First

**Assessment**: PASS

The binary continues to work standalone. Dewey integration degrades gracefully — when Dewey is unavailable or the MCP handshake fails, tools return `DEWEY_UNAVAILABLE` errors (existing behavior preserved). No new mandatory dependencies are introduced.

### III. Observable Quality

**Assessment**: PASS

All tool responses remain JSON. The MCP transport change is internal to `memory.Client` — response shapes visible to callers are unchanged and continue to match the TypeScript version's structure.

### IV. Testability

**Assessment**: PASS

Tests use `httptest.NewServer` with mock MCP handlers. No external services are required. The mock handlers simulate MCP Streamable HTTP responses (SSE format, session headers) to verify the full transport pipeline in isolation.
<!-- scaffolded by uf vdev -->
