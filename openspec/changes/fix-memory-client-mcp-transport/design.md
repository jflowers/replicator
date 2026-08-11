## Context

`memory.Client.Call()` sends bare `http.Post()` to Dewey's MCP endpoint, which expects the MCP Streamable HTTP transport. The doctor health check (`internal/doctor/checks.go:deweyHealthProbe()`) was previously fixed and correctly speaks MCP, proving the pattern. The proxy client was not updated, causing all memory proxy tools to fail with HTTP 400.

The MCP Streamable HTTP transport requires:
- `Accept: application/json, text/event-stream` header on all requests
- An `initialize` handshake before any `tools/call` invocations
- `Mcp-Session-Id` header management from initialize response
- Tool invocations wrapped in `tools/call` JSON-RPC method with `name` and `arguments` params
- SSE response parsing (`event: message\ndata: {json}\n\n`)

## Goals / Non-Goals

### Goals
- Make `memory.Client.Call()` speak MCP Streamable HTTP transport
- Manage MCP session lifecycle (initialize on first use, cache session ID)
- Parse both SSE and plain JSON responses correctly
- Wrap tool names in `tools/call` method envelope
- Preserve graceful degradation (`DEWEY_UNAVAILABLE` semantics)
- Extract shared MCP transport logic usable by both doctor and memory client
- Update tests to use MCP-compatible mock handlers
- Ensure thread safety for concurrent access to shared session state

