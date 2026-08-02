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

## 1. Restructure worker.md

All tasks in this group modify the same file (`internal/agentkit/content/agents/worker.md`), so no parallel execution.

- [x] 1.1 Rewrite the checklist to integrate constraints inline. Each numbered step MUST include both the action (tool call) and its boundary constraint as an atomic unit. Use imperative MUST/NEVER language. Specifically:
  - Step 3 (`comms_reserve`): Add "NEVER edit unreserved files" inline
  - Step 4 (implement): Add "MUST only modify reserved files" inline
  - Step 5 (`forge_progress`): Change to "MUST report progress" with mandatory framing
  - Step 6 (`hivemind_store`): Change to "MUST store learnings" with guidance on what to store (gotchas, patterns, decisions)
- [x] 1.2 Add reservation failure recovery as a sub-instruction under the `comms_reserve` step. If reservation fails, expires, or is released: STOP and report to coordinator via `comms_send`.
- [x] 1.3 Remove the separate `## Constraints` section entirely. Verify all 4 constraints have been migrated:
  - [x] "Only edit files you have reserved" → step 3/4
  - [x] "Report progress at regular intervals" → step 5
  - [x] "Store learnings for future agents" → step 6
  - [x] "Never modify files outside your assignment" → step 3/4
- [x] 1.4 Verify the restructured file stays under 35 lines (design decision D4). [20 lines]

## 2. Verification

- [x] 2.1 Run `make build` to verify the embedded asset compiles without errors.
- [x] 2.2 Run `make test` to verify no existing tests break (especially parity tests and agentkit embedding tests).
- [x] 2.3 Verify constitution alignment: the change maintains Autonomous Collaboration (reservation enforcement is structurally durable) and Observable Quality (progress reporting is mandatory). Confirm no new external dependencies introduced (Composability, Testability).
- [x] 2.4 [P] Add a content-verification test in `internal/agentkit/agentkit_test.go` that reads the embedded `worker.md` and asserts:
  - No `## Constraints` heading exists
  - The step containing `comms_reserve` includes MUST or NEVER language about file editing
  - The step containing `forge_progress` includes "MUST" language
  - The step containing `hivemind_store` includes "MUST" language
  - A reservation failure recovery instruction exists (contains `comms_send` in context of failure/stop)
  - Total line count is <= 35
  Follow the existing `TestSkillTemplates_HaveNameField` pattern for reading from the embedded FS.

<!-- spec-review: passed -->
<!-- code-review: passed -->
