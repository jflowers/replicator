# Spec: CI Checks Workflow

**Change**: adopt-org-infra-ci

## ADDED Requirements

### Requirement: MegaLinter CI Workflow

The repository MUST include a GitHub Actions workflow at
`.github/workflows/ci_checks.yml` that calls the `complytime/org-infra`
reusable CI workflow. The workflow MUST trigger on push to `main` and on
pull requests targeting `main`. The `uses:` reference MUST be SHA-pinned
per CI-001. Permissions MUST follow least-privilege (CI-020).

#### Scenario: PR opened against main
- **GIVEN** a pull request is opened targeting the `main` branch
- **WHEN** the CI pipeline runs
- **THEN** the `Standardized CI / Run linters` check MUST execute
- **AND** MegaLinter MUST scan only changed files
- **AND** PR title MUST be validated against Conventional Commits format

#### Scenario: Push to main
- **GIVEN** a commit is pushed to the `main` branch
- **WHEN** the CI pipeline runs
- **THEN** MegaLinter MUST scan the full codebase
- **AND** PR title validation MUST be skipped (not a PR context)

#### Scenario: Dependabot PR
- **GIVEN** a pull request is opened by `dependabot[bot]`
- **WHEN** the CI pipeline runs
- **THEN** MegaLinter MUST run normally
- **AND** PR title validation MUST be skipped

### Requirement: MegaLinter Configuration

The repository MUST include a `.mega-linter.yml` configuration file.
The configuration MUST preserve the org-standard linter set. The
configuration SHOULD exclude project-specific directories (`.opencode`,
`.claude`, `.uf`, `.specify`, `dist`) from style linter scanning.

#### Scenario: Style linters skip agent directories
- **GIVEN** the `.opencode/` directory contains AI agent definitions
- **WHEN** MegaLinter runs
- **THEN** style linters (markdownlint, yamllint) MUST NOT report
  findings in excluded directories

The exclusions MUST use per-linter `*_FILTER_REGEX_EXCLUDE` overrides.
They MUST NOT use `ADDITIONAL_EXCLUDED_DIRECTORIES`, because that global
setting would also prevent security linters such as betterleaks and KICS
from scanning the excluded directories.

### Requirement: Release Preflight Gating

The release workflow preflight MUST verify that the
`Standardized CI / Run linters` check passed on HEAD before allowing
a release. The `ci_checks` input to the reusable preflight workflow
MUST include both `"Build and Test"` and
`"Standardized CI / Run linters"`.

#### Scenario: Release with passing linter check
- **GIVEN** both "Build and Test" and "Standardized CI / Run linters"
  checks passed on HEAD
- **WHEN** a release is triggered via workflow_dispatch
- **THEN** the preflight MUST pass
- **AND** the release MUST proceed

#### Scenario: Release with failing linter check
- **GIVEN** "Build and Test" passed but "Standardized CI / Run linters"
  failed on HEAD
- **WHEN** a release is triggered via workflow_dispatch
- **THEN** the preflight MUST fail
- **AND** the release MUST NOT proceed

### Requirement: Branch Protection

The `.github/settings.yml` MUST include `Standardized CI / Run linters`
in the required status checks for the `main` branch, alongside the
existing `Build and Test` check.

#### Scenario: PR missing linter check
- **GIVEN** a pull request where "Build and Test" passed
- **AND** "Standardized CI / Run linters" has not run or failed
- **WHEN** a merge is attempted
- **THEN** the merge MUST be blocked by branch protection

## MODIFIED Requirements

_None._

## REMOVED Requirements

_None._
