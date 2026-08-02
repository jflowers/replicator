## ADDED Requirements

### Requirement: Critical Invariants Section

The forge command prompt MUST include a "Critical Invariants" section positioned immediately after the title and before the Workflow section. This section MUST state the non-negotiable ordering constraints and behavioral rules in explicit RFC 2119 language.

#### Scenario: Invariants appear before workflow

- **GIVEN** the forge.md file is loaded by an agent
- **WHEN** the content is parsed or compressed
- **THEN** the Critical Invariants section MUST appear before the Workflow section in document order

#### Scenario: Review-before-complete invariant present

- **GIVEN** the Critical Invariants section exists
- **WHEN** its content is read
- **THEN** it MUST contain a statement that review (step 7) MUST complete before marking work done (step 8)

#### Scenario: skip_review prohibition present

- **GIVEN** the Critical Invariants section exists
- **WHEN** its content is read
- **THEN** it MUST contain a statement that `skip_review: true` MUST NEVER be passed to `forge_complete`

#### Scenario: Review-before-complete constraint has redundant placement

- **GIVEN** the forge.md file
- **WHEN** any single section is removed entirely
- **THEN** the review-before-complete constraint MUST still be present in at least one other location

## MODIFIED Requirements

### Requirement: Workflow step 7 ordering constraint

Step 7 (Review) MUST embed an explicit ordering constraint in its text indicating it MUST be completed before step 8 (Complete). The constraint MUST use explicit language ("MUST", "FIRST", "before step 8") rather than relying on sequential numbering alone.

Previously: "7. Review: `forge_review(task_id, files_touched)` for each completed worker"

#### Scenario: Step 7 text includes ordering constraint

- **GIVEN** the Workflow section of forge.md
- **WHEN** step 7 text is read
- **THEN** it MUST contain explicit language indicating it must be completed before step 8

### Requirement: Rules section ordering

The rule "Review every worker's output before marking complete" MUST be the first item in the Rules section. Rules SHOULD be ordered by criticality (most critical first).

Previously: "Review every worker's output before marking complete" appeared as the 5th of 6 bullets.

#### Scenario: Review rule is first in list

- **GIVEN** the Rules section of forge.md
- **WHEN** the bullet items are read in order
- **THEN** the review-before-complete rule MUST be the first item

### Requirement: Strategy selection inlined with decompose

Strategy selection guidance (including `forge_get_strategy_insights`) MUST be inlined as a sub-item of the decompose step rather than appearing as a separate trailing section.

Previously: Strategy Selection was a standalone section at lines 34-42.

#### Scenario: Strategy selection is part of decompose step

- **GIVEN** the Workflow section of forge.md
- **WHEN** the decompose step is read
- **THEN** it MUST include strategy selection guidance as a sub-item
- **AND** there MUST NOT be a separate "Strategy Selection" section

### Requirement: Error recovery inlined with monitoring

Error recovery guidance MUST be inlined as sub-items of the monitoring step rather than appearing as a separate trailing section.

Previously: Error Recovery was a standalone section at lines 61-67.

#### Scenario: Error recovery is part of monitoring step

- **GIVEN** the Workflow section of forge.md
- **WHEN** the monitoring step is read
- **THEN** it MUST include error recovery guidance as sub-items
- **AND** there MUST NOT be a separate "Error Recovery" section at the end of the file

### Requirement: Completion details inlined with complete step

Completion sub-steps (`forge_complete`, `forge_record_outcome`, `hivemind_store`, `org_sync`) MUST be inlined as sub-items of step 8 (Complete) rather than appearing as a separate "Completion" section.

Previously: Completion was a standalone section at lines 52-59 repeating step 8 with more detail.

#### Scenario: Completion details are part of step 8

- **GIVEN** the Workflow section of forge.md
- **WHEN** step 8 is read
- **THEN** it MUST include completion sub-steps as sub-items
- **AND** there MUST NOT be a separate "Completion" section

## REMOVED Requirements

### Requirement: Standalone Strategy Selection section

Removed as a standalone section. Content is preserved but moved inline with the decompose step (see MODIFIED: Strategy selection inlined with decompose).

### Requirement: Standalone Error Recovery section

Removed as a standalone section. Content is preserved but moved inline with the monitoring step (see MODIFIED: Error recovery inlined with monitoring).

### Requirement: Standalone Completion section

Removed as a standalone section. Content is preserved but moved inline with step 8 (see MODIFIED: Completion details inlined with complete step).
