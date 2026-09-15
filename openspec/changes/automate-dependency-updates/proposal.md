## Why

Replicator already asks Dependabot to propose weekly Go module and GitHub Actions updates, but it has no repository workflow that evaluates dependency changes consistently or automates approval of low-risk updates. Maintainers must interpret each update manually, which delays routine maintenance and produces inconsistent review evidence.

Issue #38 requests adoption of the organization-owned dependency review workflows. The existing `.github/dependabot.yml` satisfies the update-proposal portion of that issue and will remain unchanged; this change closes the remaining review and guarded-approval gap.

## What Changes

- Add a `ci_dependencies.yml` consumer workflow that invokes the pinned org-infra general dependency reviewer and Dependabot-specific risk reviewer.
- Publish a standardized, replaceable review summary on Dependabot pull requests.
- Auto-approve Dependabot pull requests only when dependency review succeeds, risk is not high, release age is known and at least 24 hours, and no human has requested changes.
- Fail closed to manual review when any approval signal is missing, invalid, or unsafe.
- Keep merging, branch protection, required status checks, and code-owner review outside the automation; approval MUST NOT enable or perform auto-merge.
- Preserve the existing weekly `gomod` and `github-actions` Dependabot configuration without modification.

## Capabilities

### New Capabilities
- `dependency-review-automation`: Orchestrate organization-owned dependency analysis, standardized Dependabot review reporting, and constrained auto-approval for safe updates.

### Modified Capabilities
- None. The previously delivered `ci-security-hygiene` capability already defines and supplies the required Dependabot update configuration.

### Removed Capabilities
- None.

## Impact

- Adds `.github/workflows/ci_dependencies.yml` and no production Go code.
- Uses GitHub Actions token write permission only in jobs that comment on or approve Dependabot pull requests; reusable analysis jobs remain read-only.
- Adds workflow runs for pushes and pull requests targeting `main`, matching the canonical organization workflow.
- Depends operationally on SHA-pinned reusable workflows from `complytime/org-infra` and pinned third-party actions.
- Requires repository settings that permit GitHub Actions to create pull-request approvals; unavailable permissions degrade to a visible failed workflow and manual review rather than bypassing review.
- Complements existing `govulncheck`, branch protection, and code-owner requirements without replacing or weakening them.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The workflow communicates through GitHub pull-request reviews, comments, and job outputs. These are durable artifacts that can be interpreted independently by maintainers and automation; no runtime coupling is added to the Replicator binary or MCP tools.

### II. Composability First

**Assessment**: PASS

Dependency review is repository maintenance automation, not a runtime dependency. Replicator remains independently buildable, installable, and usable if GitHub Actions or org-infra workflows are unavailable.

### III. Observable Quality

**Assessment**: PASS

The workflow exposes review conclusions, risk, dependency identity, version, and release age through GitHub job outputs and a standardized pull-request comment with links to logs. Unsafe or unavailable signals produce an explicit manual-review outcome.

### IV. Testability

**Assessment**: PASS

Implementation verification will statically validate workflow structure, event filters, full-SHA pins, permissions, output wiring, and fail-closed approval predicates. Existing local CI-equivalent checks will confirm no regression to the Go project, while post-activation checks will verify GitHub-hosted comment and approval side effects that cannot be exercised without GitHub.
<!-- scaffolded by uf v0.17.0 -->
