<!--
  Requirement numbering allocation (meta-prompt MP-014):

  ci-learnings-skip: CLS-FR-001 .. CLS-FR-006 are allocated by
  this change; CLS-FR-007 onward is reserved for future changes
  of the same capability. Sibling capability deltas live under
  openspec/changes/*/specs/ and use named requirements without
  FR ids; openspec/specs/ is empty. This change allocates no FR
  ranges outside the ci-learnings-skip capability.
-->

## ADDED Requirements

### Requirement: Changes detection job

**ID**: CLS-FR-001

Both `.github/workflows/ci.yml` and `.github/workflows/ci_checks.yml`
MUST define a `changes` job that runs on every event the workflow
handles. On `pull_request` events it MUST compute the changed files
against the pull request base (`git diff --name-only --no-renames
"$BASE_SHA...$HEAD_SHA"`) and set output `code` to `'true'` when any
changed file is outside `.uf/dewey/learnings/`, and `'false'`
otherwise. Rename detection MUST be disabled in the diff so that a
file moved into `.uf/dewey/learnings/` still counts its source path
as a change outside learnings. On `push` events to `main` it MUST set
`code` to `'true'` unconditionally. If the diff step fails, the
`changes` job MUST fail rather than guess (fail closed).

#### Scenario: Learnings-only PR detected

- **GIVEN** a pull request whose changed files are all under
  `.uf/dewey/learnings/`
- **WHEN** the `changes` job runs
- **THEN** output `code` MUST equal `'false'`

#### Scenario: Mixed PR detected

- **GIVEN** a pull request adding one file under
  `.uf/dewey/learnings/` and one file anywhere else
- **WHEN** the `changes` job runs
- **THEN** output `code` MUST equal `'true'`

#### Scenario: Push to main always runs CI

- **GIVEN** a push to `main`
- **WHEN** the `changes` job runs
- **THEN** output `code` MUST equal `'true'` without consulting
  the diff

#### Scenario: Rename into learnings counts as code

- **GIVEN** a pull request that moves a file from a path outside
  `.uf/dewey/learnings/` to a path inside it
- **WHEN** the `changes` job runs
- **THEN** output `code` MUST equal `'true'`

#### Scenario: Diff failure fails closed

- **GIVEN** the base SHA cannot be resolved or `git diff` exits
  non-zero
- **WHEN** the `changes` job runs
- **THEN** the job MUST fail
- **AND** no gated job MUST report success

### Requirement: Gated execution of expensive jobs

**ID**: CLS-FR-002

The `build-and-test` job in `ci.yml` and the `call-reusable-ci`
job in `ci_checks.yml` MUST declare `needs` on `changes` and a
job-level `if:` gating on `needs.changes.outputs.code ==
'true'`, so they run only when the pull request changes files
outside `.uf/dewey/learnings/` (or the event is a push to
`main`). Their steps MUST NOT be otherwise modified.

#### Scenario: Learnings-only PR skips build, test, and lint

- **GIVEN** the `changes` job output `code` is `'false'`
- **WHEN** the workflows run
- **THEN** `build-and-test` and `call-reusable-ci` MUST be
  skipped by their `if:`
- **AND** no Go build, Go test, govulncheck, or MegaLinter step
  MUST execute

#### Scenario: Code PR runs full CI

- **GIVEN** the `changes` job output `code` is `'true'`
- **WHEN** the workflows run
- **THEN** `build-and-test` and `call-reusable-ci` MUST run with
  the same steps as before this change

#### Scenario: Push to main runs full CI

- **GIVEN** a push to `main`
- **WHEN** the workflows run
- **THEN** both gated jobs MUST run

### Requirement: Aggregate required checks

**ID**: CLS-FR-003

Each gated workflow MUST define an aggregate job -- `CI OK` in
`ci.yml` and `Standardized CI OK` in `ci_checks.yml` -- with
`if: always()` and `needs` covering its `changes` job and every
gated job. The aggregate MUST succeed when every needed job
either succeeded or was skipped, and MUST fail when any needed
job failed or was cancelled. The two aggregate names MUST be
distinct so the checks do not collide across workflows. Branch
protection on `main` MUST require exactly the two aggregate
checks, replacing `Build and Test` and
`Standardized CI / Run linters`. The declarative
branch-protection source `.github/settings.yml` MUST list the same
two aggregate checks in its `required_status_checks.contexts`, so a
settings sync cannot revert the replacement.

#### Scenario: Learnings-only PR is green within minutes

- **GIVEN** a pull request whose changed files are all under
  `.uf/dewey/learnings/`
- **WHEN** the checks settle
- **THEN** `CI OK` and `Standardized CI OK` MUST report success
- **AND** no build, test, or lint job MUST have run

#### Scenario: Gated job failure fails the aggregate

- **GIVEN** `code` is `'true'` and `build-and-test` fails
- **WHEN** `CI OK` evaluates its needs
- **THEN** `CI OK` MUST fail

#### Scenario: Cancelled job fails the aggregate

- **GIVEN** one of an aggregate's needed jobs was cancelled
- **WHEN** the aggregate evaluates its needs
- **THEN** the aggregate MUST fail

#### Scenario: Skipped gated job passes the aggregate

- **GIVEN** `code` is `'false'` so the gated job was skipped by
  its `if:`
- **WHEN** the aggregate evaluates its needs
- **THEN** the aggregate MUST succeed

#### Scenario: Detection failure fails the aggregate

- **GIVEN** the `changes` job failed
- **WHEN** the aggregate evaluates its needs
- **THEN** the aggregate MUST fail

#### Scenario: Settings sync cannot revert the required checks

- **GIVEN** branch protection on `main` requires `CI OK` and
  `Standardized CI OK`
- **WHEN** `.github/settings.yml` is applied by a settings sync
- **THEN** the required checks MUST remain `CI OK` and
  `Standardized CI OK`

### Requirement: Push runs on main are unchanged

**ID**: CLS-FR-004

On push to `main`, every job that runs before this change MUST
still run with identical steps and triggers after it, and every
check name that reported before MUST still report. Path-based
skipping MUST apply only to `pull_request` events.

#### Scenario: Main keeps its check names

- **GIVEN** a push to `main`
- **WHEN** the workflows complete
- **THEN** the `Build and Test` job and the nested
  `Standardized CI / Run linters` check MUST both report as they
  do before this change

### Requirement: No workflow-level path filtering

**ID**: CLS-FR-005

Workflows that provide required checks MUST NOT use `paths` or
`paths-ignore` on their `on:` triggers. The skip MUST be
implemented as job-level gating so that every required check
always reports on every pull request.

#### Scenario: Triggers stay unfiltered

- **GIVEN** `ci.yml` and `ci_checks.yml`
- **WHEN** their `on:` blocks are inspected
- **THEN** neither MUST contain `paths` or `paths-ignore`

#### Scenario: No required check is left waiting

- **GIVEN** a learnings-only pull request
- **WHEN** the checks settle
- **THEN** no required check MUST remain in the
  "Expected -- Waiting for status to be reported" state

### Requirement: Workflow gating regression test

**ID**: CLS-FR-006

`internal/workflowtest/` MUST contain a Go test that reads both
workflow files and asserts the gating wiring: each workflow
defines a `changes` job producing output `code`; each gated job
declares `needs` on `changes` and the gating `if:`; each
aggregate job uses `if: always()`, needs its workflow's
`changes` job and gated jobs, and fails on `failure` or
`cancelled`; and neither workflow gains `paths` or
`paths-ignore` triggers. It MUST also assert that
`.github/settings.yml` lists exactly the two aggregate checks in
its `required_status_checks.contexts`. The test MUST fail when any
of this wiring is removed or weakened.

#### Scenario: Gating wiring asserted by make check

- **GIVEN** the gating from this change is in place
- **WHEN** `make check` runs
- **THEN** the workflow gating regression test MUST pass

#### Scenario: Removed gating fails the test

- **GIVEN** the `if:` condition is deleted from
  `build-and-test`, or an aggregate job is removed
- **WHEN** the workflow gating regression test runs
- **THEN** it MUST fail

## MODIFIED Requirements

None. This change adds new requirements; no existing requirement
text is modified.

## REMOVED Requirements

None. No existing requirements are removed by this change.
