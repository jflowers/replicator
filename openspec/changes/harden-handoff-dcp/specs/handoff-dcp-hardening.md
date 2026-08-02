## ADDED Requirements

### Requirement: Handoff step ordering constraint

The `/handoff` command workflow MUST include an explicit ordering constraint statement before the step list. The constraint MUST use RFC 2119 MUST language to prohibit reordering or parallelizing steps.

#### Scenario: Ordering constraint is structurally distinct from step list
- **GIVEN** the hardened `handoff.md` file is read
- **WHEN** the workflow section is examined
- **THEN** the ordering constraint statement MUST appear as a distinct paragraph before the first numbered step in the markdown file

### Requirement: Forge precondition check

The `/handoff` command SHOULD instruct the agent to check for active forge workers before executing `comms_release_all`. If active workers are detected, the agent SHOULD warn the user and request confirmation before proceeding. If the agent has no active forge context (no known `epic_id` or `project_key`), the forge precondition check SHOULD be skipped, same as when forge tools are unavailable.

#### Scenario: Handoff invoked with active forge workers
- **GIVEN** a forge session is active with running workers
- **WHEN** the agent invokes `/handoff`
- **THEN** the agent SHOULD check `forge_status` and warn that active workers will lose their reservations
- **AND** the agent SHOULD request user confirmation before calling `comms_release_all`

#### Scenario: Handoff invoked with no forge activity
- **GIVEN** no forge session is active
- **WHEN** the agent invokes `/handoff`
- **THEN** the agent SHOULD proceed through the workflow without blocking on the forge check

#### Scenario: Forge tools unavailable
- **GIVEN** forge tools are not available in the current MCP session
- **WHEN** the agent invokes `/handoff`
- **THEN** the agent MUST proceed with handoff without the forge check (graceful degradation)

### Requirement: Inline handoff note template

The handoff note template MUST be embedded directly within the `org_session_end` workflow step rather than in a separate section. The template MUST specify all five categories: Completed, In Progress, Blocked, Next Steps, Gotchas.

#### Scenario: Handoff note template is co-located with tool call
- **GIVEN** the hardened `handoff.md` file is read
- **WHEN** the `org_session_end` step is examined
- **THEN** the handoff note template (all five categories) MUST appear within the same markdown section as the `org_session_end` tool call instruction

### Requirement: Step dependency rationale

Each workflow step MUST include a brief rationale explaining why it depends on the previous step completing first.

#### Scenario: Each step contains a dependency rationale
- **GIVEN** the hardened `handoff.md` file is read
- **WHEN** each numbered workflow step is examined
- **THEN** each step MUST contain a rationale clause following the tool call instruction that explains why it depends on the previous step

## MODIFIED Requirements

### Requirement: Workflow section structure

Previously: A numbered list of 5 steps followed by a separate "Handoff Note Template" section.

The workflow section MUST now contain:
1. An ordering constraint header
2. An optional forge precondition check
3. The existing 5 steps with inline rationale for each dependency
4. The handoff note template embedded within step 5 (`org_session_end`)

The separate "Handoff Note Template" section MUST be removed.

## REMOVED Requirements

None.
