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

- [ ] 1.1 Confirm work is on the `opsx/automate-dependency-update` feature branch and that these planning artifacts have been committed and pushed before editing implementation files.
- [ ] 1.2 Re-resolve `complytime/org-infra` v0.8.0, `peter-evans/create-or-update-comment` v5.0.0, and `actions/github-script` v9.0.0 through the GitHub API, dereference annotated tags, and record the verified 40-character commit SHAs.
- [ ] 1.3 Compare the selected org-infra revision's `workflow_call` outputs with the names consumed by the design; stop and update the spec if the upstream contract has changed.

## 2. Dependency Review Workflow

- [ ] 2.1 Create `.github/workflows/ci_dependencies.yml` with the required purpose header, `main` push and pull-request triggers, explicit read-only workflow permissions, the 24-hour minimum release age, and ref-scoped concurrency.
- [ ] 2.2 Add the general and Dependabot-specific reusable review jobs using the implementation-time-verified org-infra SHA and accurate version comments.
- [ ] 2.3 Add the Dependabot-only reporting job with job-scoped comment permission, verified `create-or-update-comment` pin, output-backed review table, workflow-log provenance, idempotent replacement, and explicit manual-review text for unknown or unsafe signals.
- [ ] 2.4 Add the Dependabot-only approval job with only `pull-requests: write`, a verified `github-script` pin, environment-delivered reviewer outputs, duplicate fail-closed validation, a non-bot `CHANGES_REQUESTED` lookup, and an evidence-bearing approval review.
- [ ] 2.5 Verify the workflow contains no merge, auto-merge, review-dismissal, branch-protection, or required-check mutation and that `.github/dependabot.yml` is unchanged from the implementation start commit.

## 3. Static and Behavioral Verification

- [ ] 3.1 [P] Run a workflow-aware YAML linter at an explicitly pinned version against `.github/workflows/ci_dependencies.yml`; resolve all syntax, expression, event, permission, and reusable-workflow-call errors.
- [ ] 3.2 [P] Audit every `uses:` value against the tag-resolution evidence and the CI convention pack, including 40-character SHAs and accurate version comments.
- [ ] 3.3 Exercise or inspect each specified decision path: eligible low/medium-risk update, high-risk or vulnerable update, release younger than 24 hours, unknown/malformed outputs, non-Dependabot author, rerun comment replacement, and human `CHANGES_REQUESTED` veto.
- [ ] 3.4 Confirm the reporting and approval jobs never check out or execute pull-request code while holding write permissions.

## 4. Project and Governance Gates

- [ ] 4.1 Run the CI-equivalent commands derived from `.github/workflows/ci.yml`, including `go vet`, the pinned `govulncheck`, Homebrew cask script tests, race-enabled Go tests with coverage, coverage-ratchet enforcement, and binary build; record all results.
- [ ] 4.2 Run `openspec validate automate-dependency-updates` and resolve every validation error.
- [ ] 4.3 Verify the final diff still satisfies all four constitution assessments: durable review artifacts, standalone runtime behavior, observable decision evidence, and locally reproducible structural checks.
- [ ] 4.4 Complete the documentation gate by confirming whether README.md, AGENTS.md, or GoDoc changes are needed; record that the website issue gate is exempt because this is CI-only behavior.
- [ ] 4.5 Run the review council on the final diff and resolve all REQUEST CHANGES findings before preparing the pull request.
- [ ] 4.6 Record an activation checklist in the pull request for a normal pull request and the next representative Dependabot pull requests, covering standardized comment output, safe approval, unsafe manual-review outcomes, and preserved branch protection without auto-merge.
<!-- scaffolded by uf v0.17.0 -->
