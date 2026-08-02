## Why

Two deployed skills -- `forge-global` and `always-on-guidance` -- contain behavioral constraints structured in ways that DCP context compression is likely to weaken or drop entirely. Critical decision logic (conditional "when to / when not to" rules), safety constraints ("never force push to main"), and specific operational parameters (TTL settings) sit in list positions with low compression survival rates.

When these constraints are lost, agents make worse decisions: inappropriate parallelization of tightly-coupled tasks, hanging file reservations, re-solving already-solved problems, and potentially force-pushing to main. This is a quality and safety issue that affects every agent session loading these skills.

Tracked by [unbound-force/replicator#51](https://github.com/unbound-force/replicator/issues/51). Related to [unbound-force/unbound-force#346](https://github.com/unbound-force/unbound-force/issues/346).

## What Changes

Restructure the content of two embedded skill files to improve survival under DCP context compression, without changing the semantic meaning of any rule.

### forge-global (`internal/agentkit/content/skills/forge-global/SKILL.md`)

1. Replace parallel "When to / Don't" lists with a decision table -- decision structures survive compression better than opposing lists
2. Inline TTL guidance into the reservation step rather than a standalone bullet
3. Add explicit ordering markers to the File Reservation Protocol steps

### always-on-guidance (`internal/agentkit/content/skills/always-on-guidance/SKILL.md`)

1. Move "Check `hivemind_find` first" from last position to first position in Tool Usage (first-position = highest survival)
2. Separate critical safety rules into a dedicated `## Critical Safety` section with its own header
3. Reduce list sizes by grouping related rules under sub-headers (2-3 items per list instead of 5-6)

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `forge-global skill`: Restructured for compression resilience -- decision table replaces parallel lists, protocol steps gain ordering language, TTL guidance inlined
- `always-on-guidance skill`: Restructured for compression resilience -- safety rules elevated to dedicated section, list sizes reduced, critical items repositioned

### Removed Capabilities
- None

## Impact

- **Files**: `internal/agentkit/content/skills/forge-global/SKILL.md`, `internal/agentkit/content/skills/always-on-guidance/SKILL.md`
- **Embedded assets**: These skills are embedded via `go:embed` and distributed in the replicator binary
- **Agent behavior**: No semantic changes to any rules -- only structural reorganization for compression resilience
- **Tests**: Existing tests in `internal/agentkit/agentkit_test.go` (specifically `TestScaffold_FreshDirectory`, `TestScaffold_FileCount`, and `TestSkillTemplates_HaveNameField`) verify embedded skill file existence, count, and frontmatter validity. These serve as the regression baseline. Parity tests are unaffected (skills are not covered by parity tests).
- **Documentation**: No updates required to AGENTS.md or README.md -- this change is structural reorganization of existing embedded content with no user-facing behavior change.

## Constitution Alignment

Assessed against the Replicator project constitution (`.specify/memory/constitution.md`), which extends the Unbound Force org constitution v1.1.0.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies skill file structure only. It does not affect MCP tool interfaces, JSON outputs, or inter-agent communication via comms. Artifacts remain self-describing.

### II. Composability First

**Assessment**: PASS

Skills are embedded in the standalone binary. This change does not introduce any new dependencies or external service requirements. The binary continues to work alone.

### III. Observable Quality

**Assessment**: PASS

The restructured skills maintain all existing quality rules and behavioral constraints. No tool response shapes change. The change improves quality by ensuring constraints survive compression to actually reach agents.

### IV. Testability

**Assessment**: PASS

Existing tests in `internal/agentkit/agentkit_test.go` verify embedded skill file existence, count, and frontmatter validity. These tests serve as the regression baseline. `make build` verifies `go:embed` compilation. `make test` runs the full suite including agentkit tests. No new test infrastructure is required because the change modifies content within existing embedded files, not the embedding mechanism itself.
