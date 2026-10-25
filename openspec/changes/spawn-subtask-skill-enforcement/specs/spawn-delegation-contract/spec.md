## ADDED Requirements

### Requirement: Spawn task classification is declared and validated

The `forge_spawn_subtask` tool MUST require a `task_classification` parameter on every call. The value MUST be exactly `deterministic` or `judgment`. A call that omits `task_classification`, supplies an empty string, or supplies any other value MUST fail input validation before any other processing, with an error message naming the parameter and the accepted values. The tool MUST NOT attempt to classify the task itself -- it validates the declared answer only.

#### Scenario: Omitted classification fails

- **GIVEN** a caller invokes `forge_spawn_subtask` with valid `bead_id`, `epic_id`, `subtask_title`, and `files`, but no `task_classification`
- **WHEN** the call is executed
- **THEN** the tool MUST return an input validation error naming `task_classification` as required and listing `deterministic` and `judgment` as the accepted values
- **AND** the subtask MUST NOT be marked ready to spawn

#### Scenario: Non-enum classification fails

- **GIVEN** a caller supplies `task_classification` with a value other than `deterministic` or `judgment`
- **WHEN** the call is executed
- **THEN** the tool MUST return an input validation error listing the accepted values

#### Scenario: Judgment-classified spawn passes without reference

- **GIVEN** a caller supplies `task_classification` of `judgment` with no `skill` and no `skill_absent_reason`
- **WHEN** the call is executed with otherwise valid arguments
- **THEN** the tool MUST return a success envelope with `task_classification` recorded as `judgment`

### Requirement: Deterministic spawns require a reference or a recorded reason

A `deterministic`-classified spawn MUST supply either a `skill` reference or a `skill_absent_reason`. A deterministic-classified call with neither MUST fail closed with a `SKILL_REFERENCE_REQUIRED` error that carries remediation text (supply a skill/script reference, or record why none applies). The check MUST run before the subtask is marked ready to spawn.

#### Scenario: Deterministic without reference or reason fails closed

- **GIVEN** a caller supplies `task_classification` of `deterministic` with no `skill` and no `skill_absent_reason`
- **WHEN** the call is executed
- **THEN** the tool MUST return an error containing the code `SKILL_REFERENCE_REQUIRED`
- **AND** the error message MUST state the remediation (supply a skill or skill_absent_reason)
- **AND** the subtask MUST NOT be marked ready to spawn

#### Scenario: Deterministic with absence reason passes

- **GIVEN** a caller supplies `task_classification` of `deterministic` and a non-empty `skill_absent_reason`, with no `skill`
- **WHEN** the call is executed with otherwise valid arguments
- **THEN** the tool MUST return a success envelope recording the `skill_absent_reason` verbatim

### Requirement: Supplied references are existence-verified before spawn

When `skill` is supplied, the tool MUST verify the reference exists on disk before returning a success envelope, resolving against the project root given by `project_path` (default `.`). Resolution rules are defined by the skill-name and repo-relative-path requirements below. An unresolvable or rejected reference MUST fail closed with a `SKILL_NOT_FOUND` error naming the searched locations.

#### Scenario: Unknown reference fails closed

- **GIVEN** a project with no skill or script matching the supplied `skill` value
- **WHEN** the call is executed
- **THEN** the tool MUST return an error containing the code `SKILL_NOT_FOUND` that names every searched location
- **AND** the subtask MUST NOT be marked ready to spawn

### Requirement: Skill names resolve against the project skill directories

A `skill` value without `/` MUST resolve as a skill name: `.opencode/skills/<name>/SKILL.md` first, then `.opencode/skill/<name>/SKILL.md`. A name present in both roots MUST resolve to the first root.

#### Scenario: Skill name resolves through the primary root

- **GIVEN** a project whose `.opencode/skills/deploy/SKILL.md` exists
- **WHEN** a deterministic-classified spawn supplies `skill` of `deploy`
- **THEN** the tool MUST resolve the reference to `.opencode/skills/deploy/SKILL.md` and return a success envelope carrying the resolved path

#### Scenario: Skill name falls back to the secondary root

- **GIVEN** a project where `.opencode/skills/<name>/SKILL.md` does not exist but `.opencode/skill/<name>/SKILL.md` does
- **WHEN** a spawn supplies `skill` of `<name>`
- **THEN** the tool MUST resolve the reference to `.opencode/skill/<name>/SKILL.md`

### Requirement: Repo-relative script paths resolve with a traversal guard

A `skill` value containing `/` MUST resolve as a repo-relative path. Absolute paths and paths containing a `..` path element remaining after path cleaning MUST be rejected as traversal attempts.

#### Scenario: Repo-relative script path resolves

- **GIVEN** a project containing `scripts/lint.sh`
- **WHEN** a spawn supplies `skill` of `scripts/lint.sh`
- **THEN** the tool MUST resolve the reference to `scripts/lint.sh` and verify its existence