### Non-Goals
- Full MCP client library (only implement what's needed for `Call()`)
- Bidirectional SSE streaming (only request-response pattern)
- MCP `initialized` notification (Dewey does not enforce it; see spec note)
- MCP notifications or progress tokens
- Changing the `NewClient()` API or tool handler signatures
- `context.Context` propagation (existing codebase has zero context usage; would require API change)

## Decisions

### D1: Extract shared MCP client package as `internal/mcpclient/`

**Decision**: Extract the MCP Streamable HTTP transport logic into a shared package called `internal/mcpclient/` (not `internal/mcphttp/`).

**Rationale**: The project already has `internal/mcp/` which is the MCP JSON-RPC **server**. Naming the new package `mcphttp` would be ambiguous — it doesn't convey whether it's a client or server, and could be confused with an HTTP transport layer for the existing server. `mcpclient` clearly signals "MCP client" vs the existing `internal/mcp/` (server). The package contains a `Client` type that speaks MCP Streamable HTTP to external services like Dewey.

**Constitution alignment**: Composability First — the shared helper is independently testable and usable by any component that needs to speak MCP Streamable HTTP.

### D2: Lazy session initialization with concurrency safety

**Decision**: Initialize the MCP session on the first `Call()` invocation, not at `NewClient()` construction time. Use `sync.Mutex` to serialize initialization and protect session state.

**Rationale**: `NewClient()` is called at server startup (`cmd/replicator/serve.go:45`). Making it lazy means the server starts fast even when Dewey is down. Failed initialization returns `UnavailableError` (existing behavior preserved). Re-initialization on subsequent calls if session was never established.

The `mcpclient.Client` is a shared instance used by all tool handlers. While the current MCP server processes requests sequentially (`server.go:103`), the shared package must be safe for concurrent use by design — future server changes or other consumers may introduce concurrency. Session initialization uses a mutex to ensure exactly one goroutine performs the `initialize` handshake.

**Constitution alignment**: Composability First — the binary works standalone even when Dewey is unreachable. Testability — tests run with `-race` flag, so all shared state must be synchronized.

**Latency note**: First-call latency includes the `initialize` round-trip overhead (~5ms for localhost Dewey, up to 100-500ms for remote endpoints). Session recovery also incurs this overhead.

### D3: `tools/call` envelope wrapping

**Decision**: `Call()` continues to accept a tool name (e.g., `"dewey_health"`) as its `method` parameter, but internally wraps it in the MCP `tools/call` JSON-RPC method with `{name: toolName, arguments: params}` as the params object.

**Rationale**: This is transparent to callers — `Health()`, `Store()`, and `Find()` don't need to change. The MCP envelope is an internal transport detail.

### D4: Dual-format response parsing (SSE + plain JSON)

**Decision**: Handle both `text/event-stream` (SSE) and `application/json` responses. For SSE, use the same line-scanning approach as `deweyHealthProbe()` — read the full response body, split on newlines, find `data:` prefixed lines (with or without trailing space), unmarshal the JSON. For plain JSON, parse the body directly as JSON-RPC.

**Rationale**: The MCP Streamable HTTP spec allows servers to respond with either content type. The `Accept` header includes both. The `initialize` response may come as plain JSON. Handling both formats is necessary for correct protocol implementation.

The response body read is bounded via `io.LimitReader` (10MB limit) to prevent unbounded memory consumption.

### D5: `Mcp-Session-Id` management

**Decision**: Capture the `Mcp-Session-Id` header from the `initialize` response and attach it to all subsequent requests. If the header is absent, proceed without a session ID.

**Rationale**: Required by the MCP Streamable HTTP spec. Without it, the server may reject follow-up requests or create new sessions per request. However, not all MCP servers set this header, so its absence is not an error.

### D6: Configurable client identity and timeout

**Decision**: The `mcpclient.Client` constructor accepts a `Config` struct with `Name` (client identity for `clientInfo.name`), `Version` (for `clientInfo.version`), and `Timeout` (per-request HTTP timeout, default 10s).

**Rationale**: The shared package serves multiple consumers with different identities: `"replicator-memory"` for the proxy client, `"replicator-doctor"` for the health probe. The timeout must be configurable because the doctor uses 5s while the memory client uses 10s.

### D7: Optional structured logging

**Decision**: The `mcpclient.Client` optionally accepts a logger (compatible with `charmbracelet/log`) for session lifecycle events. If no logger is provided, the client operates silently.

**Rationale**: Observability for session lifecycle (initialization, recovery, failures) is important for operators diagnosing Dewey connectivity issues. The logger is optional to preserve the existing silent behavior for consumers that don't need it.

## Risks / Trade-offs

### Session stale/expired

If the MCP session expires (Dewey restart, timeout), subsequent calls will fail. **Mitigation**: On HTTP 400/404 from a `tools/call`, reset session state and retry once with a fresh `initialize`. This is a simple retry, not a complex reconnect loop. Session recovery is serialized with concurrent access via mutex.

### Timeout budget for multi-step operations

A single `Call()` can involve up to 4 HTTP round-trips in the worst case: `initialize` + `tools/call` + retry-`initialize` + retry-`tools/call`. With a 10-second per-request timeout, the worst-case wall-clock time is 40 seconds. **Accepted**: This is a pathological case (Dewey responding at exactly the timeout boundary). In practice, failures are fast (connection refused, HTTP 400). The per-request timeout (not per-`Call()` timeout) is simpler to implement and reason about. A future enhancement could add an overall deadline via `context.Context`.

### Single-event assumption

The SSE parser uses the first `data:` line per response. If Dewey starts streaming multi-event responses, the parser will only use the first one. **Accepted**: The current `tools/call` pattern returns single results. If streaming is needed later, this would be a new feature, not a fix for this bug.

### Doctor code duplication during transition

The doctor's `deweyHealthProbe()` will initially remain as-is. Task 3 migrates it to use the shared `internal/mcpclient/` package, resolving the duplication. If Task 3 is deferred (it is parallel-eligible), a tracking issue MUST be created before the PR is merged to ensure the migration happens in the next change cycle.

### No `initialized` notification

The MCP protocol requires an `initialized` notification after the `initialize` response. This client skips it because: (a) Dewey does not currently enforce it, (b) the doctor's working implementation also skips it, (c) it falls under the "full MCP client library" non-goal. If a future Dewey update requires it, the client will fail gracefully with `UnavailableError`.
<!-- scaffolded by uf vdev -->
