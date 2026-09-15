<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Implementation Preconditions

- [x] 1.1 Confirm work is on the `opsx/automate-dependency-update` feature branch and that these planning artifacts have been committed and pushed before editing implementation files.
- [ ] 1.2 Re-resolve `complytime/org-infra` v0.8.0, `peter-evans/create-or-update-comment` v5.0.0, and `actions/github-script` v9.0.0 through the GitHub API, dereference annotated tags, and record the verified 40-character commit SHAs.
- [ ] 1.3 Compare the selected org-infra revision's `workflow_call` outputs with the names consumed by the design; stop and update the spec if the upstream contract has changed.

## 2. Automated Regression Tests

- [ ] 2.1 Create an automated offline regression test for `.github/workflows/ci_dependencies.yml` that validates workflow structure and evaluates fixtures for eligible updates, high-risk or vulnerable updates, releases younger than 24 hours, missing or malformed outputs, failed or cancelled reviewer jobs, non-Dependabot authors, idempotent reporting, active human vetoes, later approvals that clear earlier vetoes, dismissed vetoes, and comments that do not clear active vetoes.
- [ ] 2.2 Run the new regression test before creating the workflow and confirm it fails for the missing implementation.

## 3. Dependency Review Workflow

- [ ] 3.1 Create `.github/workflows/ci_dependencies.yml` with the required purpose header, `main` push and pull-request triggers, explicit read-only workflow permissions, the 24-hour minimum release age, and ref-scoped concurrency.
- [ ] 3.2 Add the general and Dependabot-specific reusable review jobs using the implementation-time-verified org-infra SHA and accurate version comments.
- [ ] 3.3 Add the Dependabot-only reporting job with job-scoped comment permission, verified `create-or-update-comment` pin, `always()` failure handling, output-backed review table, workflow-log provenance, idempotent replacement, eligibility-only wording, and explicit manual-review text for failed, cancelled, missing, malformed, or unsafe signals.
- [ ] 3.4 Add the Dependabot-only approval job with only `pull-requests: write`, a verified `github-script` pin, environment-delivered reviewer outputs, duplicate fail-closed validation, per-reviewer latest-effective-state evaluation, and an evidence-bearing approval review.
- [ ] 3.5 Verify the workflow contains no merge, auto-merge, review-dismissal, branch-protection, or required-check mutation and that `.github/dependabot.yml` is unchanged from the implementation start commit.

## 4. Static and Behavioral Verification

- [ ] 4.1 Run the automated offline regression test and confirm every specified structure and decision-path fixture passes.
- [ ] 4.2 Run a workflow-aware YAML linter at an explicitly pinned version against `.github/workflows/ci_dependencies.yml`; resolve all syntax, expression, event, permission, and reusable-workflow-call errors.
- [ ] 4.3 Audit every `uses:` value against the tag-resolution evidence and the CI convention pack, including 40-character SHAs and accurate version comments.
- [ ] 4.4 Confirm the reporting and approval jobs never check out or execute pull-request code while holding write permissions.

## 5. Project and Governance Gates

- [ ] 5.1 Run the CI-equivalent commands derived from `.github/workflows/ci.yml`, including `go vet`, the pinned `govulncheck`, Homebrew cask script tests, race-enabled Go tests with coverage, coverage-ratchet enforcement, and binary build; record all results.
- [ ] 5.2 Run `openspec validate automate-dependency-update` and resolve every validation error.
- [ ] 5.3 Verify the final diff still satisfies all four constitution assessments: durable review artifacts, standalone runtime behavior, observable decision evidence, and locally reproducible structural checks.
- [ ] 5.4 Complete the documentation gate by confirming whether README.md, AGENTS.md, or GoDoc changes are needed; record that the website issue gate is exempt because this is CI-only behavior.
- [ ] 5.5 Run the review council on the final diff and resolve all REQUEST CHANGES findings before preparing the pull request.
- [ ] 5.6 Record an activation checklist in the pull request for a normal pull request and the next representative Dependabot pull requests, covering standardized comment output, safe approval, unsafe manual-review outcomes, and preserved branch protection without auto-merge.
<!-- scaffolded by uf v0.17.0 -->
<!-- spec-review: passed -->
