<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file — parallel
  workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Validation core (internal/forge)

All tasks in this group modify `internal/forge/spawn.go` and
`internal/forge/spawn_test.go` (or add a sibling test file in
the same package) -- no parallel execution. Follow
red-green-refactor: write the failing test first, then the
minimal implementation.

- [x] 1.1 Introduce the `SpawnParams` struct (bead_id, epic_id,
  subtask_title, files, subtask_description, shared_context,
  task_classification, skill, skill_absent_reason,
  project_path) and refactor `SpawnSubtask` to take it.
  Port the existing `TestSpawnSubtask*` tests to the struct
  call shape; existing required-field behavior (missing
  bead_id/epic_id/title fails) stays green. (design decision
  D5)
  **Files**: `internal/forge/spawn.go`,
  `internal/forge/spawn_test.go`
- [x] 1.2 Structural validation with tests:
  `task_classification` required and enum-checked
  (`deterministic` | `judgment`); `skill` and
  `skill_absent_reason` mutually exclusive; either field
  rejected with `judgment` classification; length bounds
  (`skill` 1-256, `skill_absent_reason` 1-512) with empty
  strings invalid. Error text names the parameter and
  accepted values. (design decision D1)
  **Files**: `internal/forge/spawn.go`,
  `internal/forge/spawn_test.go`
- [x] 1.3 Deterministic-requires-reference-or-reason with
  tests: a `deterministic`-classified call with neither
  `skill` nor `skill_absent_reason` fails closed via the
  `ErrSkillReferenceRequired` sentinel whose message leads
  with `SKILL_REFERENCE_REQUIRED` and carries remediation
  text; a deterministic call with a reason (no reference)
  passes. Verify sentinel matching with `errors.Is` through
  wrapping. (design decision D2)
  **Files**: `internal/forge/spawn.go`,
  `internal/forge/spawn_test.go`
- [x] 1.4 `ResolveSkillRef(projectPath, ref)` with tests using
  `t.TempDir()`: skill name resolves
  `.opencode/skills/<name>/SKILL.md` then
  `.opencode/skill/<name>/SKILL.md` (first root wins when both
  exist); `/`-containing values resolve as repo-relative
  paths; absolute paths and `..` elements are rejected;
  unresolvable references fail via `ErrSkillNotFound`
  (`SKILL_NOT_FOUND`, naming searched locations). Wire
  resolution into `SpawnSubtask` with `project_path`
  defaulting to `.`. (design decisions D3, D5)
  **Files**: `internal/forge/spawn.go`,
  `internal/forge/spawn_test.go`
- [x] 1.5 Directive and audit envelope with tests: fixed
  `skillDirective` rendering (golden-string test); success
  envelope gains `task_classification`, resolved `skill`
  path, `skill_absent_reason` (verbatim, when present), and
  `skill_directive` (when a reference was supplied); all
  pre-existing envelope fields keep their names and shapes;
  `SubtaskPrompt` renders the same directive when given a
  reference and produces byte-identical output to today when
  no reference is given. (design decisions D4, D5)
  **Files**: `internal/forge/spawn.go`,
  `internal/forge/spawn_test.go`

## 2. Tool layer (internal/tools/forge)

Same file throughout -- no parallel execution.

- [x] 2.1 Update the `forge_spawn_subtask` tool: input schema
  documents `task_classification` (required, enum), `skill`,
  `skill_absent_reason`, and `project_path`; `Execute`
  unmarshals into `forge.SpawnParams` and returns the
  enriched envelope. (design decision D5)
  **Files**: `internal/tools/forge/tools.go`,
  `internal/tools/forge/tools_test.go`
- [x] 2.2 Add the optional `skill` parameter to the
  `forge_subtask_prompt` tool (schema + prompt composition via
  `SubtaskPrompt`). (design decision D4)
  **Files**: `internal/tools/forge/tools.go`,
  `internal/tools/forge/tools_test.go`

## 3. Embedded agent kit and docs

Different files -- parallel-eligible after groups 1-2 land
(the documented call shapes must reflect the final parameter
names).

- [x] 3.1 [P] Update the coordinator agent definition's spawn
  step to show `task_classification` and the
  reference-or-reason rule.
  **Files**: `internal/agentkit/content/agents/coordinator.md`
- [x] 3.2 [P] Update the `/forge` command workflow's spawn
  step (step 5) with the same contract.
  **Files**: `internal/agentkit/content/commands/forge.md`,
  `.opencode/commands/forge.md` (repo-level mirror synced)
- [x] 3.3 [P] Update the forge-coordination skill's
  Coordinator Protocol spawn step with the same contract.
  **Files**:
  `internal/agentkit/content/skills/forge-coordination/SKILL.md`,
  `.opencode/skills/forge-coordination/SKILL.md` (kept
  byte-identical to the embedded copy per the drift test),
  `internal/agentkit/agentkit_test.go` (contract marker test)
- [x] 3.4 Regenerate the MCP tool reference from the updated
  registration: `go run ./cmd/replicator docs` (verify the
  diff shows only the two changed tools).
  **Files**: `docs/tools.md`

## 4. Governance and verification

- [x] 4.1 File the cross-repo issue in unbound-force/unbound-force
  to update the `uf.unleash.md` Swarm-path spawn calls to pass
  `task_classification` / `skill` / `skill_absent_reason` as
  tool parameters (closes the loop opened by the parent
  change's task 5.5; link this change and issue #143). Do not
  modify unbound-force files from this repo. (design decision
  D6)
  Filed as [unbound-force/unbound-force#830](https://github.com/unbound-force/unbound-force/issues/830).
- [x] 4.2 Website documentation gate: file an issue in
  unbound-force/website covering the changed
  `forge_spawn_subtask` / `forge_subtask_prompt` contracts
  (new required parameter, failure codes, envelope fields).
  Filed as [unbound-force/website#329](https://github.com/unbound-force/website/issues/329).
- [x] 4.3 CI parity gate: run `make check` (vet + test) and
  `make check-coverage`; confirm zero failures and no
  coverage-ratchet regression. Confirm `gofmt`/`goimports`
  cleanliness on touched files.
- [x] 4.4 Constitution alignment verification: confirm every
  spec scenario in
  `specs/spawn-delegation-contract/spec.md` has a passing test
  (Principle IV), error strings lead with greppable codes
  (Principle III), and no new dependency or external service
  entered the build (Principle II).
