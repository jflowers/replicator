<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. AgentKit smart-sync core (TDD)

- [ ] 1.1 Write failing unit tests in `internal/agentkit/agentkit_test.go`: `TestScaffold_SmartSyncUpdatesDriftedFile` (pre-create kit file with different bytes, run sync mode, assert content equals embedded and action is `updated`), `TestScaffold_SmartSyncSkipsIdenticalFile` (pre-create with exact embedded bytes, assert not rewritten and action is `skipped`), `TestScaffold_SkipExistingModePreservesDriftedFile` (pre-create differing file, run skip-existing mode, assert content unchanged and action is `skipped`). Update any existing `Scaffold` callers/tests to the new signature so the package compiles for the red run where behavior allows.
- [ ] 1.2 Implement in `internal/agentkit/agentkit.go`: add `ScaffoldMode` type with `ScaffoldModeSync` and `ScaffoldModeSkipExisting`, change signature to `Scaffold(targetDir string, mode ScaffoldMode)`, read destination bytes and compare with `bytes.Equal` against embedded content, report `created`/`skipped`/`updated` (drop `overwritten` — force mode is removed), skip rewrite for identical files, overwrite drifted files in sync mode, preserve all files in skip-existing mode. Keep error paths (stat errors other than `ErrNotExist`, read/write failures) wrapped with context as today.

## 2. CLI wiring (TDD)

- [ ] 2.1 Write failing tests in `cmd/replicator/init_test.go`: `TestRunInit_UpdatesDriftedFiles` (pre-create differing kit file, run default init, assert file converges to embedded content), `TestRunInit_SkipExistingFlag` (pre-create differing file, run with `--skip-existing`, assert preserved and missing files still created), `TestRunInit_ForceIsNoOp` (run with `--force`, assert behavior identical to default smart-sync). Rewrite `TestRunInit_AgentKitSkipsExisting` to assert skip only for byte-identical files or `--skip-existing`, and remove/replace `TestRunInit_ForceOverwrites` accordingly. Keep `cells.json` and DCP assertions in `TestRunInit_AlreadyInitialized` passing unchanged.
- [ ] 2.2 Implement in `cmd/replicator/init.go`: add `--skip-existing` bool flag wired to `ScaffoldModeSkipExisting`, mark `--force` deprecated via `cmd.Flags().MarkDeprecated` (notice text: smart-sync is now the default), default mode `ScaffoldModeSync`, render the `updated` action with the warn style (reuse the previous `overwritten` rendering), update `Long` help text to describe smart-sync semantics (create missing / skip identical / update drifted; embedded content canonical; `--skip-existing` for customized repos; `cells.json` and DCP config behavior unchanged).

## 3. Documentation and website

- [ ] 3.1 [P] Add a CHANGELOG.md entry under Unreleased describing the smart-sync behavior change, the new `--skip-existing` flag, and the `--force` deprecation.
- [ ] 3.2 [P] Update the embedded `internal/agentkit/content/skills/replicator-cli/SKILL.md` Init Flags section to document smart-sync default, `--skip-existing`, and deprecated `--force` (mirrored pair — the `.opencode/` copy is synced by the dogfood task 4.2). Check README.md for init skip/force documentation and update if present.
- [ ] 3.3 [P] Create the website documentation issue in `unbound-force/website` via `gh issue create` covering the changed `replicator init` default behavior, new flag, and deprecation, per the Website Documentation Gate.

## 4. Verification and dogfood remediation

- [ ] 4.1 Run the CI-equivalent checks locally (derive exact commands from `.github/workflows/` via the pre-flight skill; at minimum `make check`) and fix any failures.
- [ ] 4.2 Dogfood: rebuild and install the binary (`make install`), run the new `replicator init` at the repo root, verify the 6 pre-existing drifted `.opencode/` files plus the `replicator-cli` skill mirror are reported `updated` and byte-identical to their embedded copies (15-pair diff clean), and commit the synced files as `fix(agentkit): sync scaffolded .opencode copies via smart-sync`.
- [ ] 4.3 Re-run the CI-equivalent checks after the drift sync (drift detection tests plus full suite) and fix any failures.
- [ ] 4.4 Verify constitution alignment: confirm no protected gate values were modified (coverage thresholds, severity definitions, CI flags), confirm tests use `t.TempDir()` isolation with stdlib assertions only, and confirm `ScaffoldResult` remains self-describing JSON with the new `updated` action.
