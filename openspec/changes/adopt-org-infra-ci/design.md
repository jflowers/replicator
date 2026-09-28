# Design: Adopt org-infra Reusable CI Workflows

**Change**: adopt-org-infra-ci

## Context

Replicator's CI pipeline currently provides Go-specific checks (vet,
govulncheck, tests, coverage ratchets, build) via a single "Build and
Test" job in `.github/workflows/ci.yml`. It lacks cross-language
linting (YAML, Markdown, shell, Dockerfiles), GitHub Actions workflow
validation, secrets scanning, infrastructure-as-code security scanning,
and PR title enforcement.

The `complytime/org-infra` repository provides a reusable CI workflow
(`reusable_ci.yml`) at v0.7.1 that bundles MegaLinter v9.6.0 with 12
linters and commitlint-based PR title validation. This workflow is
already adopted by `unbound-force/unbound-force` and `dewey`. The
canonical consumer pattern is a thin `ci_checks.yml` workflow that
calls the reusable workflow with no inputs (the reusable workflow
accepts none).

## Goals / Non-Goals

### Goals
- Add MegaLinter-based linting as a new CI check running in parallel
  with the existing "Build and Test" job
- Enforce Conventional Commits on PR titles via commitlint
- Follow the canonical consumer pattern established by
  `unbound-force/unbound-force`
- Provide a local `.mega-linter.yml` config that excludes
  replicator-specific directories from style linters
- Update the release preflight to gate on the new linter check
- Update branch protection to require the new check

### Non-Goals
- Replacing the existing `ci.yml` -- it continues to run independently
- Adding golangci-lint configuration -- the org-infra fallback config
  will be used; local customization is a future change if needed
- Fixing all existing lint issues surfaced by the first MegaLinter
  run -- these will be triaged separately
- Adding local `commitlint.config.js`, `.yamllint.yml`, or
  `ruff.toml` -- the org-infra fallback mechanism supplies these

## Decisions

### D1: Separate workflow file, not integrated into ci.yml

The new CI check lives in a separate `ci_checks.yml` file rather than
being added as a job in the existing `ci.yml`. This matches the
canonical pattern, keeps the existing "Build and Test" check name
stable, and allows the two pipelines to evolve independently.

### D2: Rely on org-infra fallback for most configs

The reusable workflow has a built-in fallback mechanism: it checks for
local config files (`.mega-linter.yml`, `.golangci.yml`, `.yamllint.yml`,
`commitlint.config.js`, `ruff.toml`) and copies missing ones from
org-infra at the pinned SHA. Replicator will provide only a local
`.mega-linter.yml` to exclude project-specific directories. All other
configs will use the org-infra defaults via fallback.

### D3: Local .mega-linter.yml with style-linter exclusions

Replicator has directories that should not be scanned by style linters:
`.opencode` (AI agent definitions), `.claude` (AI config), `.uf`
(replicator local data), `.specify` (spec framework), and `dist`
(build output). A local `.mega-linter.yml` will exclude these through
per-linter filters while preserving security scanner coverage and the
org-standard linter set.

### D4: CI check name follows nested job naming

GitHub Actions names nested reusable workflow jobs as
`<caller-job-name> / <reusable-job-name>`. The caller job is named
"Standardized CI" and the reusable job is named "Run linters",
producing the check name `Standardized CI / Run linters`. This is the
name used in the release preflight `ci_checks` array and branch
protection.

### D5: settings.yml uses `_extends: .github` inheritance

The `.github/settings.yml` uses `_extends: .github` which inherits
settings from the org-level `.github` repository. Local
`required_status_checks.contexts` additions are additive — they extend
(not override) the inherited settings. The new
`Standardized CI / Run linters` check is added to the local contexts
array alongside the existing `Build and Test` check.

### D6: SHA-pinned action references per CI-001

All `uses:` references in the new workflow will use full 40-character
commit SHAs with version comments, per CI convention pack rule CI-001.
The org-infra reference will use SHA `0c784711926c9864f027ec565fd7c06a382d80f8`
(v0.7.1).

## Risks / Trade-offs

### R1: First-run lint failures

The first MegaLinter run will likely surface existing issues in YAML,
Markdown, shell scripts, and GitHub Actions workflows. These must be
triaged: some fixed in this change, others deferred to follow-up
issues. This is acceptable -- the goal is to establish the pipeline,
not achieve zero warnings on day one.

### R2: CI time increase

Adding MegaLinter adds a parallel job that may take 2-5 minutes. Since
it runs concurrently with "Build and Test", wall-clock impact on PR
feedback time should be minimal.

### R3: Dependency on org-infra fallback configs

If org-infra changes its default configs in a future version, the
changes propagate automatically to replicator. This is by design --
it keeps repos in sync with org standards -- but could surface
unexpected new lint errors after a version bump. The SHA-pinning
mitigates this: config changes only arrive when the pinned SHA is
explicitly updated.

### R4: golangci-lint overlap

MegaLinter includes `GO_GOLANGCI_LINT` which overlaps with the
existing `go vet` in ci.yml. However, golangci-lint is a superset of
go vet and provides additional checks. The overlap is harmless (both
must pass) and the golangci-lint configuration from org-infra provides
a standardized rule set.

### R5: Concurrency group nesting hazard

If `reusable_ci.yml` defines an internal concurrency group, adding a
caller-level `concurrency` block in `ci_checks.yml` could cause
conflicts (the org hit this in PR #456 with the release preflight
workflow). The implementer MUST check whether the reusable CI workflow
defines its own concurrency group before adding one at the caller
level. If it does, omit the caller-level concurrency block.

### R6: No rollback procedure

Adding `Standardized CI / Run linters` to branch protection and the
release preflight creates a hard gate. If MegaLinter surfaces blocking
issues that cannot be quickly resolved, the repo could enter a state
where PRs cannot merge and releases cannot ship. Rollback order:
1. Remove `Standardized CI / Run linters` from `settings.yml`
   required checks
2. Remove it from `release.yml` `ci_checks` array
3. Optionally delete `ci_checks.yml` and `.mega-linter.yml`

This ensures the repo is never blocked by a non-functional check.
