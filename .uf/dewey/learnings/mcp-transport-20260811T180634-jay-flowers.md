---
tag: mcp-transport
author: jay-flowers
category: pattern
created_at: 2026-08-11T18:06:34Z
identity: mcp-transport-20260811T180634-jay-flowers
tier: draft
---

MCP Streamable HTTP Transport Pattern: When speaking MCP Streamable HTTP to a server like Dewey, the client must: (1) send an `initialize` handshake with `protocolVersion: "2025-03-26"`, `clientInfo.name/version`, and `capabilities: {}` before any `tools/call` invocations, (2) set `Accept: application/json, text/event-stream` and `Content-Type: application/json` headers on all requests, (3) capture the `Mcp-Session-Id` response header and attach it to subsequent requests, (4) wrap tool method names in a `tools/call` JSON-RPC envelope with `params.name` and `params.arguments`, (5) parse responses in dual format — check `Content-Type` for `application/json` (direct unmarshal) vs `text/event-stream` (scan for `data:` prefixed lines), (6) handle session recovery by resetting and re-initializing on HTTP 400/404. The `initialized` notification is skipped (Dewey doesn't enforce it). Bare `http.Post()` with plain JSON-RPC will get HTTP 400 from MCP endpoints.
