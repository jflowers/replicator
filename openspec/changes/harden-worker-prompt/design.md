## Context

`internal/agentkit/content/agents/worker.md` is 27 lines with two sections: a numbered Checklist (lines 14-20) and a flat-bullet Constraints section (lines 22-27). Under LLM context compression, the Checklist survives well because it has structure (numbered steps with tool names), but the Constraints section is vulnerable to summarization because it's a flat list of behavioral restrictions separated from the actions they govern.

The proposal identifies three specific fragilities: reservation-check constraints buried in a droppable section, no recovery path for reservation failures, and progress reporting steps that look like optional middleware.

## Goals / Non-Goals

### Goals
- Integrate critical constraints inline with the checklist steps they govern, so they survive compression as a unit
- Add an explicit recovery path for reservation failures (expired, released, or failed to acquire)
- Frame progress reporting and learning storage as mandatory structural steps, not optional extras
- Consolidate redundant constraint phrasings ("Only edit files you have reserved" and "Never modify files outside your assignment") into a single, clear statement at the point of action

### Non-Goals
- Changing the worker's behavioral semantics — the worker does the same things, just described more durably
- Modifying the coordinator prompt or forge tool implementations
- Adding new MCP tools or changing tool response shapes
- Restructuring other agent files (background-worker, coordinator) — those are separate concerns
- Adding programmatic enforcement of reservations at the tool level — this change is prompt-level hardening only

## Decisions

### D1: Inline constraints with checklist steps, eliminate separate Constraints section

The Constraints section will be removed entirely. Each constraint will be integrated into the checklist step it governs:

- "Only edit files you have reserved" / "Never modify files outside your assignment" → merged into step 3 (`comms_reserve`) and step 4 (implement)
- "Report progress at regular intervals" → already expressed in step 5, will be strengthened with mandatory language
- "Store learnings for future agents" → already expressed in step 6, will be strengthened

**Rationale**: A single numbered list is the most compression-resistant prompt structure. Constraints co-located with their actions form atomic units that a compressor must keep or drop together. This aligns with Autonomous Collaboration (Principle I) by making the reservation enforcement — the primary collision prevention mechanism — structurally durable.

### D2: Add reservation failure recovery as a sub-step of step 3

Step 3 will include an explicit "if reservation fails" clause instructing the worker to STOP and report to the coordinator via `comms_send`. This covers three failure modes: initial acquisition failure, TTL expiration, and coordinator-initiated release.

**Rationale**: Without recovery guidance, a worker with a lost reservation will either stall silently or proceed unsafely. The recovery path uses the existing comms protocol (Autonomous Collaboration) and produces an observable failure state (Observable Quality, Principle III).

### D3: Use imperative "MUST" / "NEVER" language inline

Rather than soft phrasing ("Only edit..."), constraints will use imperative RFC 2119 language: "NEVER edit unreserved files", "MUST report progress". This language is more likely to be preserved by compressors because it signals importance.

**Rationale**: Compressors weight imperative/directive language higher than descriptive language when deciding what to preserve.

### D4: Keep the file under 35 lines

The hardened version should stay concise. Adding recovery instructions and inline constraints should not balloon the file. Target: under 35 lines (up from 27).

**Rationale**: Longer prompts are themselves more susceptible to compression. The goal is to make 27 lines more durable, not to add 50 lines of instructions.

## Risks / Trade-offs

### R1: Removing the Constraints section reduces scanability
**Risk**: Developers reviewing the worker prompt can no longer scan a dedicated section for all constraints.
**Mitigation**: The checklist with inline constraints is still scannable — each step now reads as "do X, and Y is the boundary." The file is short enough that full reading takes seconds.

### R2: Compression resistance is not guaranteed
**Risk**: No prompt structure is fully compression-proof. Even inline constraints could be summarized away by an aggressive compressor.
**Mitigation**: This change makes compression resistance significantly better, not perfect. The structural approach (constraints co-located with actions) is a well-known hardening pattern. Further defense would require programmatic enforcement at the tool level (out of scope).

### R3: Imperative language may seem harsh to human readers
**Risk**: "NEVER" and "MUST" phrasing reads differently than "Only edit...".
**Mitigation**: These prompts are consumed by LLM agents, not human users. RFC 2119 language is the project convention for requirements (per constitution and convention packs). Human-facing documentation uses softer language.
