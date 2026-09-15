## ADDED Requirements

### Requirement: Organization Dependency Review

The repository MUST provide a `ci_dependencies.yml` consumer workflow that invokes both the general dependency reviewer and the Dependabot-specific reviewer from `complytime/org-infra`.

Every reusable workflow and action reference MUST be pinned to a verified full 40-character commit SHA and SHOULD include an accurate version comment. The workflow MUST run for pushes and pull requests targeting `main`, MUST use explicit least-privilege permissions, and SHOULD cancel redundant runs for the same ref.

#### Scenario: Pull request receives dependency analysis
- **GIVEN** a pull request targets `main`
- **WHEN** the Dependencies workflow runs
- **THEN** the general dependency reviewer MUST evaluate dependency changes
- **AND** the Dependabot-specific reviewer MUST expose its risk and release metadata for downstream jobs

#### Scenario: Workflow reference is auditable
- **GIVEN** the Dependencies workflow calls an action or reusable workflow
- **WHEN** a maintainer inspects its `uses` reference
- **THEN** the reference MUST contain a verified 40-character commit SHA
- **AND** any adjacent version comment MUST identify the tag from which that SHA was resolved

### Requirement: Standardized Dependabot Review Summary

For pull requests authored by `dependabot[bot]`, the workflow MUST create or replace one standardized review comment containing the general review conclusion, calculated risk, dependency name and version, release age, and a link to the workflow logs.

The summary MUST state whether automated approval was eligible. Unknown, malformed, or unavailable review signals MUST be represented as requiring manual review rather than as passing signals.

#### Scenario: Safe update receives an auditable summary
- **GIVEN** a Dependabot pull request has completed both reusable review jobs
- **WHEN** the reporting job runs
- **THEN** it MUST publish one standardized comment containing the review evidence and approval outcome
- **AND** a rerun MUST replace that comment rather than add a duplicate

#### Scenario: Missing release metadata requires manual review
- **GIVEN** a Dependabot pull request has an unknown or unavailable release age
- **WHEN** the reporting job composes the review summary
- **THEN** the summary MUST identify the missing release age
- **AND** it MUST state that manual review is required

#### Scenario: Non-Dependabot pull request is not commented on
- **GIVEN** a pull request was not authored by `dependabot[bot]`
- **WHEN** the Dependencies workflow runs
- **THEN** it MUST NOT post a Dependabot review summary

### Requirement: Guarded Dependabot Auto-Approval

The workflow MUST auto-approve a Dependabot pull request only when all of the following conditions hold:

- The general dependency review conclusion is successful.
- The calculated risk is not high.
- The release age is known, parseable, and at least 24 hours.
- No review from a non-bot user is currently in the `CHANGES_REQUESTED` state.

The approval job MUST fail closed when any required condition is missing, malformed, or false. It MUST NOT merge the pull request, enable auto-merge, dismiss reviews, alter branch protection, or bypass required status checks or code-owner review.

#### Scenario: Eligible update is approved
- **GIVEN** a Dependabot pull request has a successful dependency review
- **AND** its calculated risk is low or medium
- **AND** its release age is known and at least 24 hours
- **AND** no non-bot user has requested changes
- **WHEN** the approval job runs
- **THEN** the workflow MUST submit an approving review containing the decision evidence
- **AND** it MUST NOT merge or enable auto-merge for the pull request

#### Scenario: Major or vulnerable update is not approved
- **GIVEN** a Dependabot pull request is classified as high risk or its dependency review reports a vulnerability failure
- **WHEN** the approval job evaluates the pull request
- **THEN** it MUST NOT submit an approving review
- **AND** the pull request MUST remain available for manual review

#### Scenario: New release is not approved
- **GIVEN** a Dependabot pull request updates to a release less than 24 hours old
- **WHEN** the approval job evaluates the pull request
- **THEN** it MUST NOT submit an approving review

#### Scenario: Human request for changes blocks approval
- **GIVEN** a non-bot reviewer has an active `CHANGES_REQUESTED` review on a Dependabot pull request
- **WHEN** the approval job evaluates existing reviews
- **THEN** it MUST NOT submit an approving review

#### Scenario: Invalid approval signal fails closed
- **GIVEN** risk, review conclusion, or release-age output is missing or malformed
- **WHEN** the approval job evaluates eligibility
- **THEN** it MUST NOT submit an approving review
- **AND** the workflow logs or summary MUST identify that manual review is required

### Requirement: Existing Dependabot Proposal Configuration

The implementation MUST retain the existing `.github/dependabot.yml` configuration for weekly `gomod` and `github-actions` updates and MUST NOT duplicate that configuration in the dependency review workflow.

#### Scenario: Existing update proposal policy remains stable
- **GIVEN** `.github/dependabot.yml` already configures weekly root-directory updates for `gomod` and `github-actions`
- **WHEN** dependency review automation is added
- **THEN** the existing Dependabot configuration MUST remain unchanged
<!-- scaffolded by uf v0.17.0 -->
