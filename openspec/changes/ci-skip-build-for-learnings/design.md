# Design: Skip CI Build and Test for Learnings-Only PRs

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
  few minutes, without running the Go build or tests.
- Secret scanning and PR title validation still run on every PR,
  learnings-only ones included.
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
- Skipping `ci_checks.yml` (see D4: it carries the secret scan and
  the PR title check, and is cheap on learnings-only PRs).

## Decisions

### D1. Job-level gating, never `paths-ignore`

Required checks must always report. `paths-ignore` on a workflow
that provides a required check leaves the check in "Expected --
Waiting" forever. Instead `ci.yml` gains a cheap `changes`
detection job, and `build-and-test` is skipped by a job-level
`if:`. A same-name skipped job would already satisfy a required
check, but the aggregate in D5 carries the guarantee regardless.

### D2. Plain `git diff`, no new action dependency

The `changes` job uses the already-pinned
`actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1`
(v7.0.1) with `fetch-depth: 2`. On `pull_request` the checkout
is `refs/pull/N/merge`, GitHub's merge of the PR head into the
current `main`; depth 2 fetches it and both parents. The job runs

    git diff --name-only --no-renames 'HEAD^1' HEAD

after confirming that `HEAD^2` is the PR head
(`github.event.pull_request.head.sha`, supplied through the step's
`env:` block rather than interpolated into `run:` --
event-context template injection is a zizmor finding; cf.
`ci_dependencies.yml`'s env indirection). Diffing the merge
commit against its first parent lists exactly what merging the PR
would change on today's `main`.

An earlier draft used a three-dot diff against
`github.event.pull_request.base.sha`. That SHA can be older than
the `main` merged into the branch by "Update branch", and then the
merge base is the old base: every change `main` gained since is
counted against the PR, so a learnings PR in the strict-mode merge
train runs full CI anyway. The first-parent diff does not depend
on `base.sha`. Any line
lacking the `.uf/dewey/learnings/` prefix sets `code: 'true'`.
`--no-renames` disables rename detection: without it,
`--name-only` lists only a rename's new path, so a file moved into
`.uf/dewey/learnings/` would hide the deletion of its source
outside learnings -- fail-open for exactly the PRs the gate must
catch. This avoids adding `dorny/paths-filter` -- a new third-party
action to pin and vet -- for a diff the runner can do with git.
Files changed on `main` (already validated by main's own CI) are
never counted against the PR.

### D3. Push to `main` outputs `code: 'true'` unconditionally

One source of truth: on `push` events the `changes` job
short-circuits to `code: 'true'` without consulting a diff.
Every gated job then runs, and the job graph on `main` is
today's plus one cheap detection job and one aggregate. Push
behavior, step lists, and check names on `main` are unchanged.

### D4. Gate only `build-and-test`; never gate `call-reusable-ci`

`ci.yml`'s `build-and-test` gets `needs: [changes]` and
`if: needs.changes.outputs.code == 'true'`; its steps do not
change. This is the ~30-minute job the change exists to skip.

`ci_checks.yml` stays exactly as on `main`, with no `changes`
job and no gate. Its reusable call bundles more than linting:

- **Secret scanning.** `REPOSITORY_BETTERLEAKS` and
  `REPOSITORY_KICS` inspect the full tree, and `.mega-linter.yml`
  does not exclude `.uf/`. Learnings are written from session
  transcripts, so a learnings file is a plausible place for a
  leaked token. Gating the call would let such a PR pass and be
  auto-approved.
- **PR title validation.** commitlint checks the PR title, which
  becomes the squash-merge message on `main` (DR-005). A title's
  format has nothing to do with which files a PR touches.

Running it costs little: on pull requests MegaLinter runs with
`VALIDATE_ALL_CODEBASE` false, so per-language linters only see
changed files, and markdownlint and yamllint exclude `.uf/`. On a
learnings-only PR only the repository scanners and the title
check do real work. `ci_dependencies.yml` is left ungated
(Non-Goal).

### D5. One aggregate, in `ci.yml`

`ci.yml` gets a `ci-ok` aggregate job named `CI OK`, with
`needs: [changes, build-and-test]`. It uses `if: always()` and
fails when any needed job's result is `failure` or `cancelled`;
skipped counts as success. It covers future matrix jobs the same
way, and makes later changes to the gated job list free -- the
aggregate's `needs` is the only wiring to extend.

`ci_checks.yml` needs no aggregate: its reusable call is never
skipped (D4), so the nested `Standardized CI / Run linters`
check always reports, and the reusable-call caveat from the
issue does not arise.

### D6. Branch protection requires the aggregates

Required checks become `CI OK` and
`Standardized CI / Run linters`: `CI OK` replaces
`Build and Test`, and `Standardized CI / Run linters` stays
required as today. This is not weaker: on any PR where
`build-and-test` runs, a failure fails `CI OK` and blocks the
merge; where it is skipped, the skip was earned by the path
check; and the secret scan and title check gate every PR. The
change is made with the GitHub API at merge time (task 5.2),
because a learnings PR opened after the workflows merge but
before the protection update would show `Build and Test` as
skipped.

`.github/settings.yml` declares the same protection
(`required_status_checks.contexts`) and is updated by this change's
PR to the same two names, so a settings sync that re-applies
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
wiring, `build-and-test`'s `needs`/`if:`, the aggregate's
`if: always()`/`needs`/failure semantics, that
`call-reusable-ci` stays ungated, and the absence of
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
  would show the required `Build and Test` as skipped;
  conversely, PRs from branches without the new jobs would not
  emit `CI OK` once protection requires it.
  Mitigation: update protection in the same change (task 5.2,
  at merge time) and push a trivial update to any other open PR
  at that moment.
- **Wrong path filter skips CI for a code PR.** The filter is an
  exact `.uf/dewey/learnings/` prefix; diff errors fail closed
  (CLS-FR-001); the regression test (D7) pins the wiring. A
  learnings PR that also touches `go.mod` or any Go file runs
  full CI.
- **Skipped jobs hide run details.** A learnings-only PR shows
  no coverage numbers. Accepted: Markdown-only changes cannot
  affect them, and the skip is visible as `skipped`, not silent.
- **Learnings PRs still wait for MegaLinter.** The standardized
  check runs on every PR (D4), so a learnings-only PR waits for
  the MegaLinter container to start and run the repository
  scanners -- minutes, not the ~30 of the Go job. Accepted:
  secret scanning on content written from session transcripts is
  worth the wait.
- **Stale base SHA.** `github.event.pull_request.base.sha` can
  predate the `main` merged into the branch. Avoided: detection
  diffs the merge commit against its first parent (D2) and never
  reads `base.sha`; the regression test (D7) forbids it.
- **Extra jobs on every run.** One cheap detection job and one
  aggregate (seconds each) per `ci.yml` run. Accepted for the
  ~30-minute savings on learnings PRs.

Constitution alignment carries over from the proposal: no new
runtime or action dependency (II), observable `skipped` states
and fail-closed aggregates keep quality claims machine-verifiable
(III), and the structure regression test keeps the change
testable in isolation (IV).
