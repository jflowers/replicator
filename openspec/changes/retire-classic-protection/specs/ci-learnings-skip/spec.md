## ADDED Requirements

None. This change modifies existing requirements; no new
requirements are added.

## MODIFIED Requirements

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
required unchanged. The required checks on `main` are enforced
by the `safe-settings: replicator required checks` ruleset
defined in the `unbound-force/.github` repository
(`repos/replicator.yml`), and `.github/settings.yml` MUST NOT
manage branch protection: the `branches` entry for `main`
sets `protection` to `null` and the file declares no
`required_status_checks`.

Previously: "The declarative branch-protection source
`.github/settings.yml` MUST list the same two checks in its
`required_status_checks.contexts`, so a settings sync cannot
revert the replacement."

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

#### Scenario: Settings sync retires classic protection

- **GIVEN** the `safe-settings: replicator required checks`
  ruleset requires `CI OK` and `Standardized CI / Run linters`
  on `main`
- **WHEN** `.github/settings.yml` is applied by a settings sync
- **THEN** classic branch protection on `main` MUST be removed
- **AND** the required checks on `main` MUST remain `CI OK` and
  `Standardized CI / Run linters`, enforced by the ruleset

Previously this scenario was "Settings sync cannot revert the
required checks", whose THEN read: "the required checks MUST
remain `CI OK` and `Standardized CI / Run linters`" with
`.github/settings.yml` as the enforcement source.

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
It MUST also assert that `.github/settings.yml` does not manage
branch protection: the `branches` entry for `main` sets
`protection` to `null`, and the file declares no
`required_status_checks` key. These settings assertions MUST be
structural — scoped to the `main` branch entry and matched by
key shape, not as free text anywhere in the file — so comments
and formatting changes neither weaken them nor fail them
spuriously. The test MUST fail when any of this wiring is
removed or weakened.

Previously: "It MUST also assert that `.github/settings.yml`
lists exactly `CI OK` and `Standardized CI / Run linters` in its
`required_status_checks.contexts`."

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

#### Scenario: Re-introduced classic protection fails the test

- **GIVEN** `.github/settings.yml` replaces `main`'s
  `protection: null` with a protection block, or declares a
  `required_status_checks` key
- **WHEN** the workflow gating regression test runs
- **THEN** it MUST fail

## REMOVED Requirements

None. No existing requirements are removed by this change.
