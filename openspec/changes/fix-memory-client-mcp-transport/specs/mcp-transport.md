## ADDED Requirements

### Requirement: MCP Session Initialization

The `mcpclient.Client` MUST send an MCP `initialize` handshake on first use before any `tools/call` invocations. The initialize request MUST include:
- `"method": "initialize"`
- `"params.protocolVersion": "2025-03-26"` (current MCP Streamable HTTP spec version)
- `"params.clientInfo.name"`: caller-provided name (e.g., `"replicator-memory"`, `"replicator-doctor"`)
- `"params.clientInfo.version"`: caller-provided version (e.g., `"1.0.0"`)
- `"params.capabilities": {}`

The client MUST capture the `Mcp-Session-Id` response header and attach it to all subsequent requests.

> **Note**: The MCP protocol requires an `initialized` notification after the `initialize` response. This client does NOT send the `initialized` notification — Dewey does not currently enforce it. This is a known deviation, accepted per the non-goal of "full MCP client library." If a future Dewey update requires the notification, session initialization will fail and the client will return `UnavailableError` (graceful degradation).

#### Scenario: First call triggers initialization

- **GIVEN** a `memory.Client` that has not yet been initialized
- **WHEN** `Call("dewey_health", {})` is invoked
- **THEN** the client MUST first send an `initialize` request to the Dewey endpoint
- **AND** capture the `Mcp-Session-Id` from the response headers
- **AND** then send the `tools/call` request with the session ID attached

#### Scenario: Subsequent calls reuse session

- **GIVEN** a `memory.Client` that has completed initialization
- **WHEN** `Call("store_learning", params)` is invoked
- **THEN** the client MUST NOT send another `initialize` request
- **AND** MUST include the cached `Mcp-Session-Id` header

#### Scenario: Initialize returns HTTP error

- **GIVEN** a Dewey endpoint that returns HTTP 500 on `initialize`
- **WHEN** `Call()` is invoked
- **THEN** it MUST return an `UnavailableError`
- **AND** the error message MUST include context about the initialization failure

#### Scenario: Initialize response missing session header

- **GIVEN** a Dewey endpoint that returns a valid `initialize` response but no `Mcp-Session-Id` header
- **WHEN** `Call()` is invoked
- **THEN** the client MUST proceed without a session ID (Dewey may not require it)
- **AND** subsequent requests MUST omit the `Mcp-Session-Id` header

### Requirement: Concurrency Safety

The `mcpclient.Client` MUST be safe for concurrent use by multiple goroutines. Session initialization MUST be serialized — if multiple goroutines call `Call()` concurrently on an uninitialized client, exactly one MUST perform the `initialize` handshake and all others MUST wait for it to complete.

The `Mcp-Session-Id` field MUST be protected against concurrent read/write access (e.g., via `sync.Mutex` or `sync.RWMutex`).

Session recovery (reset + re-initialize on HTTP 400/404) MUST also be serialized. If multiple goroutines detect session failure concurrently, only one MUST perform the re-initialization. Other goroutines MUST wait for recovery to complete and then retry with the new session.

#### Scenario: Concurrent first calls safely initialize once

- **GIVEN** a `mcpclient.Client` that has not yet been initialized
- **WHEN** 10 goroutines call `Call()` concurrently
- **THEN** exactly one `initialize` request MUST be sent to Dewey
- **AND** all goroutines MUST receive valid responses
- **AND** the test MUST pass under `-race`

### Requirement: MCP Request Headers

All requests to the Dewey endpoint MUST include:
- `Content-Type: application/json`
- `Accept: application/json, text/event-stream`

#### Scenario: Correct headers on initialize

- **GIVEN** a `mcpclient.Client` sending an initialize request
- **WHEN** the HTTP request is constructed
- **THEN** the `Accept` header MUST be `application/json, text/event-stream`
- **AND** the `Content-Type` header MUST be `application/json`

### Requirement: tools/call Envelope Wrapping

The `Call()` method MUST wrap the tool name and arguments in an MCP `tools/call` JSON-RPC method. The request body MUST have:
- `"method": "tools/call"`
- `"params.name": <tool-name>`
- `"params.arguments": <params-object>`

JSON-RPC request IDs SHOULD be monotonically increasing within a session to aid debugging. The `mcpclient.Client` maintains an internal counter.

#### Scenario: Tool name wrapped in envelope

- **GIVEN** a call to `Call("dewey_health", {})`
- **WHEN** the JSON-RPC request body is marshalled
- **THEN** the `method` field MUST be `"tools/call"`
- **AND** `params.name` MUST be `"dewey_health"`
- **AND** `params.arguments` MUST be `{}`

### Requirement: Response Parsing

The client MUST handle both `application/json` and `text/event-stream` response content types from the MCP endpoint.

**For `text/event-stream` (SSE) responses**, the parser MUST:
- Scan response lines for `data:` prefixes (with or without trailing space)
- Extract and unmarshal the JSON-RPC response from the first `data:` line
- For `tools/call` responses: return the `result.content[0].text` field as `json.RawMessage`
- For `initialize` responses: return the `result` object directly (different shape — has `protocolVersion`, `capabilities`, etc.)

**For `application/json` responses**, the parser MUST:
- Parse the body as a direct JSON-RPC response without SSE unwrapping

The response body read SHOULD be bounded (e.g., `io.LimitReader` with 10MB limit) to prevent unbounded memory consumption on malformed responses.

#### Scenario: Successful SSE response

