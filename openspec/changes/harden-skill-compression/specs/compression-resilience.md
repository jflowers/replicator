## ADDED Requirements

### Requirement: Critical Safety Section in always-on-guidance

The `always-on-guidance` skill MUST contain a dedicated `## Critical Safety` section positioned before all other rule sections. Safety-critical constraints (e.g., "Never force push to main") MUST appear in this section, not intermixed with stylistic guidance.

#### Scenario: Safety section positioned first
- **GIVEN** the restructured always-on-guidance skill file
- **WHEN** the file content is inspected
- **THEN** the `## Critical Safety` section MUST appear before all other rule sections in the file

_Compression rationale: header-level sections with first-position placement have the highest compression survival rate._

---

### Requirement: Decision Table Format for Forge Criteria

The `forge-global` skill MUST present forge/skip criteria in a single decision table rather than parallel "do/don't" lists. The table MUST include columns for the evaluation signal, the forge action, and the skip action. The table MUST cover all criteria from the original two lists (3 forge criteria, 3 skip criteria).

#### Scenario: Decision table replaces parallel lists
- **GIVEN** the restructured forge-global skill file
- **WHEN** the file content is inspected
- **THEN** the forge/skip criteria MUST be formatted as a markdown table with Signal, Forge, and Skip columns
- **AND** all 6 original criteria (3 forge, 3 skip) MUST be present in the table

_Compression rationale: decision tables survive as a single coherent structure rather than being split into independent fragments._

---

### Requirement: Explicit Temporal Ordering in Protocol Steps

The File Reservation Protocol in `forge-global` MUST use explicit temporal ordering language (e.g., "FIRST", "THEN", "FINALLY") within each step, in addition to numbered step markers.

#### Scenario: Protocol steps contain temporal markers
- **GIVEN** the restructured forge-global skill file
- **WHEN** the File Reservation Protocol steps are inspected
- **THEN** each step MUST contain an explicit temporal marker (FIRST, THEN, or FINALLY)

_Compression rationale: temporal ordering embedded in each step is recoverable from any surviving step, not solely from list position._

---

## MODIFIED Requirements

### Requirement: Tool Usage Priority Order in always-on-guidance

The `hivemind_find` guidance ("Check hivemind_find before solving problems from scratch") MUST be positioned as the first item in the Tool Usage section. Previously: this item was positioned last (position 6) in the Tool Usage list.

#### Scenario: hivemind_find is first in Tool Usage
- **GIVEN** the restructured always-on-guidance skill file
- **WHEN** the Tool Usage section items are inspected
- **THEN** the `hivemind_find` guidance MUST be the first item in the list

_Compression rationale: first-position items have the highest compression survival rate._

---

### Requirement: TTL Parameter Inline in Reservation Protocol

The `ttl_seconds` parameter guidance MUST be inlined into the reservation step (step 1) of the File Reservation Protocol rather than appearing as a separate standalone bullet. Previously: TTL guidance appeared as an independent step 3 ("Set ttl_seconds to auto-release after timeout").

#### Scenario: TTL parameter embedded in reservation step
- **GIVEN** the restructured forge-global skill file
- **WHEN** the File Reservation Protocol is inspected
- **THEN** the first step MUST include `ttl_seconds` as part of the `comms_reserve` call
- **AND** no standalone TTL bullet MUST exist as a separate step

_Compression rationale: parameters inlined into primary actions survive because they are part of the action, not standalone details._

---

### Requirement: List Size Reduction in always-on-guidance

Rule lists in `always-on-guidance` SHOULD contain no more than 3 items each. Lists exceeding 3 items MUST be split into sub-groups under descriptive sub-headers. This applies to Code Quality, Error Handling, and Testing sections. Previously: lists contained 4-6 items under a single section header.

#### Scenario: Rule lists are within size limits
- **GIVEN** the restructured always-on-guidance skill file
- **WHEN** any rule section's sub-lists are inspected
- **THEN** each sub-list MUST contain no more than 3 items
- **AND** the total rule count in each section MUST equal the original count (no rules added or removed)

_Compression rationale: shorter lists (2-3 items) lose a lower proportion of items under compression than longer lists (5-6 items)._

---

## REMOVED Requirements

No requirements are removed. All existing rules retain their semantic meaning; only their structural presentation changes.
