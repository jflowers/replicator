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

  NOTE: All tasks in group 1 modify the same file
  (internal/agentkit/content/commands/forge.md) so
  none are parallel-eligible.
-->

## 1. Restructure forge.md for DCP compression resilience

- [x] 1.1 Add Critical Invariants section after title, before Workflow. Include: (a) review MUST complete before marking work done, (b) NEVER pass `skip_review: true` to `forge_complete`, (c) always create a forge even for small tasks, (d) coordinator orchestrates, workers execute
- [x] 1.2 Rewrite step 7 text to embed explicit ordering constraint: "Review FIRST (before complete): `forge_review(task_id, files_touched)` for each completed worker — MUST finish before step 8"
- [x] 1.3 Reorder Rules section: move "Review every worker's output before marking complete" to 1st bullet position
- [x] 1.4 Inline strategy selection (`forge_get_strategy_insights`) as sub-item of step 3 (Decompose); remove standalone Strategy Selection section
- [x] 1.5 Inline error recovery guidance as sub-items of step 6 (Monitor); remove standalone Error Recovery section
- [x] 1.6 Merge Completion section content into step 8 as sub-items (`forge_complete`, `forge_record_outcome`, `hivemind_store`, `org_sync`); remove standalone Completion section

## 2. Verification

- [x] 2.1 Verify `make build` succeeds (forge.md is embedded content)
- [x] 2.2 Verify `make test` passes (no behavioral changes, parity tests unaffected)
- [x] 2.3 Verify all MCP tool call references from the original forge.md are present in the restructured version (grep for: `forge_review`, `forge_complete`, `forge_record_outcome`, `hivemind_store`, `org_sync`, `comms_inbox`, `forge_status`, `org_cells`, `comms_read_message`, `comms_ack`, `forge_get_strategy_insights`, `forge_decompose`, `org_create_epic`, `forge_spawn_subtask`, `comms_init`, `hivemind_find`, `comms_reserve`)
- [x] 2.4 Verify review-before-complete constraint appears in at least 3 locations (Critical Invariants section, step 7 text, first rule in Rules section) for compression redundancy
<!-- spec-review: passed -->
<!-- code-review: passed -->
