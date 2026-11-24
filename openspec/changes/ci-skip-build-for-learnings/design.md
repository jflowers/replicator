# Design: Skip CI Build, Test, and Lint for Learnings-Only PRs

**Change**: ci-skip-build-for-learnings

## Context

Three workflows run on pull requests today:

- `ci.yml` (`CI`): one job, `build-and-test`, named
  `Build and Test` -- vet, govulncheck, a GoReleaser snapshot,
  the Homebrew cask integrity test,
  `go test ./... -count=1 -race` with coverage ratchets, and the
  binary build.
- `ci_checks.yml` (`Standardized CI`): one reusable workflow
  call, `call-reusable-ci`, to
  `complytime/org-infra/.github/workflows/reusable_ci.yml`
  pinned at v0.9.0. The nested check reports as
  `Standardized CI / Run linters` (MegaLinter plus PR title
  validation).
- `ci_dependencies.yml` (`Dependencies`): dependency review and
  Dependabot approval gating. Not required.

Branch protection on `main` (verified via the GitHub API)
requires `Build and Test` and `Standardized CI / Run linters`,
with `strict: true` -- branches must be up to date before
merging. Learnings PRs therefore serialize behind full CI runs
even though they only add Markdown under `.uf/dewey/learnings/`.

Two GitHub Actions behaviors constrain the design (issue #182):

- A job skipped by its `if:` still satisfies a required check of
  the same name.
- A skipped *matrix* job reports once with the expression
  unexpanded, and a skipped *reusable workflow call* reports only
  the caller's name -- the nested check (for example
  `Standardized CI / Run linters`) never reports. A required
  check that never reports blocks the PR forever.

## Goals / Non-Goals

### Goals

- Learnings-only PRs show every required check passing within a
  couple of minutes, without running build, tests, or linters.
- Any other PR -- code only, or learnings mixed with code --
  runs full CI exactly as today.
- Push runs on `main` are unchanged, so release preflight (which
  checks CI on HEAD by name) keeps working.

### Non-Goals

- Gating `ci_dependencies.yml` (not required, comparatively
  cheap; revisit separately if wanted).
- Adding replicator to gh-hud's `[auto_merge]` repos (separate
  decision referenced by issue #182).
- Generalizing the skip to other path sets (the design permits
  growth; only the learnings path is configured).
- Weakening any quality gate: coverage ratchets, `-race`,
  `-count=1`, govulncheck, and MegaLinter run unchanged wherever
  they run today.

## Decisions

### D1. Job-level gating, never `paths-ignore`

Required checks must always report. `paths-ignore` on a workflow
that provides a required check leaves the check in "Expected --
Waiting" forever. Both workflows instead gain a cheap `changes`
detection job, and the expensive jobs are skipped by job-level
`if:`. A same-name skipped job would already satisfy a required
check, but the aggregates in D5 carry the guarantee regardless
and cover the reusable-call caveat.

### D2. Plain `git diff`, no new action dependency

The `changes` job uses the already-pinned
`actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1`
(v7.0.1) with `fetch-depth: 0` and

    git diff --name-only --no-renames "$BASE_SHA...$HEAD_SHA"

