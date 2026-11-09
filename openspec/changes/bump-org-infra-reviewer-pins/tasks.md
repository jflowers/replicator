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

## 1. Resolve the Target SHA

- [ ] 1.1 Re-derive the v0.9.0 SHA from `complytime/org-infra`
  (CI-002): resolve the `v0.9.0` tag, dereference it to a commit SHA, and
  confirm `reusable_dependabot_reviewer.yml` exists at that commit. Record
  the resolved full 40-character SHA. Do not copy the SHA from issue prose
  alone.

## 2. Apply the Lockstep Bump

- [ ] 2.1 [P] In `.github/workflows/ci_dependencies.yml`, update both
  `uses:` references — line 29 (`call_deps_reviewer`,
  `reusable_deps_reviewer.yml@`) and line 37 (`call_dependabot_reviewer`,
  `reusable_dependabot_reviewer.yml@`) — to the resolved v0.9.0 SHA with
  the `# v0.9.0` comment. Leave `ci_checks.yml` and `release.yml`
  references on v0.8.0.
- [ ] 2.2 [P] In `internal/workflowtest/ci_dependencies_test.go:19`,
  update `orgInfraSHA` to the resolved v0.9.0 SHA. Keep the full
  40-character SHA assertions and `# vX.Y.Z` comment checks intact; do not
  loosen or regex-relax them.

## 3. Verification

- [ ] 3.1 Run `make check` and confirm
  `TestCIDependenciesWorkflow_StructureEnforcesGuardedApproval` passes with
  no assertion weakened.
- [ ] 3.2 Confirm both reviewer references are pinned to the same full
  40-character SHA with accurate version comments, and that the
  out-of-scope org-infra references remain v0.8.0.
- [ ] 3.3 Verify constitution alignment per the proposal: I (N/A — no
  artifact interfaces), II (PASS — no runtime dependency), III (PASS —
  machine-verifiable pins and observable fail-closed reporting), IV (PASS —
  isolated workflow regression test).
- [ ] 3.4 Note the fail-closed behaviour change (failed dependency scan
  halts Dependabot auto-approval) in the PR description.
