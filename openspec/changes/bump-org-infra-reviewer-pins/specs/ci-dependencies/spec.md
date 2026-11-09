## ADDED Requirements

### Requirement: Dependency Review Workflow Pins

The `.github/workflows/ci_dependencies.yml` workflow MUST pin both
org-infra reusable reviewer references to the same full 40-character
commit SHA with an accurate `# vX.Y.Z` version comment, per CI convention
CI-001. The SHA MUST be re-derived from `complytime/org-infra` when the
reference is written (CI-002), not copied from prose.

#### Scenario: Both reviewer references share the v0.9.0 pin

- **GIVEN** the `Dependencies` workflow defines `call_deps_reviewer` and
  `call_dependabot_reviewer` jobs
- **WHEN** the workflow file is inspected
- **THEN** both `uses:` references MUST point at
  `8f347ba1aa2c7e9eb3b70e69c1709ddb456cd64a`
- **AND** both MUST carry the `# v0.9.0` version comment

#### Scenario: Out-of-scope references are unchanged

- **GIVEN** the `ci_checks.yml` and `release.yml` workflows also reference
  org-infra
- **WHEN** this change is applied
- **THEN** those references MUST remain on their existing v0.8.0 pin
  because they are not asserted by `orgInfraSHA`

### Requirement: Workflow Regression Test Matches the Pin

`internal/workflowtest/ci_dependencies_test.go` MUST assert both reviewer
references against the shared `orgInfraSHA` constant, and `orgInfraSHA`
MUST equal the SHA pinned in the workflow file.

#### Scenario: make check passes with the lockstep bump

- **GIVEN** `orgInfraSHA` equals
  `8f347ba1aa2c7e9eb3b70e69c1709ddb456cd64a`
- **AND** both workflow references are pinned to the same SHA
- **WHEN** `make check` runs
- **THEN** `TestCIDependenciesWorkflow_StructureEnforcesGuardedApproval`
  MUST pass
- **AND** the full 40-character SHA assertion MUST remain in force

#### Scenario: A single-reference bump is rejected

- **GIVEN** only one of the two reviewer references is bumped to a new SHA
- **AND** `orgInfraSHA` still holds the old SHA
- **WHEN** the workflow regression test runs
- **THEN** the test MUST fail at the assertion for the other reference
- **AND** the pins MUST NOT be loosened or regex-relaxed to pass

### Requirement: Failed Dependency Scan Halts Auto-Approval

Because the v0.9.0 deps reviewer reports `REVIEW_CONCLUSION` from the step
`outcome`, a failed dependency scan MUST surface as `failure` rather than
the previously masked `success`.

#### Scenario: Failed scan routes to manual review

- **GIVEN** the dependency-review step completes with a failure but
  `continue-on-error: true`
- **WHEN** the approval job evaluates `review_conclusion`
- **THEN** `review_conclusion` MUST be `failure`
- **AND** Dependabot auto-approval MUST NOT occur
- **AND** the pull request MUST go to manual review
