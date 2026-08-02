<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Harden handoff.md

All tasks in this group modify the same file (`internal/agentkit/content/commands/handoff.md`), so none are parallel-eligible.

- [x] 1.1 Add ordering constraint header to the workflow section: insert "Steps MUST execute in this exact order -- do not reorder or parallelize" before the numbered step list
- [x] 1.2 Add forge precondition check before the existing step 1 (renumber existing steps accordingly): "SHOULD check for active forge workers. If workers are active, warn and request confirmation before proceeding. If forge tools are unavailable or the agent has no active forge context, skip this check."
- [x] 1.3 Add step dependency rationale to each workflow step: brief inline explanation of why each step depends on the previous one completing first (e.g., "summarize first so you know what to report in handoff notes")
- [x] 1.4 Inline the handoff note template into the `org_session_end` step: move the 5-category template (Completed, In Progress, Blocked, Next Steps, Gotchas) from the separate "Handoff Note Template" section into the step 5 instruction, formatted as a code block showing the expected `handoff_notes` argument structure
- [x] 1.5 Remove the separate "## Handoff Note Template" section (currently the last section in the file)

## 2. Verify

- [x] 2.1 Run `make build` to verify the embedded asset compiles cleanly
- [x] 2.2 Run `make test` to verify agentkit embedding tests pass (file presence and prefix matching)
- [x] 2.3 Add a structural content test in `internal/agentkit/agentkit_test.go` that reads the embedded `handoff.md` and asserts: (1) ordering constraint text appears before the first numbered step, (2) the handoff note categories (Completed, In Progress, Blocked, Next Steps, Gotchas) appear within the `org_session_end` step section, (3) the separate "## Handoff Note Template" section is absent, (4) the forge precondition check text is present
- [x] 2.4 Verify constitution alignment: confirm the forge check uses SHOULD (not MUST), no new runtime dependencies are introduced, and the handoff note template is co-located with the tool call instruction
<!-- spec-review: passed -->
<!-- code-review: passed -->
