## Context

The parent change `delegation-skill-enforcement`
(unbound-force/unbound-force#787) closed the prose-only
posture on `uf_delegate_command` with a fail-closed parameter
contract. The Swarm parallel path was explicitly out of scope
there: `/uf.unleash` delegates `[P]` tasks through
`replicator_forge_spawn_subtask`, whose classification and
skill references live only in spawn-prompt prose (template
requirement plus divisor-guard review flagging). Issue #143
tracks the Replicator-owned closure.

Current state of the seam:

- `forge.SpawnSubtask` (`internal/forge/spawn.go`) is a pure
  function taking positional args (`beadID, epicID, title,
  files, description, sharedContext`), validating only that
  the IDs and title are non-empty. No filesystem access, no
  project anchor.
- `forgeSpawnSubtask` (`internal/tools/forge/tools.go:238`)
  registers the MCP tool with a documentation-only JSON
  schema; the registry (`internal/tools/registry`) does not
  auto-validate arguments -- validation happens inside each
  tool's `Execute`.
- Tool errors surface through the MCP server
  (`internal/mcp/server.go`) as `isError` responses with
  `Error: <message>` text content, so greppable codes must
  lead the error string.
- `forge.SubtaskPrompt` (exposed as `forge_subtask_prompt`)
  composes the worker prompt; it is the Swarm path's
  prompt-composition hook.
- No parity fixture covers `forge_spawn_subtask` (fixtures
  exist only for `forge_decompose`, `forge_init`,
  `forge_status`, `forge_worktree_list`), so additive envelope
  fields are parity-safe.
- The embedded agent kit (`internal/agentkit/content/` --
  coordinator agent, `/forge` command, forge-coordination
  skill) documents the old spawn call shape.

Constitution alignment is assessed per principle in
proposal.md; this design is Principle III (Observable
Quality) applied to the spawn boundary, with Principle IV
shaping the test strategy.

## Goals / Non-Goals

### Goals

- A `deterministic`-classified spawn without a verified
  skill/script reference or a recorded absence reason fails
  closed at the tool, before the subtask is marked ready to
  spawn -- zero wasted child sessions.
- Every spawn declares its classification, closing the
  silent-omission loophole (an unclassified spawn is
  evadable-by-omission otherwise).
- Supplied references are existence-verified against the
  project's skill directories or as repo-relative script
  paths, with traversal attempts rejected.
- The classification, resolved reference, and absence reason
  are durable, machine-parseable envelope fields.
- The skill directive rendered for the worker is
  tool-produced (fixed wording), never agent-reconstructed.
- Judgment-classified spawns pass through unchanged -- zero
  false positives by construction.

### Non-Goals

- No semantic classification inside the tool. Deciding
  whether work is deterministic is judgment; the tool
  validates the declared answer (parent D1, mirrored).
- No enforcement against user-level (`~/.config/opencode`)
  skills. The rule binds delegated mechanical work to REPO
  tooling.
- No embedding of the SKILL.md body into the composed prompt
  -- a directive naming the skill, not the document.
- No skill registry or manifest; an existence check on two
  known roots needs none.
- No schema-validation framework changes: the registry stays
  documentation-schema-only; validation lives in the tool's
  `Execute`, as every other tool does today.
- No modification of unbound-force templates in this repo --
  the `uf.unleash.md` update is a cross-repo issue filed from
  this change's tasks.
- No change to the review/complete chain, comms, or org
  layers.

## Decisions

### D1: The caller classifies; the tool enforces

`task_classification` is REQUIRED on every
`forge_spawn_subtask` call. Every spawn is a freeform task
delegation -- the direct analog of the parent's prompt-only
delegation, for which classification is mandatory. Making it
optional here would recreate the exact loophole this change
closes: skip classification, skip the rule.

Structural rules validated before any filesystem work:

- enum membership (only `deterministic` and `judgment`)
- `skill` and `skill_absent_reason` mutually exclusive
- either field valid only with `deterministic` classification
- length bounds (`skill` 1-256, `skill_absent_reason`
  1-512); empty strings are invalid input -- absence is
  expressed by omitting the field

Structural failures return plain input-validation errors in
the existing house style (`task_classification is required
(accepted values: deterministic, judgment)`). The two rules that
need dedicated codes run next and return sentinel-based
envelopes (D2).

Alternatives rejected:

- **Optional classification, rules fire only when
  deterministic is supplied.** Evadable by omission; recreates
  the prose-only posture.
- **Heuristic skill detection in prompt text.** The tool
  cannot reliably recognize a reference in free text (parent
  D1, mirrored).

### D2: Fail-closed rules are sentinel errors with greppable codes

`internal/forge` defines two sentinel errors, wrapped with
call context but `errors.Is`-matchable:

- `ErrSkillReferenceRequired` --
  `SKILL_REFERENCE_REQUIRED: ...` with remediation (supply a
  skill/script reference or record why none applies)
- `ErrSkillNotFound` -- `SKILL_NOT_FOUND: ...` naming the
  searched locations

The MCP server renders tool errors as `Error: <message>` text
in an `isError` response, so the code leads the error string
and is greppable in logs and transcripts. This follows the
remediation-carrying error-contract precedent
(`DEWEY_UNAVAILABLE`, parent D4/D5) and the repo convention
of sentinel errors over string comparison.

Validation order: structural rules, then
deterministic-requires-reference-or-reason, then existence
verification. A rejected call costs one tool round-trip, zero
spawned agents.

### D3: Skill resolution is an existence check on two roots plus a traversal guard

New `forge.ResolveSkillRef(projectPath, ref)` resolves a
reference against the project root:

- No `/` in `ref`: skill name -- check
  `.opencode/skills/<name>/SKILL.md`, then
  `.opencode/skill/<name>/SKILL.md` (both roots exist across
  org repos; a name in both resolves to the first root).
- `/` in `ref`: repo-relative path -- reject absolute paths
  and any `..` path element after cleaning, then existence-
  check the path.

Failures return `ErrSkillNotFound` naming the searched
locations. Resolution uses `os.Stat` only -- no registry, no
manifest, no new dependencies. Tests use `t.TempDir()` with
synthetic skill roots and script files (Constitution IV).

Alternatives rejected:

- **Skill registry/manifest** -- adds a drift surface for
  zero need (parent D2, mirrored).
- **Searching user-level skill directories** -- the rule
  binds delegated mechanical work to repo tooling.

### D4: The directive is a tool-rendered constant

A fixed `skillDirective(reference)` renders the worker
directive ("load the skill / run the script and execute its
full procedure; do not substitute an improvised procedure").
It appears in exactly two places, both tool-produced:

- the `forge_spawn_subtask` success envelope as
  `skill_directive`
- the `forge_subtask_prompt` composed prompt, via a new
  optional `skill` parameter on that tool

The wording is a design constant covered by golden-string
tests. The coordinator never reconstructs it from prose.

### D5: The envelope is additive; the signature becomes a struct

`SpawnSubtask` grows from six positional parameters to ten
inputs. Rather than a ten-arg call, the function takes a
`SpawnParams` struct (`internal/forge`), and the tool layer
unmarshals JSON into it. The success envelope keeps every
existing field name and shape, adding `task_classification`,
the resolved `skill` path, `skill_absent_reason` (verbatim),
and `skill_directive` (when a reference was supplied).
Existing internal callers are updated in the same change;
`SpawnSubtask` is not referenced outside this repo.

`project_path` is optional with default `.` -- consistent with
the `project_path` convention of comms and worktree tools --

and is only consulted when a reference needs resolution, so
judgment-classified spawns need no filesystem anchor.

### D6: Embedded agent kit and docs move with the contract

The coordinator agent, `/forge` command, and
forge-coordination skill show `forge_spawn_subtask` call
shapes; each is updated to carry `task_classification` and
the reference-or-reason rule (agent-kit changes are
spec-gated -- this change is that spec). `docs/tools.md` is
regenerated via `replicator docs`.

Cross-repo: an issue is filed in unbound-force/unbound-force
to move the `uf.unleash.md` Swarm-path prose requirement into
tool parameters, closing the loop the parent change opened in
its task 5.5. Replicator's error remediation text names the
missing parameter and accepted values, so a version-skewed
caller fails self-describingly rather than confusingly.

## Risks / Trade-offs

- **Breaking change for version-skewed callers.** Hero-repo
  templates that omit `task_classification` fail closed until
  the unbound-force template update lands. Mitigation:
  remediation-carrying error text, a tracked cross-repo
  issue, and internal mirrors updated in the same change.
  Accepted -- this is the requirement working as intended
  (parent precedent: "future delegating commands must
  classify").
- **Misclassification evades the tool.** A caller can label
  judgment work to skip the requirement. The classification
  is now an explicit, recorded envelope field the guard layer
  audits against the task text; the prior alternative was no
  claim at all. Layered mitigation, accepted.
- **Escape-hatch abuse.** `skill_absent_reason` could become
  a free pass. It is recorded verbatim in durable state and
  is the first thing a reviewer reads; vacuous reasons are
  flaggable findings at the guard layer (already in place per
  the parent change). Accepted.
- **Skill-root hardcoding.** The two roots are org
  convention; the resolution order is a documented constant.
  A future layout change is a one-constant edit.
- **Prompt-regression surface.** `SubtaskPrompt` gains a
  conditional section; golden-string tests pin both the
  with-directive and without-directive shapes so existing
  prompt consumers see no diff when no skill is supplied.
