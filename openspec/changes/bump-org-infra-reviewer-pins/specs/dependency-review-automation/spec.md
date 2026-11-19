## ADDED Requirements

### Requirement: Dependency Review Workflow Pinning

The `.github/workflows/ci_dependencies.yml` workflow MUST pin both org-infra
reusable reviewer references to the same full 40-character commit SHA
(CI-001), with an accurate `# vX.Y.Z` version comment reflecting the
resolved tag (CI-003) and a SHA re-derived at authoring time (CI-002), not
copied from prose.

#### Scenario: Both reviewer references share the resolved v0.9.0 pin

- **GIVEN** the `Dependencies` workflow defines `call_deps_reviewer` and
  `call_dependabot_reviewer` jobs
- **WHEN** the workflow file is inspected
- **THEN** both `uses:` references MUST point at the same full 40-character
  SHA resolved and verified from `complytime/org-infra` v0.9.0 in task 1.1
- **AND** both MUST carry an accurate `# v0.9.0` version comment

#### Scenario: Workflow reference is auditable

- **GIVEN** the Dependencies workflow calls an action or reusable workflow
- **WHEN** a maintainer inspects its `uses` reference
- **THEN** the reference MUST contain a verified 40-character commit SHA
- **AND** any adjacent version comment MUST identify the tag from which
  that SHA was resolved

### Requirement: Workflow Regression Test Matches the Pin

`internal/workflowtest/ci_dependencies_test.go` MUST assert both reviewer
references against the shared `orgInfraSHA` constant, and `orgInfraSHA`
MUST equal the SHA pinned in the workflow file.

#### Scenario: make check passes with the lockstep bump

- **GIVEN** `orgInfraSHA` equals the full 40-character SHA resolved and
  verified from `complytime/org-infra` v0.9.0 in task 1.1
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
