## Why

The parent change `delegation-skill-enforcement`
(unbound-force/unbound-force#787) extended `uf_delegate_command`
with a delegation parameter contract --
`task_classification` (`deterministic` | `judgment`), `skill`
(existence-verified skill name or repo-relative script path),
and `skill_absent_reason` -- enforced fail-closed at the tool
boundary, before any child session spawns.

The `/uf.unleash` Swarm parallel path delegates `[P]` tasks
through `replicator_forge_spawn_subtask`. On that path,
classification and skill references live only inside the
spawn-prompt prose: a template-level requirement in the
command file plus divisor-guard review flagging. Nothing at
the tool level validates them. This is the exact prose-only
posture the parent change closed on `uf_delegate_command` --
agent compliance with no detection when a spawn skips the
reference (post-compaction drift, a hurried re-run, or a
future command repeating the defect).

Tool-level validation on the Swarm path is Replicator-owned
and was recorded as a design Non-Goal in the parent change;
this change is its tracked closure
([unbound-force/replicator#143](https://github.com/unbound-force/replicator/issues/143)).

## What Changes

Add equivalent parameter-level validation to
`forge_spawn_subtask`, mirroring `uf_delegate_command`'s
contract:

1. **New parameters** on the `forge_spawn_subtask` MCP tool:
   - `task_classification` (`deterministic` | `judgment`) --
     REQUIRED on every call. Every spawn is a freeform task
     delegation (the analog of the parent's prompt-only
     delegation), so requiring classification closes the
     silent-omission loophole: an unclassified spawn fails
     with input validation, not a silent pass-through.
   - `skill` (min 1, max 256 chars) -- a skill name resolved
     against the project's skill directories
     (`.opencode/skills/<name>/SKILL.md`, then
     `.opencode/skill/<name>/SKILL.md`), or a repo-relative
     script path (any value containing `/`), existence-
     verified before the subtask is marked ready to spawn.
   - `skill_absent_reason` (min 1, max 512 chars) -- the
     recorded justification when deterministic work has no
     applicable repo skill or script.
   - `project_path` (optional, defaults to `.`) -- the
     filesystem anchor for reference resolution, matching the
     `project_path` convention used by comms and worktree
     tools.

2. **Fail-closed validation rules** (all evaluated before the
   success envelope is returned):
   - `deterministic` classification with no `skill` and no
     `skill_absent_reason` fails with
     `SKILL_REFERENCE_REQUIRED`.
   - An unresolvable `skill` reference fails with
     `SKILL_NOT_FOUND`, naming the searched locations.
   - `skill` and `skill_absent_reason` are mutually
     exclusive.
   - Either field supplied with `judgment` classification
     fails input validation (the rendered directive asserts
     deterministic work).
   - `task_classification` values outside the enum fail input
     validation; empty strings are invalid input for all three
     fields (absence is expressed by omitting the field).

3. **Tool-rendered skill directive**: the spawn result
   envelope gains a `skill_directive` field -- a fixed,
   tool-rendered block instructing the worker to load the
   skill or run the script, never agent-reconstructed.
   `forge_subtask_prompt` gains an optional `skill` parameter
   that renders the same fixed directive into the composed
   worker prompt.

4. **Audit record**: the spawn result envelope gains
   `task_classification`, `skill` (the resolved reference),
   and `skill_absent_reason` (when present) -- additive,
   machine-parseable fields reviewers audit without prompt
   archaeology.

5. **Embedded agent-kit call shapes updated**: the
   coordinator agent, `/forge` command, and forge-coordination
   skill documents that demonstrate
   `forge_spawn_subtask(...)` calls are updated to carry the
   new parameters.

6. **Cross-repo coordination**: file an issue in
   unbound-force/unbound-force to update the `uf.unleash.md`
   Swarm-path spawn calls to pass the new parameters (the
   template already requires classification + skill reference
   in prose per the parent change -- this moves them into
   parameters).

## Capabilities

### New Capabilities

- `spawn-delegation-contract`: parameter-level task
  classification and skill/script reference validation on the
  Swarm spawn path, fail-closed before a subtask is marked
  ready to spawn, with a durable audit record in the spawn
  result envelope.

### Modified Capabilities

- `forge_spawn_subtask`: input contract extended with
  `task_classification` (required), `skill`,
  `skill_absent_reason`, and `project_path`; result envelope
  extended with classification, resolved reference, absence
  reason, and the tool-rendered `skill_directive`. BREAKING
  for callers that omit `task_classification`.
- `forge_subtask_prompt`: optional `skill` parameter renders
  the fixed skill directive into the composed worker prompt.
- Embedded agent kit (coordinator agent, `/forge` command,
  forge-coordination skill): documented spawn call shapes
  carry the new parameters.

### Removed Capabilities

- None.

## Impact

- **Files**:
  - `internal/forge/spawn.go` -- validation, reference
    resolution, directive rendering, sentinel errors
  - `internal/tools/forge/tools.go` -- tool schemas and
    parameter plumbing for `forge_spawn_subtask` and
    `forge_subtask_prompt`
  - `internal/forge/spawn_test.go` -- unit coverage for every
    validation rule and resolution path
  - `internal/tools/forge/tools_test.go` -- tool-layer schema,
    envelope, and fail-closed coverage
  - `internal/agentkit/content/` -- coordinator agent, forge
    command, forge-coordination skill call-shape updates
  - `internal/agentkit/agentkit_test.go` -- contract marker test
    pinning the delegation contract in the embedded kit
  - `.opencode/skills/forge-coordination/SKILL.md` and
    `.opencode/commands/forge.md` -- repo-level mirrors kept in
    sync with the embedded kit
  - `docs/tools.md` -- regenerated via `replicator docs`
- **Breaking change**: existing callers that omit
  `task_classification` fail closed with a remediation-carrying
  error. Internal agent-kit call shapes are updated in the
  same change; hero-repo templates (`uf.unleash.md`) are
  updated through the cross-repo coordination issue. The
  failure message names the missing parameter and accepted
  values, so a version-skewed caller is self-describing.
- **No semantic classification**: the tool validates the
  declared answer; deciding whether work is deterministic
  remains the delegating agent's judgment (parent design D1).
- **No enforcement against user-level skills**: the rule
  binds delegated mechanical work to REPO tooling;
  `~/.config/opencode` skills are out of scope (parent
  Non-Goal, mirrored).
- **Parity**: `forge_spawn_subtask` has no TypeScript fixture
  in `test/parity/fixtures/` (only `forge_decompose`,
  `forge_init`, `forge_status`, `forge_worktree_list` are
  covered), so the additive envelope fields introduce no
  parity break.
- **Risk**: misclassification evasion (labeling judgment work
  to skip the requirement) is an explicit, recorded claim the
  guard layer already audits against the task text -- the
  same layered-mitigation posture the parent accepted.

## Constitution Alignment

Assessed against the Replicator constitution
(`.specify/memory/constitution.md`), which extends the
Unbound Force org constitution v1.1.0.

### I. Autonomous Collaboration

**Assessment**: PASS

The contract moves a coordination rule from prose (template
requirement + review flagging) into the tool boundary, making
the spawn result a self-describing artifact: classification,
resolved reference, absence reason, and the rendered directive
are machine-parseable fields any consumer can audit without
prompt archaeology. The tool remains callable independently;
no runtime coupling is introduced.

### II. Composability First

**Assessment**: PASS

The binary remains standalone: reference resolution is a pure
filesystem existence check against the project's own skill
directories, with no registry, manifest, or external service.
Skill resolution degrades to a structured, greppable
`SKILL_NOT_FOUND` error rather than crashing. No new
dependencies. The required-classification breaking change is
coordinated with the unbound-force template update through a
tracked cross-repo issue, mirroring the parent change's
downstream-lag mitigation.

### III. Observable Quality

**Assessment**: PASS

Every validation failure is a structured, greppable error
code (`SKILL_REFERENCE_REQUIRED`, `SKILL_NOT_FOUND`) with
remediation text, surfaced through the existing MCP
error-content mechanism. The success envelope records the
classification, resolved reference, and justification --
durable, machine-parseable provenance for the spawn decision.
The fixed directive wording is covered by golden-string tests.

### IV. Testability

**Assessment**: PASS

Validation rules are pure-function unit tests; reference
resolution uses `t.TempDir()` with synthetic
`.opencode/skills/<name>/SKILL.md` and script-file fixtures --
no external services, no network, no shared state. Every rule
in this proposal has a corresponding test scenario in the
specification delta.
