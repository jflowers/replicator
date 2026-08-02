## Context

The `forge-coordination` skill (`internal/agentkit/content/skills/forge-coordination/SKILL.md` and its mirror at `.opencode/skills/forge-coordination/SKILL.md`) encodes critical multi-agent safety constraints. These constraints are currently expressed in low-salience constructs — parenthetical asides, buried bullets, and separate sections — that DCP context compression discards first.

The proposal (constitution-aligned, all principles PASS or N/A) calls for restructuring the document to promote safety-critical constraints to positions that survive compression.

## Goals / Non-Goals

### Goals
- Restructure the skill to use role-scoped sections that clearly separate coordinator-only and worker-only operations
- Inline conflict resolution steps at the point where reservation failure occurs, rather than in a separate section
- Make `exclusive=true` the documented default for `comms_reserve`, not an option to remember
- Ensure the strongest phrasing (MUST/NEVER) appears first and at section-level prominence
- Keep both copies of the file in sync (agentkit embed + opencode skills)

### Non-Goals
- Changing MCP tool behavior or adding new tools
- Modifying Go source code
- Adding tests (no testable behavior changes)
- Changing the agentkit embed mechanism
- Restructuring other skills — this change targets forge-coordination only

## Decisions

### 1. Role-scoped sections replace mixed bullet lists

**Decision**: Replace the flat "File Reservation Rules" section with two explicit sections: "Coordinator-Only Operations" and "Worker-Only Operations." Each section states what that role MUST and MUST NOT do.

**Rationale**: Section headers are high-salience constructs that survive compression. A parenthetical like `(coordinator only)` is exactly the kind of qualifier DCP drops. Promoting it to a section header makes it structurally impossible to compress away without losing the entire section.

### 2. Inline conflict resolution at point of use

**Decision**: Move the conflict resolution steps from a separate section into the Worker Protocol, immediately after the "Reserve files" step.

**Rationale**: When compression removes "less important" sections, a separate "Conflict Resolution" section is a candidate for removal. If it's inlined at the step where reservation failure occurs, it's part of the protocol flow and survives as long as the protocol itself does.

### 3. `exclusive=true` as the documented default

**Decision**: Change the `comms_reserve` call in the Worker Protocol to include `exclusive=true` directly: `comms_reserve(paths=[...], exclusive=true, reason="...")`.

**Rationale**: When the parameter appears in the primary protocol step, agents copy it by default. When it's a separate bullet explaining an option, it's an optimization to drop.

### 4. Strongest constraint phrasing in primary position

**Decision**: Each role section opens with its MUST/MUST NOT rules as the first lines, before any protocol steps.

**Rationale**: First-position content in a section is the last thing compression removes. Burying MUST rules after descriptive text makes them candidates for trimming.

### 5. Both files updated identically

**Decision**: Both `internal/agentkit/content/skills/forge-coordination/SKILL.md` and `.opencode/skills/forge-coordination/SKILL.md` receive the same content.

**Rationale**: These files are currently identical. One is embedded into the binary for scaffolding; the other is loaded by opencode at runtime. Divergence would create inconsistent agent behavior.

## Risks / Trade-offs

### Risk: Increased document length
The restructured document will be slightly longer due to inlining and explicit role sections. This is an acceptable trade-off — a longer document with redundant safety constraints is better than a shorter document where safety constraints are compressed away.

### Risk: Divergence between the two file copies
Both files must be updated identically. The implementation should update one, then copy to the other, to minimize divergence risk. Existing agentkit embed tests will catch if the embedded copy is missing or malformed.

### Trade-off: Redundancy vs. DRY
Some constraints will appear in multiple places (e.g., "Workers MUST reserve files" appears in both the Worker Protocol steps and the Worker-Only Operations rules). This intentional redundancy ensures the constraint survives even if one occurrence is compressed.
