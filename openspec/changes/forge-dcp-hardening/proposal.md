## Why

The `/forge` command prompt (`internal/agentkit/content/commands/forge.md`) defines a 9-step workflow where Step 7 (review) must happen before Step 8 (complete). Under DCP context compression, this ordering dependency is the most likely constraint to be lost -- the two steps get compressed into a single "review and complete workers" action, allowing the agent to skip reviews or complete before reviewing.

This is a quality gate vulnerability. The review step is the only point where the coordinator validates worker output before marking it done. If skipped, broken or incomplete work gets marked as complete with no human-visible signal that review was bypassed.

Related: [unbound-force/replicator#47](https://github.com/unbound-force/replicator/issues/47), [unbound-force/unbound-force#346](https://github.com/unbound-force/unbound-force/issues/346).

## What Changes

Restructure `forge.md` to survive DCP context compression by applying prompt hardening techniques:

1. Add a "Critical Invariants" section at the top of the file (before Workflow) stating non-negotiable rules -- compressors prioritize opening content
2. Embed ordering constraints directly in step text: "7. Review FIRST: ... -- MUST complete before step 8"
3. Move "Review every worker before marking complete" from 5th bullet to 1st position in Rules
4. Inline error recovery guidance at the monitoring step rather than as a separate droppable section at the end
5. Add strategy selection reminder inline with decompose step

## Capabilities

### New Capabilities

- Review gate mandatory constraint added to Critical Invariants — states the review gate is non-negotiable using positive constraint language (avoids naming bypass parameters)

### Modified Capabilities

- `/forge` command: Same workflow behavior, restructured prompt for compression resilience

### Removed Capabilities

- None

## Impact

- **File**: `internal/agentkit/content/commands/forge.md` (single file change)
- **Behavioral**: No change to forge workflow semantics -- agents follow the same steps in the same order
- **Risk**: Low -- restructuring prompt text only, no code changes
- **Testing**: Existing parity tests continue to pass. `TestForgeMD_StructuralHardening` verifies structural invariants (section ordering, redundant constraint placement, prohibited standalone sections) via 7 subtests against the embedded forge.md content.

## Constitution Alignment

Assessed against the Replicator constitution (`.specify/memory/constitution.md`), which extends the Unbound Force org constitution v1.1.0.

### I. Autonomous Collaboration

**Assessment**: PASS

This change improves the reliability of artifact-based coordination. The forge command orchestrates autonomous workers through well-defined tool calls. Hardening the prompt ensures the review gate -- which validates worker artifacts before marking them complete -- survives context compression. Self-describing outputs are unaffected.

### II. Composability First

**Assessment**: N/A

No new dependencies introduced. The forge command continues to work standalone. This is a prompt text restructuring within an existing embedded file.

### III. Observable Quality

**Assessment**: PASS

The change strengthens observable quality by ensuring the review step (which validates machine-parseable worker output) is not skipped under compression. No changes to tool response shapes or JSON output.

### IV. Testability

**Assessment**: PASS

`TestForgeMD_StructuralHardening` (7 subtests) validates structural invariants of the embedded forge.md content: section ordering, constraint presence, redundant placement, and absence of prohibited standalone sections. Tests use in-memory embedded filesystem — no external services required.
