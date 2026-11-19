# Tasks: Bump Both org-infra Reviewer Pins to v0.9.0

**Change**: bump-org-infra-reviewer-pins

<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
-->

## 1. Resolve and Verify the Target SHA

- [x] 1.1 Re-derive the v0.9.0 SHA from `complytime/org-infra`
  (CI-002): resolve the `v0.9.0` tag and inspect the API object type —
  when it is `tag` (annotated), dereference to the commit SHA; when it is
  `commit` (lightweight), use the commit directly. Verify that BOTH files
  being pinned exist at the resolved SHA (`reusable_deps_reviewer.yml` and
  `reusable_dependabot_reviewer.yml`), and confirm the consumer output
  contract still holds (`review_conclusion`; `risk_level`, `dep_name`,
  `dep_version`, `release_age_hours`; v0.9.0 also adds `org_precedent` and
  `precedent_detail` additively). Record the resolved full 40-character SHA
  and the verification evidence. Do not copy the SHA from issue prose alone.

## 2. Apply the Lockstep Bump

- [x] 2.1 [P] In `.github/workflows/ci_dependencies.yml`, update both
  `uses:` references — line 29 (`call_deps_reviewer`,
  `reusable_deps_reviewer.yml@`) and line 37 (`call_dependabot_reviewer`,
  `reusable_dependabot_reviewer.yml@`) — to the SHA resolved in task 1.1
  with the `# v0.9.0` comment. Leave `ci_checks.yml` and `release.yml`
  references on v0.8.0.
- [x] 2.2 [P] In `internal/workflowtest/ci_dependencies_test.go:19`,
  update `orgInfraSHA` to the SHA resolved in task 1.1. Keep the full
  40-character SHA assertions and `# vX.Y.Z` comment checks intact; do not
  loosen or regex-relax them.

## 3. Verification and Documentation

- [x] 3.1 Run `make check` and confirm
  `TestCIDependenciesWorkflow_StructureEnforcesGuardedApproval` passes with
  no assertion weakened.
- [x] 3.2 Confirm both reviewer references are pinned to the same full
  40-character SHA with accurate version comments (CI-001, CI-003), and
  that the out-of-scope org-infra references are unchanged.
- [x] 3.3 Run `openspec validate bump-org-infra-reviewer-pins` and resolve
  every validation error (carry over the predecessor's artifact gate).
- [x] 3.4 Verify constitution alignment per the proposal: I (N/A — no
  artifact interfaces), II (PASS — no runtime dependency), III (PASS —
  machine-verifiable pins and observable fail-closed reporting), IV (PASS —
  isolated workflow regression test).
- [x] 3.5 Update `docs/ci-workflow-security.md` to record that a failed
  dependency scan now surfaces as `failure` and halts Dependabot
  auto-approval (fail-closed → manual review).
- [x] 3.6 Confirm the shared `orgInfraSHA` constant is an intentional
  lockstep invariant (documented in design R2); no constant split is
  planned, because independent per-reference pins would risk version drift.
- [x] 3.7 Note the fail-closed behaviour change in the PR description.
