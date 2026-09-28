# Tasks: Adopt org-infra Reusable CI Workflows

**Change**: adopt-org-infra-ci

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

## 1. Add CI Checks Workflow and MegaLinter Config

- [x] 1.1 [P] Create `.github/workflows/ci_checks.yml` calling
  `complytime/org-infra/.github/workflows/reusable_ci.yml@0c784711926c9864f027ec565fd7c06a382d80f8`
  (v0.7.1). Trigger on push to main and PRs to main. Follow the canonical
  consumer pattern from `unbound-force/unbound-force`. SHA-pin all action
  references per CI-001. Set least-privilege permissions per CI-020. Include a
  header comment block per CI-011. Check whether `reusable_ci.yml` defines an
  internal concurrency group before adding a caller-level one (see design R5).
- [x] 1.2 [P] Create `.mega-linter.yml` at the repository root. Preserve the
  org-standard `ENABLE_LINTERS` list. Add per-linter style exclusions for
  `.opencode`, `.claude`, `.uf`, `.specify`, `dist`, and `vendor` without
  excluding those directories from security scanners.

## 2. Update Release and Branch Protection

- [x] 2.1 [P] Update `.github/workflows/release.yml`: add `"Standardized CI / Run linters"` to the `ci_checks` array in the preflight job's `with:` block, resulting in `ci_checks: '["Build and Test", "Standardized CI / Run linters"]'`.
- [x] 2.2 Update `.github/settings.yml`: add `"Standardized CI / Run linters"` to the `required_status_checks.contexts` array alongside `"Build and Test"`. **Depends on task 3.1** — branch protection must not gate on the new check until lint issues are resolved. See design R6 for rollback procedure if the check blocks the repo.

## 3. Fix Existing Lint Issues

- [x] 3.1 Inspect likely MegaLinter findings and fix issues that would cause the `Standardized CI / Run linters` check to fail (errors). Include Go golangci-lint findings in the triage scope. Defer warnings and non-blocking findings to follow-up issues. Completion criterion: the CI check passes on the PR branch with zero blocking errors.

## 4. Verification

- [x] 4.1 Verify all new and modified workflow files have SHA-pinned action references with version comments (CI-001).
- [x] 4.2 Verify permissions follow least-privilege principle (CI-020).
- [x] 4.3 Verify constitution alignment: no MCP tools affected (I), existing CI continues independently (II), new checks produce machine-readable output (III), no test infrastructure changes (IV).
- [x] 4.4 Run `make check` to confirm existing CI-equivalent checks still pass.
- [x] 4.5 Update AGENTS.md to document the new `Standardized CI / Run linters` check alongside the existing `Build and Test` check documentation.

<!-- spec-review: passed -->
<!-- code-review: passed -->
