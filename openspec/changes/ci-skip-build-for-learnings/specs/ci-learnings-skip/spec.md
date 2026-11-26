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

`.github/workflows/ci.yml` MUST define a `changes` job that runs on
every event the workflow handles. On `pull_request` events it MUST
compute the changed files as the diff of the pull request merge
commit against its first parent (`git diff --raw --no-renames
'HEAD^1' HEAD` on `refs/pull/N/merge`) -- what merging the PR would
change on the current `main` -- and set output `code` to `'false'`
only when the diff is non-empty and every change is a regular `.md`
file (mode `100644` or `100755`, or a deletion) under
`.uf/dewey/learnings/`. It MUST set `code` to `'true'` for any other
file, a symlink (`120000`) or submodule (`160000`) on either side
of a change, an empty diff, or a diff line it cannot parse. Rename detection MUST be disabled in the diff so that a
file moved into `.uf/dewey/learnings/` still counts its source path
as a change outside learnings. On `push` events to `main` it MUST set
`code` to `'true'` unconditionally. If the checkout is not a merge
commit whose second parent is the pull request head, or the diff
fails, the `changes` job MUST fail rather than guess (fail closed).

#### Scenario: Learnings-only PR detected

- **GIVEN** a pull request whose changes are all regular `.md`
  files under `.uf/dewey/learnings/` (added, edited or deleted)
- **WHEN** the `changes` job runs
- **THEN** output `code` MUST equal `'false'`

#### Scenario: Non-learning content under the learnings path

- **GIVEN** a pull request adding only a non-`.md` file, a symlink,
  or a submodule under `.uf/dewey/learnings/`
- **WHEN** the `changes` job runs
- **THEN** output `code` MUST equal `'true'`

#### Scenario: Empty diff runs CI

- **GIVEN** a pull request whose merge commit changes nothing
- **WHEN** the `changes` job runs
- **THEN** output `code` MUST equal `'true'`

#### Scenario: Mixed PR detected

- **GIVEN** a pull request adding one file under
  `.uf/dewey/learnings/` and one file anywhere else
- **WHEN** the `changes` job runs
- **THEN** output `code` MUST equal `'true'`

#### Scenario: Learnings PR updated from main stays learnings-only

- **GIVEN** a pull request whose own changes are all under
  `.uf/dewey/learnings/`, whose branch was updated from a `main`
  that has since gained changes outside learnings, and whose event
  base SHA predates that `main`
- **WHEN** the `changes` job runs
- **THEN** output `code` MUST equal `'false'`

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

- **GIVEN** the checkout is not the pull request merge commit,
  its second parent is not the pull request head, or `git diff`
  exits non-zero
- **WHEN** the `changes` job runs
- **THEN** the job MUST fail
- **AND** no gated job MUST report success

### Requirement: Gated build and test, ungated security and title checks

**ID**: CLS-FR-002

The `build-and-test` job in `ci.yml` MUST declare `needs` on
`changes` and a job-level `if:` gating on
`needs.changes.outputs.code == 'true'`, so it runs only when the
pull request changes files outside `.uf/dewey/learnings/` (or
the event is a push to `main`). Its steps MUST NOT be otherwise
modified.

The `call-reusable-ci` job in `ci_checks.yml` MUST NOT be gated
(no `needs` on `changes`, no job-level `if:`). It bundles the
repository secret scanners (betterleaks, KICS) and commitlint PR
title validation. Learnings files can carry leaked secrets, and a
PR title's format does not depend on which files the PR touches,
so both MUST run on every pull request. On pull requests
MegaLinter lints only changed files, so on a learnings-only PR
the per-language linters have nothing to lint and the job stays
cheap.

#### Scenario: Learnings-only PR skips build and test

- **GIVEN** the `changes` job output `code` is `'false'`
- **WHEN** the workflows run
- **THEN** `build-and-test` MUST be skipped by its `if:`
- **AND** no Go build, Go test, or govulncheck step MUST execute
- **AND** `call-reusable-ci` MUST still run secret scanning and
  PR title validation

#### Scenario: Leaked secret in a learnings file is caught

- **GIVEN** a pull request whose changed files are all under
  `.uf/dewey/learnings/`, one of which contains a credential
- **WHEN** the workflows run
- **THEN** the repository secret scan MUST run and fail
- **AND** the `Standardized CI / Run linters` required check
  MUST fail

#### Scenario: Code PR runs full CI

- **GIVEN** the `changes` job output `code` is `'true'`
- **WHEN** the workflows run
- **THEN** `build-and-test` and `call-reusable-ci` MUST run with
  the same steps as before this change

#### Scenario: Push to main runs full CI

- **GIVEN** a push to `main`
- **WHEN** the workflows run
- **THEN** `build-and-test` and `call-reusable-ci` MUST run

### Requirement: Aggregate required check

**ID**: CLS-FR-003

