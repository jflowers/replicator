<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Create shared MCP client package

- [x] 1.1 Create `internal/mcpclient/` package with `Client` type that handles MCP Streamable HTTP transport: `initialize` handshake, `Mcp-Session-Id` management, `tools/call` envelope wrapping, dual-format response parsing (SSE + plain JSON), session recovery on HTTP 400/404, concurrency safety via `sync.Mutex`, configurable timeout and client identity via `Config` struct, optional structured logging, and `io.LimitReader` bounds on response bodies. Following TDD: write failing tests from Task 1.2 first, then implement to make them pass. Files: `internal/mcpclient/client.go`
- [x] 1.2 Write tests for the MCP client using `httptest.NewServer` with a stateful mock MCP handler that tracks initialization state, returns SSE-formatted responses, simulates session headers, and supports error scenarios. Must cover all scenarios from the coverage strategy: happy paths, error paths, edge cases (empty body, malformed JSON, no data line, missing session header, plain JSON response, empty content array), session recovery (400 + 404), concurrency safety (10 goroutines, must pass under `-race`), and timeout behavior. Files: `internal/mcpclient/client_test.go`

## 2. Integrate MCP client into memory proxy

- [x] 2.1 Rewrite `memory.Client` to use `mcpclient.Client` for all calls. Update `Call()` to delegate to the MCP client. Update `NewClient()` to construct an `mcpclient.Client` internally with `Name: "replicator-memory"`, `Version: "1.0.0"`, `Timeout: 10s`. Keep `Health()`, `Store()`, and `Find()` signatures unchanged. Update GoDoc comment on `Call()` to reflect MCP transport semantics. Following TDD: write/update failing tests from Task 2.2 first. Files: `internal/memory/proxy.go`
- [x] 2.2 Update `proxy_test.go` mock handlers to simulate MCP Streamable HTTP responses (SSE format, session headers, `tools/call` envelope validation). Verify `Health()`, `Store()`, and `Find()` work end-to-end through the MCP transport. Include a regression scenario: mock that rejects bare JSON-RPC (non-MCP) requests, confirming the old behavior would fail and the new behavior succeeds. Files: `internal/memory/proxy_test.go`

## 3. Migrate doctor to shared client

- [x] 3.1 [P] Refactor `deweyHealthProbe()` in `internal/doctor/checks.go` to use `mcpclient.Client` with `Name: "replicator-doctor"`, `Version: "1.0.0"`, `Timeout: 5s` instead of its inline MCP implementation. Remove the duplicated SSE parsing, header management, and initialize logic. Verify existing doctor tests still pass. Files: `internal/doctor/checks.go`

## 4. Verification

- [x] 4.1 Run `make check` — all tests pass, go vet clean (includes parity tests)
- [x] 4.2 Run `make check-coverage` — coverage ratchets pass
- [x] 4.3 Verify constitution alignment: Autonomous Collaboration (tools remain independently callable), Composability First (graceful degradation preserved), Observable Quality (JSON response shapes unchanged), Testability (all tests use `httptest`, no external services, `-race` passes)
- [x] 4.4 Update `AGENTS.md` project structure to include `internal/mcpclient/` package description
<!-- scaffolded by uf vdev -->
