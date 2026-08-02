## Context

The `/forge` command prompt at `internal/agentkit/content/commands/forge.md` is a 67-line markdown file embedded in the replicator binary. It instructs the coordinator agent to decompose tasks, spawn workers, review their output, and mark work complete. The file currently structures its content as: workflow steps, rules, strategy selection, monitoring, completion, error recovery -- in that order.

Under DCP (Dynamic Context Protocol) context compression, long prompts get summarized. Research on prompt compression behavior shows:

- **Opening content** is preserved with highest fidelity
- **Numbered sequences** retain better than bullet lists, but adjacent steps with similar semantics get merged
- **Middle items** in lists are most likely to be dropped
- **Trailing sections** (especially "error handling" or "edge cases") are first to be entirely removed
- **Explicit constraint language** (MUST, NEVER, FIRST) survives better than implicit ordering

The current forge.md has all four fragilities identified in issue #47.

## Goals / Non-Goals

### Goals

- Restructure forge.md so the review-before-complete ordering constraint survives context compression
- Move critical invariants to the top of the file where compressors preserve them
- Inline error recovery and strategy selection at their point of use rather than as droppable trailing sections
- Preserve all existing workflow semantics -- same steps, same tools, same order

### Non-Goals

- Changing the forge workflow behavior or adding new steps
- Adding new MCP tool calls or modifying tool signatures
- Restructuring other command prompts (those are separate changes)
- Implementing programmatic enforcement of review-before-complete (that would be a code change, not a prompt change)

## Decisions

### D1: Critical Invariants section at top of file

Place a "Critical Invariants" section immediately after the title and before the Workflow section. This section states the non-negotiable rules in explicit constraint language. Compressors prioritize opening content, so these rules have the highest survival rate.

Content: The review-before-complete ordering, the "always create a forge" rule, the coordinator-orchestrates/workers-execute boundary, and the prohibition on `skip_review: true` in `forge_complete` calls.

### D2: Inline ordering constraint in step text

Change step 7 from:
```
7. Review: forge_review(task_id, files_touched) for each completed worker
```
To:
```
7. Review FIRST (before complete): forge_review(task_id, files_touched) for each completed worker — MUST finish before step 8
```

This embeds the ordering constraint directly in the step text so it cannot be separated from the step by compression.

### D3: Reorder Rules section by survival priority

Move "Review every worker's output before marking complete" from 5th bullet (lowest survival position) to 1st bullet. First and last items in lists survive compression best.

### D4: Inline strategy selection with decompose step

Instead of a separate "Strategy Selection" section (lines 34-42) that can be dropped entirely, inline the `forge_get_strategy_insights` call as a sub-step of step 3 (Decompose). This ensures strategy selection is never separated from decomposition.

### D5: Inline error recovery with monitoring step

Instead of a trailing "Error Recovery" section (lines 61-67), inline recovery guidance as sub-items of step 6 (Monitor). Trailing sections are first to be dropped; inlined content at the point of use survives with the parent step.

### D6: Consolidate Completion section into step 8

The separate "Completion" section (lines 52-59) repeats step 8 with more detail. Merge its content into step 8's text to eliminate the redundancy that compressors exploit (they drop the "duplicate" section).

## Risks / Trade-offs

### Risk: Longer step descriptions reduce readability

Inlining strategy selection, error recovery, and completion details into workflow steps makes each step longer. This trades readability for compression resilience.

**Mitigation**: Use sub-items (indented bullets) under steps rather than paragraph text. This preserves scanability while keeping content co-located.

### Risk: Over-engineering for a theoretical problem

DCP compression behavior is based on observed patterns, not guaranteed specifications. The fragilities may never manifest in practice.

**Mitigation**: The restructuring preserves all content and semantics. Even without compression, the new structure is arguably better organized (invariants first, related content co-located). No downside risk.

### Risk: `forge_complete` API has `skip_review` parameter

The `forge_complete` MCP tool accepts `skip_review: true` and `skip_verification: true` parameters (see `internal/tools/forge/tools.go`). Under DCP compression, a compressed agent could discover or hallucinate this parameter and bypass the review gate entirely, regardless of prompt hardening.

**Mitigation**: Add "NEVER pass `skip_review: true` to `forge_complete`" to the Critical Invariants section. This is a prompt-level defense; programmatic enforcement (removing or guarding the parameter) is out of scope for this change but should be tracked as a follow-up. The redundancy of the constraint (invariants + step 7 text + first rule) means even partial compression still leaves the prohibition visible.

**Follow-up**: Consider removing or access-gating the `skip_review` parameter in a separate change.

### Risk: Worker prompts also contain completion instructions

The `forge_spawn_subtask` tool generates worker prompts that include "Complete with `forge_complete` when done" — with no mention of the review gate. If a worker calls `forge_complete` directly, it bypasses the coordinator's review step.

**Mitigation**: The Critical Invariants section should state "Workers MUST NOT call `forge_complete` — only the coordinator completes work after review." This is a follow-up hardening target for the worker prompt (`internal/forge/spawn.go`), tracked separately from this change.

### Trade-off: File grows slightly from explicit constraint language

Adding "MUST", "FIRST", "before step 8" makes the file marginally longer. This is acceptable because explicit constraint language has higher compression survival than implicit ordering.
