## ADDED Requirements

None. This change modifies existing requirements; no new
requirements are added.

## MODIFIED Requirements

### Requirement: Branch Protection

Branch protection on `main` is enforced by the
`safe-settings: replicator required checks` ruleset defined in
the `unbound-force/.github` repository (`repos/replicator.yml`),
not by this repository's `.github/settings.yml`. The ruleset
MUST require `CI OK` and `Standardized CI / Run linters` on
`main`, superseding the earlier `Build and Test` +
`Standardized CI / Run linters` pair, and MUST NOT use strict
required-status-checks policy, so open pull requests are not
blocked by stale required-check runs after target-branch
merges. `.github/settings.yml` in this repository MUST NOT
re-introduce classic branch protection or declare
`required_status_checks` for `main`.

Previously: "The `.github/settings.yml` MUST include
`Standardized CI / Run linters` in the required status checks
for the `main` branch, alongside the existing `Build and Test`
check."

#### Scenario: PR missing linter check

- **GIVEN** a pull request where `CI OK` passed
- **AND** `Standardized CI / Run linters` has not run or failed
- **WHEN** a merge is attempted
- **THEN** the merge MUST be blocked by branch protection

## REMOVED Requirements

None. No existing requirements are removed by this change.
