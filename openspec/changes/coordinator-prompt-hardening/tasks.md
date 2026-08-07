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

## 1. Restructure coordinator prompt

- [x] 1.1 Rewrite `internal/agentkit/content/agents/coordinator.md` with the hardened structure:
  - Preserve existing YAML front matter (`name: coordinator`, `description`, `mode: subagent`)
  - Add identity-first opening statement embedding key constraints (NEVER reserves files, NEVER edits code directly)
  - Add "Critical Constraints" section with uppercase MUST/NEVER keywords, positioned before workflow
  - Include explicit ordering: MUST call `forge_review` for every worker BEFORE `forge_complete` (codifying implicit ordering from forge coordination skill)
  - Convert unordered rules list to numbered "Protocol" section with explicit workflow ordering
  - Preserve "Available Tools" section at the end
  - Maintain behavioral parity: all 7 rules (6 original + 1 codified: comms init, no file reservation, no code editing, review completions, store learnings, check inbox, broadcast context) MUST remain with equivalent semantics — verified by the presence of these markers: `comms_init`, `reserve`, `edit code`, `forge_review`, `hivemind_store`, `comms_inbox`, `forge_broadcast`

## 2. Automated structural test

- [x] 2.1 [P] Write a structural test in `internal/agentkit/agentkit_test.go` that reads the embedded `coordinator.md` and verifies:
  - YAML front matter contains `name: coordinator` and `mode: subagent`
  - The first paragraph after front matter contains the uppercase keyword "NEVER" and the phrase "reserve" (identity-first prohibition)
  - A "Critical Constraints" section header appears before a "Protocol" section header
  - Both appear before the "Available Tools" section header
  - All 7 behavioral rule markers are present: `comms_init`, `reserve`, `edit code`, `forge_review`, `hivemind_store`, `comms_inbox`, `forge_broadcast`
  - Constraint lines use uppercase RFC 2119 keywords (MUST, NEVER, ALWAYS)
  - The first 50% of lines after front matter contain both the file reservation prohibition and the review-before-complete ordering

## 3. Build verification

- [x] 3.1 Run `make build` to verify the embedded asset compiles without errors
- [x] 3.2 Run `make test` to verify all tests pass (including the new structural test)

## 4. Documentation

- [x] 4.1 [P] Verify AGENTS.md does not need updates (coordinator prompt structure is an internal detail, not a convention documented in AGENTS.md)
- [x] 4.2 [P] Verify GoDoc comments on the agentkit embed package do not need updates

## 5. Constitution alignment

- [x] 5.1 Verify the change aligns with Constitution Principle I (Autonomous Collaboration): the restructured prompt reinforces artifact-based collaboration and separation of concerns between coordinator and workers
- [x] 5.2 Verify the change aligns with Constitution Principle IV (Testability): the structural test verifies all compression-resilience properties are maintained
<!-- scaffolded by uf vdev -->
<!-- spec-review: passed -->
<!-- code-review: passed -->
