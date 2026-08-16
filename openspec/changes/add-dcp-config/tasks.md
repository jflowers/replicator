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

## 1. Add `ScaffoldDCP()` to agentkit

- [x] 1.1 Write failing tests for `ScaffoldDCP()` in `internal/agentkit/agentkit_test.go`: fresh directory (creates `dcp.jsonc`), existing config with `protectTags` (skips), existing config without `protectTags` (replaces with canonical content), `.dcp.json` alias (operates on `.json` not `.jsonc`), both files exist (prefers `.jsonc`), `.opencode/` directory creation
- [x] 1.2 Implement `ScaffoldDCP(targetDir string) (ScaffoldResult, error)` in `internal/agentkit/agentkit.go`: check for `.opencode/dcp.jsonc` then `.opencode/dcp.json`, use `strings.Contains` for `protectTags` detection, create/skip/update with appropriate action string, DCP config content matches design spec D5

## 2. Integrate into `replicator init`

- [x] 2.1 [P] Add `dcp.jsonc` assertions to `cmd/replicator/init_test.go`: verify `TestRunInit_FreshDirectory` includes `dcp.jsonc` in spot-check, verify `TestRunInit_AlreadyInitialized` skips `dcp.jsonc` on re-run
- [x] 2.2 [P] Call `ScaffoldDCP()` in `runInit()` in `cmd/replicator/init.go`: call after `Scaffold()`, render result with same styled output (green/dim/yellow)

## 3. Add `checkDCPConfig()` to doctor

- [x] 3.1 Write failing tests for `checkDCPConfig()` in `internal/doctor/checks_test.go`: DCP config present with protectTags (pass), no protect-tagged commands (pass), protect-tagged commands but no DCP config (warn), DCP config missing protectTags (warn)
- [x] 3.2 Implement `checkDCPConfig(projectDir string) CheckResult` in `internal/doctor/checks.go`: accept explicit directory parameter, scan `.opencode/commands/*.md` for `<protect>` tags, check `.opencode/dcp.jsonc` then `.opencode/dcp.json` for `protectTags`, return warn (not fail) when missing
- [x] 3.3 Update `Run()` signature to `Run(store *db.Store, cfg *config.Config, projectDir string)`, register `checkDCPConfig(projectDir)` as 5th check, update caller in `cmd/replicator/doctor.go` to pass `os.Getwd()`, update `TestRun_AllChecks` to expect 5 results with `dcp_config` name

## 4. Verification

- [x] 4.1 Run `make check` and `make check-coverage` to verify all tests pass and coverage ratchets are maintained
- [x] 4.2 Verify constitution alignment: `ScaffoldDCP()` is independently callable (Composability), returns machine-parseable `ScaffoldResult` (Observable Quality), testable with `t.TempDir()` (Testability), `checkDCPConfig()` uses `t.TempDir()` with no external services (Testability)
<!-- scaffolded by uf vdev -->
<!-- spec-review: passed -->
<!-- code-review: passed -->
