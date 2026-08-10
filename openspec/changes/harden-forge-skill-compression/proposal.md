## Why

The `forge-coordination` skill contains critical access-control and operational constraints expressed as parenthetical asides, weak bullet items, and separated sections. DCP context compression drops these low-salience constructs first, causing agents to lose safety-critical rules:

- `(coordinator only)` parenthetical on `comms_release_all()` — compressed away, workers could release all reservations system-wide
- `exclusive=true` as one bullet in a list — omitted, workers call `comms_reserve` without exclusivity, enabling concurrent edits
- Conflict resolution steps separated from the point of failure — compressed to "resolve conflicts," skipping negotiation
- Duplicate constraints at different strength levels across files — compressor may pick the weakest phrasing

References: [unbound-force/replicator#48](https://github.com/unbound-force/replicator/issues/48), [unbound-force/unbound-force#346](https://github.com/unbound-force/unbound-force/issues/346)

## What Changes

Restructure the `forge-coordination` skill to survive DCP context compression by promoting critical constraints to high-salience positions (section headers, MUST rules at section top, inline at point-of-use).

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `forge-coordination skill`: Restructured to use explicit role-scoped sections ("Coordinator-Only Operations", "Worker-Only Operations"), inline conflict resolution at point of use, strongest constraint phrasing in primary position, and `exclusive=true` as the documented default

### Removed Capabilities
- None

## Impact

- **File**: `internal/agentkit/content/skills/forge-coordination/SKILL.md` (embedded agentkit copy)
- **File**: `.opencode/skills/forge-coordination/SKILL.md` (opencode skills copy)
- Both files must stay in sync — they are currently identical
- The Worker Protocol is expanded from 7 to 8 steps (explicit "Release files" step added as step 7; see design decision D5) — this is an intentional behavioral modification to the protocol, not a Go code or MCP tool change
- A structural test (`TestForgeCoordinationSkill_StructuralHardening`) validates compression-critical patterns survive editing
- Agents loading this skill will receive compression-resistant constraints

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change strengthens artifact-based communication by ensuring the skill document (an artifact consumed by autonomous agents) retains its critical constraints under compression. The change maintains self-describing outputs and does not alter inter-agent communication protocols.

### II. Composability First

**Assessment**: N/A

No dependencies are introduced or removed. The skill file remains a standalone document. No changes to binary functionality or Dewey integration.

### III. Observable Quality

**Assessment**: N/A

No MCP tool responses or machine-parseable outputs are changed. This is a prompt document restructuring.

### IV. Testability

**Assessment**: N/A

No testable components are added or modified. The change targets prompt content only. Existing agentkit embed tests continue to verify the file is properly embedded.
