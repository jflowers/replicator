---
tag: mcpclient-architecture
author: jay-flowers
category: pattern
created_at: 2026-08-11T18:06:43Z
identity: mcpclient-architecture-20260811T180643-jay-flowers
tier: draft
---

Shared MCP Client Architecture (replicator): The `internal/mcpclient/` package provides a reusable MCP Streamable HTTP client shared by both `memory.Client` (proxy to Dewey tools) and `doctor.deweyHealthProbe()`. Key design decisions: (1) `mcpclient.Config` struct with `Name`, `Version`, `Timeout`, `Logger` for dependency injection, (2) lazy session initialization on first `Call()` rather than at construction time (avoids blocking `NewClient()` if Dewey is down), (3) `sync.Mutex` protects session state (`sessionID`, `inited`) for concurrent goroutine safety, (4) `sync/atomic.Int64` for monotonically increasing JSON-RPC request IDs, (5) `io.LimitReader` at 10MB bounds response body reads, (6) `memory.UnavailableError` preserved via `wrapError()` bridge for backward compatibility with existing callers that use `errors.As`. The doctor went from ~70 lines of inline MCP code to 7 lines.
