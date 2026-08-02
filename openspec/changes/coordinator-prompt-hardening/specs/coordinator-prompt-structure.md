## ADDED Requirements

### FR-001: Identity-first opening statement

The coordinator prompt MUST begin (after YAML front matter) with an identity statement that embeds critical constraints inline. The opening paragraph MUST include the coordinator's role AND at least one key prohibition (e.g., NEVER reserves files).

#### Scenario: Opening paragraph contains prohibition

- **GIVEN** the coordinator prompt after YAML front matter
- **WHEN** the first paragraph is extracted
- **THEN** it MUST contain the uppercase keyword "NEVER" and the phrase "reserve files" (or equivalent prohibition)

### FR-002: Dedicated critical constraints section

The coordinator prompt MUST contain a section titled "Critical Constraints" (or equivalent strong header) that appears BEFORE any workflow or protocol section. This section MUST contain all negative constraints and mandatory ordering rules using uppercase RFC 2119 keywords (MUST, NEVER, ALWAYS).

#### Scenario: Critical constraints appear in first half of file

- **GIVEN** the coordinator prompt has N total lines after YAML front matter
- **WHEN** only the first floor(N/2) lines after front matter are retained
- **THEN** the retained lines MUST contain: (1) the file reservation prohibition (`NEVER` + `reserve`), (2) the review-before-complete ordering (`forge_review` + `forge_complete`), and (3) the code editing prohibition (`NEVER` + `edit code`) — these are the three critical constraints that must survive truncation

Note: The 50% threshold is a conservative structural heuristic, not a measured DCP compression ratio. The defense-in-depth rationale (redundancy with the forge coordination skill) mitigates the inherent non-determinism of LLM-based compression.

### FR-003: Explicit review-before-complete ordering

The coordinator prompt MUST state that `forge_review` MUST be called for every worker completion BEFORE calling `forge_complete`. This ordering constraint MUST appear in both the critical constraints section and the numbered protocol. This codifies an ordering that was previously implicit in the forge coordination skill (steps 8→9 in the Coordinator Protocol) but absent from the coordinator prompt itself.

#### Scenario: Ordering constraint present in file structure

- **GIVEN** the restructured coordinator prompt
- **WHEN** the file content is searched
- **THEN** it MUST contain both the strings `forge_review` and `forge_complete` with explicit ordering language (e.g., "BEFORE", "prior to", sequential numbering where review precedes complete)

### FR-004: Uppercase severity keywords

All behavioral constraints in the coordinator prompt MUST use uppercase RFC 2119 keywords (MUST, MUST NOT, NEVER, ALWAYS, SHALL) for severity signaling.

#### Scenario: Constraint keyword casing

- **GIVEN** the coordinator prompt contains behavioral constraints
- **WHEN** lines containing prohibitions or mandatory behaviors are extracted
- **THEN** each such line MUST contain at least one uppercase RFC 2119 keyword (NEVER, MUST, ALWAYS) rather than lowercase equivalents

### FR-005: YAML front matter preservation

The restructured coordinator prompt MUST preserve the original YAML front matter values.

#### Scenario: Front matter parity

- **GIVEN** the restructured coordinator prompt
- **WHEN** the YAML front matter is parsed
- **THEN** it MUST contain `name: coordinator`, a `description` field, and `mode: subagent`

### FR-006: Behavioral rule presence markers

Each of the 6 original behavioral rules MUST be verifiable by the presence of specific string patterns in the restructured file:

| Rule | Required Pattern |
|------|-----------------|
| comms init | `comms_init` |
| no file reservation | `NEVER` + `reserve` |
| review completions | `forge_review` |
| store learnings | `hivemind_store` |
| check inbox | `comms_inbox` |
| broadcast context | `forge_broadcast` |

#### Scenario: Automated parity check

- **GIVEN** the restructured coordinator prompt
- **WHEN** file content is searched for each required pattern in the table above
- **THEN** all 6 patterns MUST be found in the file

## MODIFIED Requirements

### FR-007: Coordinator behavioral rules restructuring

Previously: Six unordered bullet points in a single "Rules" section with lowercase constraint language ("Always", "Never").

The coordinator's behavioral rules MUST be restructured into:
1. A "Critical Constraints" section containing prohibitions and mandatory orderings (positioned first)
2. A numbered "Protocol" section containing the ordered workflow steps

The 6 original behavioral rules MUST remain with equivalent semantics. Additionally, the implicit `forge_review` → `forge_complete` ordering (present in the forge coordination skill but absent from the coordinator prompt) is made explicit — this is a codification of existing behavior, not a net-new rule.

#### Scenario: Behavioral parity verification

- **GIVEN** the original coordinator prompt contains 6 rules (comms init, no file reservation, review completions, store learnings, check inbox, broadcast context)
- **WHEN** the restructured prompt is compared to the original
- **THEN** all 6 rules MUST still be present (verified by pattern matching per FR-006), and the `forge_review` → `forge_complete` ordering MUST be explicitly stated

### FR-008: Coordinator prompt structure ordering

Previously: Single flat structure (front matter → header → description → rules → tools).

The coordinator prompt MUST follow this section ordering:
1. YAML front matter
2. Identity heading and opening statement (with embedded constraints)
3. Critical Constraints section
4. Protocol section (numbered workflow steps)
5. Available Tools section

#### Scenario: Section ordering validation

- **GIVEN** the restructured coordinator prompt
- **WHEN** section headers (lines starting with `##`) are parsed in document order
- **THEN** the "Critical Constraints" header MUST appear before the "Protocol" header, and both MUST appear before the "Available Tools" header

## REMOVED Requirements

None. No existing requirements are removed by this change.
<!-- scaffolded by uf vdev -->