where `BASE_SHA` is `github.event.pull_request.base.sha` and
`HEAD_SHA` is `github.event.pull_request.head.sha`, both supplied
through the step's `env:` block rather than interpolated into
`run:` (event-context template injection is a zizmor finding; cf.
`ci_dependencies.yml`'s env indirection). Any line
lacking the `.uf/dewey/learnings/` prefix sets `code: 'true'`.
`--no-renames` disables rename detection: without it,
`--name-only` lists only a rename's new path, so a file moved into
`.uf/dewey/learnings/` would hide the deletion of its source
outside learnings -- fail-open for exactly the PRs the gate must
catch. This avoids adding `dorny/paths-filter` -- a new third-party
action to pin and vet -- for a diff the runner can do with git.
The three-dot diff uses the merge base, so files changed on
`main` after the PR forked (already validated by main's own CI)
are not counted against the PR.

### D3. Push to `main` outputs `code: 'true'` unconditionally

One source of truth: on `push` events the `changes` job
short-circuits to `code: 'true'` without consulting a diff.
Every gated job then runs, and the job graph on `main` is
today's plus two cheap detection jobs and two aggregates. Push
behavior, step lists, and check names on `main` are unchanged.

### D4. Gated jobs: `build-and-test` and `call-reusable-ci`

`ci.yml`'s `build-and-test` and `ci_checks.yml`'s
`call-reusable-ci` each get `needs: [changes]` and
`if: needs.changes.outputs.code == 'true'`. Neither job's steps
change. `ci_dependencies.yml` is left ungated (Non-Goal).

### D5. One aggregate per workflow, with distinct names

Each workflow gets a `ci-ok` aggregate job:

- `ci.yml`: job name `CI OK`,
  `needs: [changes, build-and-test]`
- `ci_checks.yml`: job name `Standardized CI OK`,
  `needs: [changes, call-reusable-ci]`

Both use `if: always()` and fail when any needed job's result is
`failure` or `cancelled`; skipped counts as success. The
aggregate covers the reusable-call caveat from the issue (the
nested `Standardized CI / Run linters` check never reports when
the caller is skipped), covers future matrix jobs the same way,
and makes later changes to the gated job list free -- the
aggregate's `needs` is the only wiring to extend. Distinct names
keep the two workflows' checks from colliding, since check-run
names are not namespaced by workflow.

### D6. Branch protection requires the aggregates

Required checks become `CI OK` and `Standardized CI OK`,
replacing `Build and Test` and
`Standardized CI / Run linters`. This is strictly stronger, not
weaker: on any PR where the gated jobs run, an aggregate failure
blocks the merge; where they are skipped, the skip was earned by
the path check. The change is made with the GitHub API at merge
time (task 5.2), because a learnings PR opened after the
workflows merge but before the protection update would hang on
`Standardized CI / Run linters` never reporting.

`.github/settings.yml` declares the same protection
(`required_status_checks.contexts`) and is updated by this change's
PR to the two aggregate names, so a settings sync that re-applies
the file agrees with the live configuration instead of reverting
it. The regression test (D7) pins the file's contexts alongside
the workflow wiring.

Release preflight is unaffected: it inspects CI on `main` HEAD,
where push runs still report every check under today's names
(D3).

### D7. Workflow structure regression test

`internal/workflowtest/` gains a test in the style of
`ci_dependencies_test.go`, which reads repository workflow files
and asserts structure via the existing `yamlSection` /
`readRepositoryFile` helpers (no yaml dependency, no Node.js, no
network for structural assertions). It asserts the `changes`
wiring, the gated jobs' `needs`/`if:`, the aggregates'
`if: always()`/`needs`/failure semantics, and the absence of
`paths`/`paths-ignore`. This keeps the constitution's
Testability principle honest: the gating is verified by
`make check`, in isolation, with no external services.

### D8. Capability and numbering

New capability `ci-learnings-skip` (delta at
`specs/ci-learnings-skip/spec.md`) allocates `CLS-FR-001`
through `CLS-FR-006`; `CLS-FR-007` onward is reserved for
future changes of the same capability (meta-prompt MP-014).

### D9. Task phase structure and checkpoints

tasks.md closes each code-bearing work phase with a dedicated
single-task checkpoint phase (phase 2 closes phase 1; phase 4
closes phase 3) so that no unmarked task follows `[P]` tasks
inside a phase -- that would invert MP-017's
sequential-then-parallel ordering -- while the work still ends
at a test-suite checkpoint (MP-018). The final docs and
verification phase (phase 5) produces no code, so it
intentionally ends without a test-suite checkpoint: its gates
are `openspec validate` (task 5.3), the constitution check
(task 5.4), and the issue #182 post-merge acceptance
(task 5.5). This records the MP-018 exception for that phase.

## Risks / Trade-offs

- **Required-check transition window.** Between merging the
  workflows and updating branch protection, learnings-only PRs
  would hang on `Standardized CI / Run linters` never reporting;
  conversely, PRs from branches without the new jobs would not
  emit the new aggregate checks once protection requires them.
  Mitigation: update protection in the same change (task 5.2,
  at merge time) and push a trivial update to any other open PR
  at that moment.
- **Wrong path filter skips CI for a code PR.** The filter is an
  exact `.uf/dewey/learnings/` prefix; diff errors fail closed
  (CLS-FR-001); the regression test (D7) pins the wiring. A
  learnings PR that also touches `go.mod` or any Go file runs
  full CI.
- **Skipped jobs hide run details.** A learnings-only PR shows
  no coverage numbers or lint results. Accepted: Markdown-only
  changes cannot affect them, and the skip is visible as
  `skipped`, not silent.
- **Stale base SHA.** `github.event.pull_request.base.sha` is
  the base at event time; the merge-base semantics of the
  three-dot diff (D2) keep the result correct for the question
  being asked -- did this PR change anything outside learnings.
- **Extra detection jobs on every run.** Two cheap jobs
  (seconds each) per workflow per run. Accepted for the
  ~30-minute savings on learnings PRs.

Constitution alignment carries over from the proposal: no new
runtime or action dependency (II), observable `skipped` states
and fail-closed aggregates keep quality claims machine-verifiable
(III), and the structure regression test keeps the change
testable in isolation (IV).
