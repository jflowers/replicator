## Why

The `/handoff` command in `internal/agentkit/content/commands/handoff.md` defines a 5-step session teardown workflow where step ordering is critical: summarize before release, release before sync, sync before session end. Under DCP (Dynamic Context Protocol) context compression, three fragilities emerge:

1. **Step ordering lost** -- compression can flatten or reorder the numbered steps, causing an agent to release reservations before summarizing (losing awareness of held files), or sync before closing cells (persisting stale state).
2. **Handoff note template dropped** -- the structured template (Completed, In Progress, Blocked, Next Steps, Gotchas) is exactly the kind of content compression reduces to "write handoff notes," losing critical categories the next session depends on.
3. **No forge precondition check** -- nothing prevents an agent from invoking `/handoff` mid-forge and releasing active worker reservations.

This is the same class of vulnerability as [unbound-force/unbound-force#346](https://github.com/unbound-force/unbound-force/issues/346), where the `/review-pr` command's confirmation gate was bypassed under compressed context.

Fixes [#50](https://github.com/unbound-force/replicator/issues/50).

## What Changes

Harden `internal/agentkit/content/commands/handoff.md` against DCP context compression by restructuring the prompt to survive lossy summarization.

## Capabilities

### New Capabilities
- `forge-precondition-guard`: Handoff command checks for active forge workers before releasing reservations

### Modified Capabilities
- `handoff-workflow`: Restructured with DCP-resistant ordering constraints, inline template, and mandatory precondition check

### Removed Capabilities
- None

## Impact

- **File**: `internal/agentkit/content/commands/handoff.md` -- single file change
- **Behavioral**: Agents invoking `/handoff` will now verify no forge workers are active before proceeding, and will produce structured handoff notes even under compressed context
- **Embedded asset**: The file is embedded via `agentkit.go` and distributed with the binary -- a rebuild is required to pick up the change
- **Test**: The agentkit embedding test (`agentkit_test.go`) validates file presence. A new structural content test will verify the hardening properties (ordering constraint placement, template co-location, forge precondition presence) on the embedded file

## Constitution Alignment

Assessed against the Replicator constitution (`.specify/memory/constitution.md`).

### I. Autonomous Collaboration

**Assessment**: PASS

This change improves artifact-based collaboration by ensuring handoff notes remain structured and complete under compression. The precondition check uses the existing `forge_status` tool via MCP, maintaining artifact-based coordination rather than introducing runtime coupling.

### II. Composability First

**Assessment**: N/A

This change modifies an embedded prompt file, not a tool or binary capability. The handoff command remains functional whether forge tools are available or not -- the precondition check is a SHOULD, not a hard gate that blocks session teardown.

### III. Observable Quality

**Assessment**: PASS

The structured handoff note template produces consistently formatted output that downstream consumers (the next session's `org_session_start`) can parse reliably. Moving the template inline with the tool call step eliminates the gap between instruction and execution.

### IV. Testability

**Assessment**: PASS

Structural content assertions verify the hardening properties of the embedded markdown (ordering constraint placement, template co-location, forge precondition presence). DCP compression behavior is inherently untestable and is accepted as a known limitation. No external services or mutable state are involved.
