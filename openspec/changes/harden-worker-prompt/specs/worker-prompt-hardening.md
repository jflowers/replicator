## ADDED Requirements

### Requirement: Reservation Failure Recovery

The worker MUST include an explicit recovery instruction for reservation failures. If `comms_reserve` fails, the reservation expires (TTL), or the reservation is released by the coordinator, the worker MUST STOP all work and report the failure to the coordinator via `comms_send`.

#### Scenario: Initial reservation acquisition fails
- **GIVEN** a worker has initialized comms and is attempting to reserve files
- **WHEN** `comms_reserve` returns an error (files already reserved by another worker)
- **THEN** the worker MUST stop, send a message to the coordinator via `comms_send` describing which files could not be reserved, and NOT proceed to implementation

#### Scenario: Worker discovers reservation no longer held
- **GIVEN** a worker has reserved files and is implementing changes
- **WHEN** the coordinator notifies the worker via `comms_send` that its reservation has been released, or the worker discovers the loss through a subsequent `comms_reserve` call
- **THEN** the worker MUST stop editing files immediately and report the situation to the coordinator via `comms_send`

Note: Detection depends on external notification (coordinator message) or a subsequent reserve/release call. There is no proactive TTL-expiry notification mechanism at the prompt level.

### Requirement: Inline Constraint Co-location

Critical behavioral constraints MUST be co-located with the checklist step they govern, not in a separate section. Each checklist step MUST contain both the action and its boundary condition as an atomic unit.

#### Scenario: Constraints are co-located with actions
- **GIVEN** the restructured worker.md content is read from the embedded FS
- **WHEN** the checklist step containing `comms_reserve` is examined
- **THEN** the same step text contains at least one MUST or NEVER constraint keyword about file editing

### Requirement: Mandatory Progress Reporting

Progress reporting via `forge_progress` MUST be described as a mandatory structural step, not an optional reporting activity. The checklist step MUST use imperative language ("MUST report") rather than descriptive language.

#### Scenario: Progress step uses mandatory language
- **GIVEN** the restructured worker.md content is read from the embedded FS
- **WHEN** the checklist step containing `forge_progress` is examined
- **THEN** the step text contains the word "MUST"

### Requirement: File Conciseness

The restructured worker.md MUST remain under 35 lines (including frontmatter) to avoid creating a new compression vulnerability through increased prompt length.

#### Scenario: File stays concise
- **GIVEN** the restructured worker.md
- **WHEN** the total line count is measured
- **THEN** the file MUST be under 35 lines (including frontmatter)

## MODIFIED Requirements

### Requirement: Checklist Structure

Previously: The checklist was a 7-step numbered list (lines 14-20) with constraints in a separate "Constraints" section (lines 22-27).

The checklist MUST be a self-contained numbered list where each step includes both the action and its associated constraints inline. The separate "Constraints" section SHALL be removed. Redundant constraint phrasings ("Only edit files you have reserved" and "Never modify files outside your assignment") MUST be consolidated into a single clear statement at the point of action.

#### Scenario: File reservation constraint is inline
- **GIVEN** the worker prompt has been restructured
- **WHEN** a reader examines step 3 (`comms_reserve`)
- **THEN** the step includes an explicit "NEVER edit unreserved files" constraint in the same text block

#### Scenario: Separate Constraints section removed
- **GIVEN** the worker prompt has been restructured
- **WHEN** the full prompt content is examined
- **THEN** there is no separate "## Constraints" heading or section

### Requirement: Learning Storage Framing

Previously: Step 6 read "hivemind_store — store any learnings discovered" and constraint "Store learnings for future agents" was in a separate section.

The learning storage step MUST use mandatory language and SHOULD include guidance on what constitutes a useful learning (gotchas, patterns, decisions).

#### Scenario: Learning storage reads as mandatory
- **GIVEN** the restructured worker prompt
- **WHEN** a reader examines the learning storage step
- **THEN** the step uses "MUST" language and is not dismissable as optional

## REMOVED Requirements

### Requirement: Separate Constraints Section

The standalone "## Constraints" section with flat bullet points is removed. All constraint content is migrated into the checklist steps. This removal is the core mechanism of the hardening — eliminating the structurally vulnerable section that compressors are likely to summarize away.
