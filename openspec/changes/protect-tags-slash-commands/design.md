## Context

Five slash command files in `internal/agentkit/content/commands/` deliver
workflow instructions to agents as user messages. DCP can compress these
during long sessions, losing critical steps. The proposal identifies
adding `<protect>` tags as the mechanism to prevent this, complementing
prior structural hardening work.

The proposal's constitution alignment confirms this change is PASS on
Autonomous Collaboration (strengthens agent instruction delivery),
Composability First (tags are inert without DCP), and Testability (no
external services needed). Observable Quality is N/A (no tool output
changes).

## Goals / Non-Goals

### Goals
- Wrap safety-critical content (invariants, review gates, ordered
  workflows, exit conditions) in `<protect>` tags across all five
  command files.
- Enable `protectTags: true` in `opencode.json` so the DCP engine
  recognizes the tags.
- Keep unprotected content available for DCP compression to maintain
  context budget efficiency.

### Non-Goals
- Restructuring command file content (already done in prior DCP
  hardening changes).
- Adding `<protect>` tags to skill files (skills are delivered as
  `tool_use` pairs and protected by `protectedTools`).
- Modifying replicator Go source code or MCP tool behavior.
- Adding tests for tag presence (the tags are consumed by OpenCode's DCP
  engine, not by replicator).

## Decisions

### D1: What to protect vs. leave unprotected

**Decision**: Protect content whose loss causes behavioral failure.
Leave explanatory/contextual content unprotected.

**Protected categories** (from issue #74):
1. **Guardrails / safety rules** -- invariant sections, MUST/NEVER
   constraints, review gate requirements.
2. **Execution checklists / step definitions** -- numbered workflow
   steps, tool call sequences.
3. **Exit conditions / resume instructions** -- session end procedures,
   handoff note structure.

**Unprotected categories**:
- Section headings (redundant with protected content).
- Descriptive text explaining *why* a step exists (the step itself is
  protected).
- Examples and code blocks that illustrate usage patterns (agents can
  reconstruct these from tool schemas).
- "Interpreting Results" guidance in forge-status.md (advisory, not
  behavioral).

**Rationale**: Over-protecting defeats the purpose of DCP compression.
The goal is surgical protection of content that, if lost, causes agents
to skip steps, violate ordering, or bypass gates.

### D2: Protect tag granularity

**Decision**: Use one `<protect>` block per logical section, not per
line.

For example, in forge.md:
- One `<protect>` block wrapping the entire Critical Invariants section.
- One `<protect>` block wrapping the entire Workflow section (steps 1-8).
- One `<protect>` block wrapping the Rules section.

**Rationale**: Per-line tagging adds noise and makes files harder to
maintain. Section-level tagging aligns with how DCP processes content
(it compresses by sections, not by lines). A single block per logical
unit keeps the markup clean.

### D3: File-by-file protect coverage

| File | Protected sections | Unprotected sections |
|------|-------------------|---------------------|
| forge.md | Critical Invariants, Workflow (1-8), Rules | Title, description, `$ARGUMENTS` placeholder |
| handoff.md | Workflow (steps 0-5 with ordering constraint and handoff note structure) | Title, description |
| inbox.md | Workflow (3 steps) | Title, description, Filtering, Sending |
| forge-status.md | Workflow (3 steps) | Title, description, Interpreting Results, Quick Health Check |
| org.md | Common Actions, Sessions | Title, description, Usage, Epics |

### D4: DCP configuration

**Decision**: Add `"protectTags": true` to the top level of
`opencode.json`.

**Rationale**: This is the toggle that tells OpenCode's DCP engine to
recognize `<protect>` tags. Without it, the tags are ignored. The
setting belongs in the project config so it applies to all agents using
this repository.

## Risks / Trade-offs

### R1: Over-protection reduces DCP effectiveness

If too much content is protected, DCP cannot compress enough and the
conversation may hit context limits sooner. **Mitigation**: The design
limits protection to behavioral content only -- roughly 60-70% of
forge.md and handoff.md, 30-40% of the smaller files.

### R2: Tag maintenance burden

Future edits to command files must maintain `<protect>` tag placement.
**Mitigation**: The tag pattern is simple (`<protect>` / `</protect>`)
and the files are small (24-52 lines each). The tags are visually
obvious during editing.

### R3: Tags visible in non-DCP contexts

Agents reading command files outside OpenCode's DCP engine will see the
`<protect>` tags as literal text. **Mitigation**: The tags are inert
HTML-style markup. They do not interfere with markdown rendering or
agent comprehension. This aligns with the Composability First principle.
<!-- scaffolded by uf vdev -->
