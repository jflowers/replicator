## Context

The `/handoff` command (`internal/agentkit/content/commands/handoff.md`) is a 25-line embedded prompt that instructs agents to perform a 5-step session teardown. The current structure uses a numbered list for step ordering and a separate section for the handoff note template. Under DCP context compression, numbered lists can be reordered or merged, and standalone template sections can be dropped entirely.

The related issue [unbound-force/unbound-force#346](https://github.com/unbound-force/unbound-force/issues/346) demonstrated that `/review-pr`'s confirmation gate was bypassed under the same compression conditions. The fix there added explicit "MANDATORY GATE" markers and session-resume guards -- patterns we adapt here.

## Goals / Non-Goals

### Goals
- Make step ordering survive DCP context compression by adding explicit ordering constraints
- Embed the handoff note template directly in the `org_session_end` step so it cannot be separated from the action
- Add a forge precondition check that prevents releasing reservations while workers are active
- Keep the prompt concise -- hardening should add minimal token overhead

### Non-Goals
- Changing the `org_session_end` MCP tool itself (the tool is fine; the prompt is the issue)
- Adding runtime enforcement of step ordering (this is a prompt-level fix, not a code-level fix)
- Restructuring other commands for DCP resilience (that's a separate effort per command)
- Testing DCP compression behavior (compression is non-deterministic and not controllable by tests; structural content assertions verify the hardening properties, while DCP behavior is accepted as a known limitation per R2)

## Decisions

### D1: Inline the template into the tool call step

**Decision**: Move the handoff note template from a separate "Handoff Note Template" section into the `org_session_end` step itself, formatted as a code block showing the expected argument structure.

**Rationale**: When the template is a separate section, compression can drop it while keeping the workflow. When it's part of the step instruction, dropping it means dropping the step -- which is much harder for compression to justify since the step contains a tool call. This aligns with the Autonomous Collaboration principle: the artifact (handoff note) is fully described at the point of creation.

### D2: Add explicit ordering constraint language

**Decision**: Add "Steps MUST execute in this exact order -- do not reorder or parallelize" at the top of the workflow section, and add sequencing rationale to each step.

**Rationale**: Numbered lists are not strong enough ordering signals under compression. Explicit MUST language with rationale for each dependency makes reordering require actively contradicting a stated constraint.

### D3: Forge precondition as SHOULD, not MUST

**Decision**: The forge worker check is a SHOULD ("Check for active forge workers before proceeding") rather than a MUST hard gate.

**Rationale**: Per the Composability First principle, the handoff command must remain functional even when forge tools are unavailable. A MUST gate would break handoff in environments where forge is not configured. The check uses `forge_status` which is an existing MCP tool, maintaining artifact-based coordination.

### D4: Single-file change

**Decision**: All hardening changes are contained within `handoff.md`. No changes to Go source, MCP tools, or other embedded assets.

**Rationale**: This is a prompt-level fix for a prompt-level vulnerability. The underlying tools (`comms_release_all`, `org_session_end`, `forge_status`) function correctly; the issue is that the prompt instructions degrade under compression.

## Risks / Trade-offs

### R1: Prompt length increase
Adding ordering constraints, inline templates, and precondition checks increases the prompt from ~25 lines to ~45-50 lines. This adds token overhead to every session that loads the command. Accepted because the reliability improvement outweighs the marginal token cost.

### R2: Cannot fully prevent compression reordering
No prompt structure can guarantee preservation under aggressive compression. The hardening reduces the probability of step reordering and template loss but does not eliminate it. This is an inherent limitation of prompt-based workflows.

### R3: Forge precondition may produce false positives
If `forge_status` returns stale data (e.g., a worker that crashed without cleanup), the agent may hesitate to proceed with handoff. The SHOULD (not MUST) designation means the agent can proceed with a warning rather than blocking entirely.