#### Scenario: Traversal attempts are rejected

- **GIVEN** a spawn supplying `skill` of `../secrets/env.sh` or an absolute path
- **WHEN** the call is executed
- **THEN** the tool MUST return a `SKILL_NOT_FOUND` error and MUST NOT touch any path outside the project root

### Requirement: Reference and reason are mutually exclusive and deterministic-only

`skill` and `skill_absent_reason` MUST NOT be supplied together. Either field supplied with a `judgment` classification MUST fail input validation, because the rendered directive asserts deterministic work. Empty strings MUST be invalid input for `skill` (bounds 1-256 chars) and `skill_absent_reason` (bounds 1-512 chars); absence MUST be expressed by omitting the field.

#### Scenario: Both reference and reason fail

- **GIVEN** a deterministic-classified spawn supplying both `skill` and `skill_absent_reason`
- **WHEN** the call is executed
- **THEN** the tool MUST return an input validation error stating the fields are mutually exclusive

#### Scenario: Reference with judgment classification fails

- **GIVEN** a spawn supplying `task_classification` of `judgment` and a `skill` value
- **WHEN** the call is executed
- **THEN** the tool MUST return an input validation error stating `skill` is valid only with `deterministic` classification

#### Scenario: Empty-string fields are invalid

- **GIVEN** a spawn supplying `skill` or `skill_absent_reason` as an empty string
- **WHEN** the call is executed
- **THEN** the tool MUST return an input validation error stating that absence is expressed by omitting the field

### Requirement: Spawn envelope records the delegation facts

On success, the `forge_spawn_subtask` result envelope MUST include `task_classification`; the resolved `skill` reference (the path resolved on disk) when one was supplied; and `skill_absent_reason` verbatim when one was supplied. These fields are additive -- existing envelope fields MUST retain their names and shapes.

#### Scenario: Envelope carries classification and resolved reference

- **GIVEN** a deterministic-classified spawn with a verified `skill` reference
- **WHEN** the tool returns its success envelope
- **THEN** the envelope MUST contain `task_classification` of `deterministic`, the resolved skill path, and no `skill_absent_reason` field

#### Scenario: Envelope carries absence reason

- **GIVEN** a deterministic-classified spawn with a `skill_absent_reason`
- **WHEN** the tool returns its success envelope
- **THEN** the envelope MUST contain the reason verbatim and no resolved skill path

### Requirement: Skill directive is tool-rendered

When a verified `skill` reference exists, the tool MUST render a fixed skill-directive block instructing the worker to load the named skill or run the referenced script and follow its full procedure without substitution. The directive MUST be produced by the tool, not reconstructed by the calling agent.

#### Scenario: Envelope carries the rendered directive

- **GIVEN** a deterministic-classified spawn with a verified `skill` reference
- **WHEN** the tool returns its success envelope
- **THEN** the envelope MUST contain a `skill_directive` field whose text names the skill or script and instructs the worker to execute its full procedure

#### Scenario: Subtask prompt embeds the directive

- **GIVEN** a `forge_subtask_prompt` call supplying a `skill` reference
- **WHEN** the prompt is composed
- **THEN** the composed prompt MUST contain the same fixed directive block
- **AND** a spawn without a skill reference MUST NOT contain the directive block

### Requirement: forge_spawn_subtask input contract

The `forge_spawn_subtask` tool input contract MUST accept the existing parameters (`bead_id`, `epic_id`, `subtask_title`, `files`, `subtask_description`, `shared_context`) plus `task_classification` (required), `skill`, `skill_absent_reason`, and `project_path` (optional, default `.`). The JSON schema MUST document the enum, bounds, and required status of each new parameter.

#### Scenario: Schema documents the new contract

- **GIVEN** the registered `forge_spawn_subtask` tool
- **WHEN** its input schema is inspected
- **THEN** the schema MUST list `task_classification` as required with enum values `deterministic` and `judgment`
- **AND** MUST document `skill`, `skill_absent_reason`, and `project_path`
- **AND** the previously required inputs (`bead_id`, `epic_id`, `subtask_title`, `files`) MUST remain required with unchanged types

### Requirement: Embedded agent kit demonstrates the parameter contract

The embedded coordinator agent definition, `/forge` command, and forge-coordination skill MUST show `forge_spawn_subtask` call shapes that include `task_classification` and, for deterministic work, a `skill` reference or `skill_absent_reason`. Previously the documented call shapes were `forge_spawn_subtask(bead_id, epic_id, subtask_title, files)` with no classification parameters.

#### Scenario: Coordinator protocol carries the contract

- **GIVEN** the embedded coordinator agent or forge command content
- **WHEN** the spawn step is read
- **THEN** the documented call MUST include `task_classification` and the reference-or-reason rule
