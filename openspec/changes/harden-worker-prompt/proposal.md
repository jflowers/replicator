## Why

The `worker.md` agent prompt (27 lines) contains critical behavioral constraints — file reservation enforcement, progress reporting, and learning storage — in a separate "Constraints" section that is vulnerable to context compression. When LLM context windows fill up, compressors tend to summarize or drop flat bullet lists of constraints while preserving structured checklists. This means workers can silently edit unreserved files (causing collisions with other workers), skip progress reporting (leaving coordinators blind), or ignore learning storage.

This is the same class of vulnerability identified in [unbound-force/unbound-force#346](https://github.com/unbound-force/unbound-force/issues/346). Fixing it in `worker.md` hardens the most critical agent in the forge pipeline — the one that actually touches code files.

Ref: [unbound-force/replicator#49](https://github.com/unbound-force/replicator/issues/49)

## What Changes

Restructure `worker.md` to make critical constraints compression-resistant by integrating them into the numbered checklist steps rather than keeping them in a separate section. Add a recovery path for reservation failures. Consolidate redundant constraint phrasings.

## Capabilities

### New Capabilities
- `reservation-failure-recovery`: Worker now has explicit instructions for handling expired, released, or failed reservations — STOP and report to coordinator via `comms_send`.

### Modified Capabilities
- `worker-checklist`: Checklist steps now embed their constraints inline (e.g., step 3 includes "NEVER edit unreserved files" directly). Progress reporting and learning storage are framed as mandatory structural steps, not optional middleware.

### Removed Capabilities
- _None_

## Impact

- **File changed**: `internal/agentkit/content/agents/worker.md`
- **Embedded asset**: This file is embedded via Go's `embed` package into the binary, so changes take effect at next build
- **Behavioral**: Workers become more resilient to context compression — critical constraints survive alongside the checklist they govern
- **Test impact**: Parity tests for agent file embedding should still pass (content changes, not structural changes to the embedding mechanism)
- **Coordinator impact**: None — the coordinator's protocol for spawning/reviewing workers is unchanged

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change strengthens autonomous collaboration by ensuring worker agents retain their reservation constraints and progress reporting requirements even under context compression. Reservation enforcement is the mechanism that prevents workers from colliding on shared files. Progress reporting via `forge_progress` is the mechanism coordinators use to track worker state. Both are artifact-based coordination patterns that this change makes more durable.

### II. Composability First

**Assessment**: N/A

This change modifies an agent prompt file. It does not affect binary independence, Dewey integration graceful degradation, or database schema compatibility. The worker agent continues to function identically regardless of which external services are available.

### III. Observable Quality

**Assessment**: PASS

Progress reporting (`forge_progress` at milestones) is the worker's primary observability mechanism. By integrating it as a mandatory structural step rather than a droppable constraint, this change improves the reliability of observable quality in multi-agent workflows. No changes to MCP tool response shapes or parity test expectations.

### IV. Testability

**Assessment**: N/A

This change modifies prompt content, not testable code. The embedding mechanism and parity tests are unaffected. No new external service dependencies are introduced.
