## Why

The coordinator agent prompt (`internal/agentkit/content/agents/coordinator.md`) is 22 lines with all behavioral constraints expressed as 6 bullet points in a single "Rules" section. When tools like DCP compress session context, critical constraints — particularly the prohibition on file reservation and the mandatory review-before-complete ordering — are likely to be lost or weakened.

This is the same class of vulnerability identified in [unbound-force/unbound-force#346](https://github.com/unbound-force/unbound-force/issues/346). A coordinator that silently drops constraints under compression could reserve files (causing deadlocks with workers), skip reviews (bypassing quality gates), or call `forge_complete` before `forge_review` (circumventing verification).

Fixes [#46](https://github.com/unbound-force/replicator/issues/46).

## What Changes

Restructure the coordinator agent prompt to survive context compression by applying patterns already proven in the project's other agent files (`worker.md`, `background-worker.md`) and the forge coordination skill (`SKILL.md`).

Specific changes:
1. Add an opening identity statement that embeds key constraints inline, ensuring compressors retain them in any summary's opening sentence.
2. Create a dedicated "Critical Constraints" section with uppercase severity keywords (MUST/NEVER) positioned before the workflow section.
3. Add an explicit ordered protocol (numbered checklist) that enforces `forge_review` before `forge_complete`.
4. Separate boundary rules (what the coordinator must NOT do) from procedural steps (what it does).

## Capabilities

### New Capabilities
- None — this change modifies an existing embedded asset, not runtime code.

### Modified Capabilities
- `coordinator agent prompt`: Restructured for compression resilience with explicit ordering, identity reinforcement, and prominent negative constraints.

### Removed Capabilities
- None.

## Impact

- **File**: `internal/agentkit/content/agents/coordinator.md` (single file change)
- **Embedded asset**: The file is embedded via `go:embed` into the binary; changes take effect at next build.
- **Behavioral**: No runtime code changes. The coordinator's behavioral contract is preserved — this change makes existing constraints more explicit, not different.
- **Testing**: The agent prompt is an embedded text asset. Existing parity tests and embed tests cover the file's inclusion. No new test infrastructure needed, though a structure validation test could verify constraint positioning.

## Constitution Alignment

Assessed against the Replicator constitution (`.specify/memory/constitution.md`), which extends the Unbound Force org constitution v1.1.0.

### I. Autonomous Collaboration

**Assessment**: PASS

The coordinator prompt defines how the coordinator collaborates with workers through well-defined MCP tools and comms messaging. This change reinforces that contract by making the separation of concerns (coordinator orchestrates, workers reserve files and edit code) more explicit and compression-resistant.

### II. Composability First

**Assessment**: N/A

This change modifies an embedded text asset. It does not affect standalone functionality or introduce dependencies.

### III. Observable Quality

**Assessment**: PASS

The restructured prompt explicitly requires the coordinator to call `forge_review` for every worker completion before `forge_complete`, strengthening the quality gate. The use of uppercase MUST/NEVER keywords aligns with the severity signaling convention used in the forge coordination skill.

### IV. Testability

**Assessment**: PASS

The change affects a single embedded markdown file. The file is already covered by embed tests. A structural test MUST be added to verify that critical constraints appear before workflow steps, ensuring the compression-resilience property is maintained over time.