- **GIVEN** a Dewey endpoint that returns `event: message\ndata: {"jsonrpc":"2.0","id":2,"result":{"content":[{"type":"text","text":"{\"status\":\"ok\"}"}]}}\n\n`
- **WHEN** the client parses the response
- **THEN** the result MUST be the JSON value `{"status":"ok"}`

#### Scenario: SSE response with JSON-RPC error

- **GIVEN** a Dewey endpoint that returns `event: message\ndata: {"jsonrpc":"2.0","id":2,"error":{"code":-32600,"message":"invalid request"}}\n\n`
- **WHEN** the client parses the response
- **THEN** `Call()` MUST return an error containing `"invalid request"`

#### Scenario: Plain JSON response (non-SSE)

- **GIVEN** a Dewey endpoint that returns `Content-Type: application/json` with body `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-03-26"}}`
- **WHEN** the client parses the response
- **THEN** the result MUST be parsed as a direct JSON-RPC response

#### Scenario: Empty SSE response body

- **GIVEN** a Dewey endpoint that returns HTTP 200 with an empty body
- **WHEN** the client parses the response
- **THEN** `Call()` MUST return an error indicating no valid response was found

#### Scenario: Malformed JSON in SSE data line

- **GIVEN** a Dewey endpoint that returns `data: {not-valid-json}`
- **WHEN** the client parses the response
- **THEN** `Call()` MUST return an error with unmarshal context

#### Scenario: SSE response with no data line

- **GIVEN** a Dewey endpoint that returns `event: message\n\n` (no `data:` line)
- **WHEN** the client parses the response
- **THEN** `Call()` MUST return an error indicating no valid response was found

#### Scenario: Empty content array in tools/call response

- **GIVEN** a Dewey endpoint that returns a `tools/call` response with `"content": []`
- **WHEN** the client parses the response
- **THEN** `Call()` MUST return an error (not panic with index-out-of-bounds)

### Requirement: Session Recovery on Failure

If a `tools/call` request fails with HTTP 400 or HTTP 404, the client MUST reset session state and retry once with a fresh `initialize` handshake.

Session recovery MUST be serialized with concurrent access (see Concurrency Safety requirement).

#### Scenario: Session expired and recovered (HTTP 400)

- **GIVEN** a `mcpclient.Client` with a cached session ID
- **WHEN** a `tools/call` request returns HTTP 400
- **THEN** the client MUST clear the cached session ID
- **AND** re-initialize with a new `initialize` request
- **AND** retry the original `tools/call` once

#### Scenario: Session expired and recovered (HTTP 404)

- **GIVEN** a `mcpclient.Client` with a cached session ID
- **WHEN** a `tools/call` request returns HTTP 404
- **THEN** the client MUST follow the same recovery procedure as HTTP 400

#### Scenario: Retry also fails

- **GIVEN** a `mcpclient.Client` attempting session recovery
- **WHEN** the re-initialized `tools/call` also fails
- **THEN** the client MUST return an `UnavailableError`
- **AND** MUST NOT retry further

### Requirement: Timeout Budget

The `mcpclient.Client` MUST accept a configurable per-request HTTP timeout. The timeout applies to each individual HTTP request, not to the entire `Call()` operation. The combined `initialize` + `tools/call` sequence has a maximum wall-clock time of `2 * timeout` on first call (or `4 * timeout` during session recovery with retry).

The default timeout SHOULD be 10 seconds (matching the existing `memory.Client` behavior).

#### Scenario: Timeout on initialize

- **GIVEN** a Dewey endpoint that does not respond within the timeout
- **WHEN** `Call()` is invoked on an uninitialized client
- **THEN** the client MUST return an `UnavailableError` after the timeout expires

### Requirement: Observability

The `mcpclient.Client` SHOULD accept an optional logger (compatible with `charmbracelet/log`) for structured logging at key lifecycle points:
- `INFO`: Session initialized successfully (Dewey URL)
- `WARN`: Session recovery triggered (HTTP status code)
- `WARN`: Session recovery failed (error detail)

If no logger is provided, the client MUST operate silently (no stdout/stderr output).

## MODIFIED Requirements

### Requirement: Call() Method Transport

`Call()` MUST use `http.NewRequest` with explicit `Accept` and `Content-Type` headers instead of `http.Post()`.

Previously: `Call()` used `c.http.Post(c.url, "application/json", ...)` which did not set the `Accept` header required by MCP Streamable HTTP.

### Requirement: Health Check via MCP

`Health()` MUST work correctly against MCP-speaking Dewey endpoints.

Previously: `Health()` called `Call("dewey_health", {})` which sent bare JSON-RPC without the MCP envelope, causing HTTP 400.

## REMOVED Requirements

None — no requirements are being removed.

## Coverage Strategy

The `internal/mcpclient/` package MUST achieve ≥80% line coverage. The following paths require dedicated test cases:

| Path | Test Category |
|------|---------------|
| Successful initialize + tools/call | Happy path |
| Session reuse (no re-init) | Happy path |
| Initialize HTTP error (500) | Error path |
| Initialize missing session header | Edge case |
| tools/call envelope correctness | Contract |
| SSE response parsing (valid) | Happy path |
| SSE response with JSON-RPC error | Error path |
| Plain JSON response (non-SSE) | Edge case |
| Empty response body | Edge case |
| Malformed JSON in SSE | Error path |
| No data line in SSE | Edge case |
| Empty content array | Edge case |
| Session recovery on 400 | Recovery |
| Session recovery on 404 | Recovery |
| Retry failure → UnavailableError | Recovery |
| Concurrent initialization | Concurrency |
| Timeout on request | Error path |

Each scenario in this spec MUST have at least one corresponding test function.
<!-- scaffolded by uf vdev -->
