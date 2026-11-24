## Why

Dewey learnings PRs (`chore(dewey): store learnings from ...`, branches
`chore/dewey-learnings-*`) only add Markdown files under
`.uf/dewey/learnings/`. gh-hud opens them from the fleet's checkouts and
uf-divisor approves them when every file is a learning. Nothing in them
can break a build, a test, or a lint, yet each one runs the repo's full
CI on every push and every update from main. On 2026-11-24, gaze had 16
open learnings PRs at roughly 30 minutes each, and the queue stalled
behind unbound-force/gaze#334, whose duplicate runs were cancelled.

In a repo that requires branches to be up to date (branch protection on
`main` has `strict: true`), the merge train lands these PRs one at a
time: update from main, wait for full CI, merge, repeat. Skipping CI
for learnings-only PRs collapses that wait to a couple of minutes.
replicator is not in gh-hud's `[auto_merge]` repos yet, so its approved
learnings PRs wait for a hand merge regardless -- once CI is skipped
for them, adding replicator there would let them land on their own (a
gh-hud-side follow-up, outside this repo).

Workflow-level `paths-ignore` cannot be used: these workflows provide
required checks, and a workflow that does not run never reports them,
leaving PRs waiting forever on "Expected -- Waiting for status to be
reported".

Fixes: https://github.com/unbound-force/replicator/issues/182

## What Changes

1. **Change detection job**: Add a cheap `changes` job to
   `.github/workflows/ci.yml` and `.github/workflows/ci_checks.yml`.
   On `pull_request` events it runs
   `git diff --name-only --no-renames "$BASE_SHA...$HEAD_SHA"` (base
   and head SHAs supplied through the step's `env:` block, never
   interpolated into `run:`, matching the repo's zizmor posture) and
   sets output `code` to `'true'` when any changed file is outside
   `.uf/dewey/learnings/`, `'false'` otherwise. `--no-renames` makes a
   file moved into `.uf/dewey/learnings/` still count its source path
   as a change outside learnings. On `push` events to `main` it sets
   `code` to `'true'` unconditionally. Detection fails closed: any
   diff error fails the job.

2. **Gated expensive jobs**: `build-and-test` (ci.yml) and
   `call-reusable-ci` (ci_checks.yml) get `needs: [changes]` and
   `if: needs.changes.outputs.code == 'true'`. Job steps, flags
   (`-race`, `-count=1`, govulncheck, coverage ratchets), action pins,
   and the reusable workflow call are otherwise unchanged.

3. **Aggregate required checks**: Each workflow gains an aggregate
   job (id `ci-ok`) -- named `CI OK` in ci.yml and
   `Standardized CI OK` in ci_checks.yml -- with `if: always()` and
   `needs` covering its `changes` job and every gated job. The
   aggregate fails when any needed job failed or was cancelled;
   skipped counts as success. The distinct names keep the two
   workflows' check runs from colliding, since check-run names are
   not namespaced by workflow.

4. **Branch protection, live and declarative**: Required checks move
   from `Build and Test` and `Standardized CI / Run linters` to
   `CI OK` and `Standardized CI OK`. The live protection is updated
   with the GitHub API at merge time (a maintainer step; not an
   in-repo action), and `.github/settings.yml` -- which declares the
   same `required_status_checks.contexts` -- is updated by this
   change's PR so a settings sync cannot revert the replacement.

5. **Workflow gating regression test**:
   `internal/workflowtest/ci_gating_test.go`, in the style of the
   existing `ci_dependencies_test.go` (reusing its
   `yamlSection`/`readRepositoryFile` helpers), asserts the wiring:
   the `changes` jobs and their `code` output, the gated jobs'
   `needs`/`if:`, the aggregates' `if: always()`/`needs`/failure
   semantics, the absence of `paths`/`paths-ignore` triggers, and the
   settings.yml contexts. It runs under `make check` with no network
   or external services.

6. **Documentation**: `AGENTS.md` notes the skip behavior and updates
   required-check references to the aggregate names.

Unchanged: `push` runs on `main` (always full CI; every check name
that reports today still reports), `ci_dependencies.yml`, the release
preflight (`release.yml` passes
`ci_checks: '["Build and Test", "Standardized CI / Run linters"]'` to
ship, and both names still report on main pushes), and every job's
internal behavior.

## Capabilities

### New Capabilities

- `ci-learnings-skip`: Change-detection gating for CI jobs on pull
  requests -- which PRs run the expensive build/test/lint jobs, how
  the decision is made (learnings-only path filter, rename-safe,
  fail-closed), and how required checks report when jobs are skipped
  (aggregate jobs, live and declarative branch protection). Allocates
  `CLS-FR-001` through `CLS-FR-006`.

### Modified Capabilities

- None. `openspec/specs/` has no established capabilities; this change
  adds the `ci-learnings-skip` capability.

### Removed Capabilities

- None.

## Impact

- `.github/workflows/ci.yml` -- `changes` job, gated `build-and-test`,
  new `CI OK` aggregate
- `.github/workflows/ci_checks.yml` -- `changes` job, gated
  `call-reusable-ci`, new `Standardized CI OK` aggregate
- `.github/settings.yml` -- `required_status_checks.contexts` becomes
  the two aggregate names
- `internal/workflowtest/ci_gating_test.go` -- new gating regression
  test
- `AGENTS.md` -- required-check references and skip-behavior note
- GitHub branch protection settings (live) -- re-pointed to
  `CI OK` and `Standardized CI OK` at merge time (manual maintainer
  step, part of tasks)
- CI behavior: learnings-only PRs show required checks passing within
  a couple of minutes without running build, tests, or linters; all
  other PRs (including learnings mixed with any other file) run full
  CI exactly as today; push runs on main are unchanged
- Gatekeeping note: this change adds conditionality to when jobs run
  but weakens no gate -- every PR still gets a passing/failing verdict
  from the aggregates, detection fails closed, and the gated jobs keep
  their flags, scanners, ratchets, and pins untouched

## Constitution Alignment

Assessed against the Replicator constitution
(`.specify/memory/constitution.md`), which extends the Unbound Force
org constitution v1.1.0.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies CI workflow files, branch-protection
configuration, and a Go regression test only. No MCP tools,
inter-agent communication, or tool outputs are affected. gh-hud and
uf-divisor continue to operate on learnings PRs exactly as before;
they simply wait on different (faster) check names.

### II. Composability First

**Assessment**: PASS

The binary is untouched. No new third-party actions are introduced --
the new jobs reuse the already-pinned `actions/checkout` SHA, and
detection is a plain git step. Standalone operation, Dewey
integration, and graceful degradation are unaffected.

### III. Observable Quality

**Assessment**: PASS

The `changes` job emits a machine-readable `code` output; every skip
or run decision is visible as a job result with logged reasoning; the
aggregate jobs (`CI OK`, `Standardized CI OK`) give each workflow a
single observable pass/fail check. Detection fails closed -- a broken
detector fails the required checks rather than silently skipping CI,
so the quality signal remains trustworthy.

### IV. Testability

**Assessment**: PASS

The gating wiring is verified by a Go regression test in
`internal/workflowtest/` that reads the workflow files and settings.yml
from the repository and asserts their structure -- running under
`make check` in isolation, with no network, no Node.js, and no
external services (the package's established pattern). The
`needs`/`if` runtime behavior itself is GitHub Actions glue, verified
by inspection, actionlint/zizmor in CI, and post-merge observation of
the three PR classes.