`ci.yml` MUST define an aggregate job named `CI OK` with
`if: always() && !(github.event_name == 'push' && cancelled())` and `needs` covering its `changes` job and every
gated job. The aggregate MUST succeed when every needed job
either succeeded or was skipped, and MUST fail when any needed
job failed or was cancelled. It MUST be skipped only for a push
run that was itself cancelled, and never for a `pull_request`
run, since a skipped job satisfies a required check. Branch protection on `main` MUST
require exactly `CI OK` and `Standardized CI / Run linters`,
replacing `Build and Test` with `CI OK`; the ungated
`Standardized CI / Run linters` check (CLS-FR-002) stays
required unchanged. The two checks MUST be declared in
unbound-force/.github's safe-settings
(`safe-settings/repos/replicator.yml`, the
`safe-settings: replicator required checks` ruleset), without
the up-to-date rule. `.github/settings.yml` MUST set `main`'s
`protection` to `null` and MUST NOT set required status checks,
so the Settings app cannot restore classic branch protection or
its up-to-date rule.

#### Scenario: Learnings-only PR is green within minutes

- **GIVEN** a pull request whose changed files are all under
  `.uf/dewey/learnings/`
- **WHEN** the checks settle
- **THEN** `CI OK` and `Standardized CI / Run linters` MUST
  report success
- **AND** no build or test job MUST have run

#### Scenario: Gated job failure fails the aggregate

- **GIVEN** `code` is `'true'` and `build-and-test` fails
- **WHEN** `CI OK` evaluates its needs
- **THEN** `CI OK` MUST fail

#### Scenario: Cancelled job fails the aggregate

- **GIVEN** a `pull_request` run in which one of `CI OK`'s
  needed jobs was cancelled, including by cancelling the run
- **WHEN** `CI OK` evaluates its needs
- **THEN** `CI OK` MUST fail

#### Scenario: Superseded push run leaves no red check

- **GIVEN** a push run on `main` cancelled because a newer push
  started (`concurrency.cancel-in-progress`)
- **WHEN** the run finishes
- **THEN** `CI OK` MUST be skipped rather than fail

#### Scenario: Skipped gated job passes the aggregate

- **GIVEN** `code` is `'false'` so `build-and-test` was skipped
  by its `if:`
- **WHEN** `CI OK` evaluates its needs
- **THEN** `CI OK` MUST succeed

#### Scenario: Detection failure fails the aggregate

- **GIVEN** the `changes` job failed
- **WHEN** `CI OK` evaluates its needs
- **THEN** `CI OK` MUST fail

#### Scenario: Settings sync cannot revert the required checks

- **GIVEN** the `safe-settings: replicator required checks`
  ruleset requires `CI OK` and `Standardized CI / Run linters`
- **WHEN** `.github/settings.yml` is applied by the Settings app
- **THEN** `main` MUST have no classic branch protection
- **AND** the required checks MUST remain `CI OK` and
  `Standardized CI / Run linters`, from the ruleset

### Requirement: Push runs on main are unchanged

**ID**: CLS-FR-004

On push to `main`, every job that runs before this change MUST
still run with identical steps and triggers after it, and every
check name that reported before MUST still report. Path-based
skipping MUST apply only to `pull_request` events. The `changes`
job MUST NOT check out the repository on `push` events, since it
never diffs there.

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
workflow files and asserts the gating wiring: `ci.yml` defines a
`changes` job producing output `code`; `build-and-test` declares
`needs` on `changes` and the gating `if:`; the `CI OK` aggregate
uses `if: always() && !(github.event_name == 'push' && cancelled())`, needs `changes` and `build-and-test`, and
fails on `failure` or `cancelled`; `ci_checks.yml`'s only job is
`call-reusable-ci`, which declares neither `if:` nor `needs`;
and neither workflow gains `paths` or `paths-ignore` triggers.
It MUST also assert that `.github/settings.yml` sets `main`'s
`protection` to `null` and sets no `required_status_checks`, so
the required checks stay in safe-settings (CLS-FR-003). The test
MUST fail when any of this wiring is removed or weakened, or
when `.github/settings.yml` restores classic branch protection.

#### Scenario: Gating wiring asserted by make check

- **GIVEN** the gating from this change is in place
- **WHEN** `make check` runs
- **THEN** the workflow gating regression test MUST pass

#### Scenario: Removed gating fails the test

- **GIVEN** the `if:` condition is deleted from
  `build-and-test`, the `CI OK` aggregate is removed, or
  `call-reusable-ci` gains an `if:` or `needs`
- **WHEN** the workflow gating regression test runs
- **THEN** it MUST fail

## MODIFIED Requirements

- **CLS-FR-003** and **CLS-FR-006** (2026-11-26): `main`'s
  required checks moved from the classic branch protection in
  `.github/settings.yml` to unbound-force/.github's safe-settings
  (unbound-force/.github#82, replicator#205), without the
  up-to-date rule, which left every open pull request behind after
  each auto-merge. The checks themselves are unchanged.
  `.github/settings.yml` now sets `main`'s protection to `null`,
  and the regression test asserts that instead of the check list.

## REMOVED Requirements

None. No existing requirements are removed by this change.
