---
tag: review-council
author: jay-flowers
category: pattern
created_at: 2026-08-11T18:06:52Z
identity: review-council-20260811T180652-jay-flowers
tier: draft
---

Review Council Spec Review Patterns: When running the review council on spec artifacts, the most consistently flagged issues across all reviewers were: (1) Thread safety for shared mutable state — 4/5 reviewers flagged missing concurrency safety for session state shared across goroutines (CRITICAL), (2) SSE parsing edge cases — 3/5 reviewers flagged missing scenarios for empty body, malformed JSON, no data line (HIGH), (3) Package naming collisions — architect flagged `mcphttp` vs existing `mcp` package confusion, renamed to `mcpclient` (CRITICAL), (4) Timeout budget for multi-round-trip operations — 3/5 reviewers flagged lazy init adding latency (HIGH), (5) Missing coverage strategy — testing reviewer flagged constitution violation (CRITICAL). Lesson: spec review council catches real implementation bugs before code is written. The thread safety finding alone would have caused CI `-race` failures.
