## ADDED Requirements

### Requirement: Role-Scoped Operation Sections

The forge-coordination skill MUST organize operations under explicit role-scoped section headers: "Coordinator-Only Operations" and "Worker-Only Operations."

#### Scenario: Coordinator-only operations are visually separated
- **GIVEN** an agent loads the forge-coordination skill
- **WHEN** the skill content is processed (with or without compression)
- **THEN** coordinator-only operations (e.g., `comms_release_all`) appear under a "Coordinator-Only Operations" section header, not as parenthetical annotations

#### Scenario: Worker-only operations are visually separated
- **GIVEN** an agent loads the forge-coordination skill
- **WHEN** the skill content is processed (with or without compression)
- **THEN** worker-only operations (e.g., `comms_reserve`, file editing) appear under a "Worker-Only Operations" section header

### Requirement: MUST/MUST NOT Rules at Section Top

Each role-scoped section MUST open with its access-control constraints (MUST/MUST NOT rules) as the first content after the section header, before protocol steps.

#### Scenario: Coordinator constraints appear before protocol
- **GIVEN** an agent reads the "Coordinator-Only Operations" section
- **WHEN** the section is parsed top-to-bottom
- **THEN** the constraint "Coordinators MUST NOT reserve files" appears before any protocol steps

#### Scenario: Worker constraints appear before protocol
- **GIVEN** an agent reads the "Worker-Only Operations" section
- **WHEN** the section is parsed top-to-bottom
- **THEN** the constraint "Workers MUST reserve files before editing" appears before any protocol steps

### Requirement: Inline Conflict Resolution

The conflict resolution procedure MUST be inlined in the Worker Protocol at the point where file reservation failure occurs, not in a separate section.

#### Scenario: Reservation failure triggers inline resolution
- **GIVEN** a worker follows the Worker Protocol
- **WHEN** the worker reaches the "Reserve files" step
- **THEN** conflict resolution steps (check holder, negotiate via `comms_send`, escalate to coordinator) are documented immediately following that step, within the same protocol flow

### Requirement: comms_release_all Coordinator-Only Access Control

`comms_release_all` MUST appear exclusively in the "Coordinator-Only Operations" section with a MUST-level access control statement. Workers MUST NOT call `comms_release_all`.

#### Scenario: comms_release_all is scoped to coordinators
- **GIVEN** an agent reads the "Coordinator-Only Operations" section
- **WHEN** the section is parsed
- **THEN** `comms_release_all` appears with a MUST-level access control statement restricting it to coordinators

#### Scenario: Workers are explicitly prohibited from release_all
- **GIVEN** an agent reads the "Worker-Only Operations" section
- **WHEN** the section's MUST/MUST NOT constraints are read
- **THEN** a constraint "Workers MUST NOT call `comms_release_all`" is present

### Requirement: Exclusive Reservation as Default

The documented `comms_reserve` call in the Worker Protocol MUST include `exclusive=true` in the function signature, making exclusive access the default documented pattern.

#### Scenario: Worker copies reserve call from protocol
- **GIVEN** a worker copies the `comms_reserve` call from the skill
- **WHEN** the call is executed as documented
- **THEN** `exclusive=true` is included in the call parameters

## MODIFIED Requirements

### Requirement: File Reservation Rules

Previously: Flat bullet list mixing coordinator and worker rules with `(coordinator only)` parenthetical and `exclusive=true` as an optional parameter.

The file reservation rules MUST be distributed into the appropriate role-scoped sections. The flat "File Reservation Rules" section MUST be replaced by role-specific constraint blocks within "Coordinator-Only Operations" and "Worker-Only Operations."

### Requirement: Conflict Resolution Section

Previously: Standalone "## Conflict Resolution" section at the end of the document.

The conflict resolution procedure MUST be inlined at the point of use in the Worker Protocol. The standalone section MUST be removed.

### Requirement: File Sync

Both copies of the skill file MUST contain identical content:
- `internal/agentkit/content/skills/forge-coordination/SKILL.md`
- `.opencode/skills/forge-coordination/SKILL.md`

#### Scenario: Files stay in sync after change
- **GIVEN** the hardening changes are applied
- **WHEN** both files are compared
- **THEN** they are byte-identical

## REMOVED Requirements

None.
