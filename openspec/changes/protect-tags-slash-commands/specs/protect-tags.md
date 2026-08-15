## ADDED Requirements

### Requirement: Protect tags on slash command guardrails

All safety rules, invariant sections, and MUST/NEVER constraints in
slash command files MUST be wrapped in `<protect>` / `</protect>` tags.

#### Scenario: forge.md Critical Invariants survive compression
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** a long conversation triggers DCP compression on forge.md content
- **THEN** the Critical Invariants section (review gate rules, mandatory forge creation, coordinator/worker separation) MUST be preserved verbatim

#### Scenario: forge.md Rules survive compression
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP compresses forge.md content
- **THEN** the Rules section MUST be preserved verbatim

### Requirement: Protect tags on slash command workflows

All numbered workflow step sequences in slash command files MUST be
wrapped in `<protect>` / `</protect>` tags.

#### Scenario: forge.md Workflow steps survive compression
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP compresses forge.md content
- **THEN** all 8 workflow steps (initialize through complete) MUST be preserved verbatim with step ordering intact

#### Scenario: handoff.md ordered workflow survives compression
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP compresses handoff.md content
- **THEN** the ordering constraint ("Steps MUST execute in this exact order") and all 6 steps (0-5) with their dependency explanations MUST be preserved verbatim

#### Scenario: inbox.md workflow survives compression
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP compresses inbox.md content
- **THEN** the 3-step workflow (inbox, read, ack) MUST be preserved verbatim

#### Scenario: forge-status.md workflow survives compression
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP compresses forge-status.md content
- **THEN** the 3-step workflow (forge_status, comms_inbox, org_cells) MUST be preserved verbatim

### Requirement: Protect tags on exit conditions and resume instructions

Session end procedures and handoff note structures MUST be wrapped in
`<protect>` / `</protect>` tags.

#### Scenario: handoff.md session end and handoff structure survive compression
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP compresses handoff.md content
- **THEN** step 5 (org_session_end with handoff note structure) MUST be preserved verbatim

### Requirement: Protect tags on org tool reference and session instructions

The Common Actions tool reference and Sessions section in org.md MUST be
wrapped in `<protect>` / `</protect>` tags.

#### Scenario: org.md Common Actions survive compression
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP compresses org.md content
- **THEN** the Common Actions tool reference (org_ready, org_cells, org_create, org_update, org_close, org_start) MUST be preserved verbatim

#### Scenario: org.md Sessions survive compression
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP compresses org.md content
- **THEN** the Sessions section (org_session_start, org_session_end) MUST be preserved verbatim

### Requirement: DCP protectTags configuration

The project configuration (`opencode.json`) MUST include
`"protectTags": true` at the top level to enable DCP recognition of
`<protect>` tags.

#### Scenario: protectTags enabled in project config
- **GIVEN** the opencode.json configuration file
- **WHEN** OpenCode loads the project configuration
- **THEN** `protectTags` MUST be set to `true`

### Requirement: Unprotected content remains compressible

Descriptive text, examples, section headings, and advisory guidance
SHOULD NOT be wrapped in `<protect>` tags to preserve DCP compression
effectiveness.

#### Scenario: forge-status.md Interpreting Results remains compressible
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP evaluates forge-status.md for compression
- **THEN** the Interpreting Results and Quick Health Check sections SHOULD be available for compression

#### Scenario: org.md Usage and Epics remain compressible
- **GIVEN** DCP is active with `protectTags: true`
- **WHEN** DCP evaluates org.md for compression
- **THEN** the Usage and Epics sections SHOULD be available for compression

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
