## Why

OpenCode's Dynamic Context Protocol (DCP) compresses user-message content
during long conversations to stay within context limits. Slash command
files (`internal/agentkit/content/commands/*.md`) are delivered as user
messages, making them eligible for compression. When DCP compresses these
files, critical instructions -- execution order, safety invariants,
mandatory review gates -- can be lost or summarized, causing agents to
skip steps or violate workflow constraints.

DCP supports `<protect>` tags that mark content as compression-resistant.
Content inside `<protect>` blocks is preserved verbatim during
compression. This change adds `<protect>` tags to the five slash command
files, targeting the content whose loss would cause behavioral failures.

Prior work (forge-dcp-hardening, harden-forge-skill-compression,
harden-skill-compression) restructured content for better DCP survival
through positioning and constraint language. This change uses `<protect>`
tags -- a complementary mechanism that explicitly marks content as
non-compressible.

Resolves: unbound-force/replicator#74

## What Changes

Add `<protect>` tags to five slash command files in
`internal/agentkit/content/commands/`:

1. **forge.md** -- Protect the Critical Invariants section (review gate
   rules) and the ordered Workflow steps (1-8). The Rules section
   reiterates invariants and should also be protected.
2. **handoff.md** -- Protect the sequential workflow steps (0-5) and the
   ordering constraint ("Steps MUST execute in this exact order").
3. **inbox.md** -- Protect the 3-step Workflow section.
4. **forge-status.md** -- Protect the 3-step Workflow section.
5. **org.md** -- Protect the Common Actions tool reference and the
   Sessions section (session start/end is critical for handoff
   continuity).

Additionally, enable `protectTags: true` in the DCP configuration
(`opencode.json`) so that the tags are recognized by the compression
engine.

## Capabilities

### New Capabilities
- `dcp-protect-tags`: Slash command guardrails, workflows, and safety
  rules survive DCP compression intact via `<protect>` tag markup.

### Modified Capabilities
- `forge-command`: Critical invariants and workflow steps are now
  compression-resistant.
- `handoff-command`: Sequential workflow steps are now
  compression-resistant.
- `inbox-command`: Workflow steps are now compression-resistant.
- `forge-status-command`: Workflow steps are now compression-resistant.
- `org-command`: Tool reference and session instructions are now
  compression-resistant.

### Removed Capabilities
- None.

## Impact

- **Files modified**: 5 command markdown files under
  `internal/agentkit/content/commands/`, plus `opencode.json` for DCP
  config.
- **No Go code changes**: The `<protect>` tags are consumed by
  OpenCode's DCP engine, not by replicator's Go code. The agentkit
  embeds these files as-is.
- **No test changes**: The embedded content is delivered verbatim. Tags
  do not affect tool registration or MCP protocol behavior.
- **Risk**: Low. `<protect>` tags are inert markup when DCP is not
  active or when `protectTags` is disabled. The tags have no effect on
  agents that read the content directly.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change strengthens autonomous collaboration by ensuring that
workflow instructions agents rely on survive context compression. Agents
following slash command workflows (forge coordination, handoff sequences)
will receive the complete instructions regardless of conversation length.
No runtime coupling is introduced.

### II. Composability First

**Assessment**: PASS

The `<protect>` tags are inert when DCP is not active. The replicator
binary remains independently installable and usable without OpenCode's
DCP engine. The tags degrade gracefully -- they appear as no-op markup
in non-DCP contexts.

### III. Observable Quality

**Assessment**: N/A

This change modifies embedded markdown content, not tool outputs or
machine-parseable artifacts. No observable output formats are affected.

### IV. Testability

**Assessment**: PASS

The change is purely additive markup in embedded files. No test isolation
requirements are affected. The embedded files can be verified by
inspecting their content directly. No external services are required.
<!-- scaffolded by uf vdev -->
