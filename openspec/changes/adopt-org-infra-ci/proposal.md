# Proposal: Adopt org-infra Reusable CI Workflows

**Change**: adopt-org-infra-ci
**Issue**: [#25](https://github.com/unbound-force/replicator/issues/25)
**Status**: proposal

## Why

Replicator's CI pipeline (`ci.yml`) currently runs only `go vet`,
`govulncheck`, tests, coverage ratchets, and a build step. It lacks
standardized linting (MegaLinter) and PR title validation (Conventional
Commits via commitlint). Other repos in the org -- including
`unbound-force/unbound-force` and `dewey` -- have already adopted the
`complytime/org-infra` reusable CI workflow that provides these
capabilities. Replicator should follow suit for consistency, supply-chain
hygiene (zizmor, KICS, gitleaks), and PR quality enforcement.

The reusable workflow at `complytime/org-infra` v0.7.1 bundles
MegaLinter v9.6.0 with 12 linters (actionlint, zizmor, shellcheck,
hadolint, golangci-lint, markdownlint, yamllint, ruff, protolint,
gitleaks, KICS, ansible-lint) plus commitlint-based PR title
validation. It uses a fallback mechanism: if a consumer repo lacks any
of the managed config files (`.mega-linter.yml`, `.golangci.yml`,
`.yamllint.yml`, `commitlint.config.js`, `ruff.toml`), the workflow
checks them out from org-infra at the pinned SHA.

## What Changes

1. **New workflow**: `.github/workflows/ci_checks.yml` -- a thin
   consumer workflow calling `complytime/org-infra` reusable CI
   workflow, matching the canonical pattern from `unbound-force`.
   SHA-pinned to `0c784711` (v0.7.1).

2. **New config**: `.mega-linter.yml` -- a local override that
   excludes replicator-specific directories (`.opencode`, `.claude`,
   `.uf`, `.specify`, `dist`) from style linters while preserving
   the org-standard linter set.

3. **Release preflight update**: Add the new CI check name
   (`Standardized CI / Run linters`) to the `ci_checks` array in
   `release.yml` so the release preflight verifies both the existing
   "Build and Test" check and the new linter check passed on HEAD.

4. **Branch protection update**: Add `Standardized CI / Run linters`
   to the required status checks in `.github/settings.yml`.

## Capabilities

- MegaLinter runs on every PR (changed files only) and on push to
  main (full codebase scan).
- PR titles are validated against Conventional Commits format via
  commitlint (skipped for Dependabot PRs).
- GitHub Actions workflows are scanned by actionlint and zizmor for
  correctness and supply-chain risks.
- YAML files are validated by yamllint.
- Markdown files are validated by markdownlint.
- Secrets scanning via gitleaks (REPOSITORY_BETTERLEAKS).
- Infrastructure-as-code scanning via KICS (fail on high severity).
- Shell scripts validated by shellcheck.
- Go code scanned by golangci-lint (via MegaLinter, complementing
  the existing `go vet` in the build-and-test job).

## Impact

- **CI**: New parallel job runs alongside existing "Build and Test".
  No changes to existing job. CI time may increase slightly due to
  MegaLinter execution but runs concurrently.
- **Developer workflow**: PR titles must follow Conventional Commits.
  Code must pass MegaLinter checks. Both are already org standards.
- **Release pipeline**: Preflight will verify both checks passed,
  preventing releases with linting failures.
- **Existing code**: May surface existing lint issues on first run.
  These should be addressed as part of this change or triaged.

## Constitution Alignment

### I. Autonomous Collaboration — N/A
CI workflow changes do not affect MCP tool interfaces or inter-agent
communication. No tools are added, modified, or removed.

### II. Composability First — PASS
The new workflow is additive. The existing `ci.yml` continues to run
independently. MegaLinter config falls back to org-infra defaults
when local overrides are absent, maintaining composability with the
org infrastructure.

### III. Observable Quality — PASS
Adds automated quality checks (linting, security scanning, PR title
validation) that produce machine-readable output. MegaLinter produces
structured reports. This directly supports the principle that "all
quality claims MUST be backed by automated, reproducible tests."

### IV. Testability — N/A
CI workflow changes do not affect application testability. No changes
to test infrastructure, database access patterns, or test isolation.
