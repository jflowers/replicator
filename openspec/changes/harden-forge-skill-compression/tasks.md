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

## 1. Restructure the Skill Document

- [x] 1.1 Rewrite `internal/agentkit/content/skills/forge-coordination/SKILL.md` with the hardened structure:
  - Replace "## File Reservation Rules" with "## Coordinator-Only Operations" and "## Worker-Only Operations" sections
  - Open each role section with MUST/MUST NOT constraints before protocol steps
  - Inline conflict resolution steps into the Worker Protocol after the "Reserve files" step (remove standalone "## Conflict Resolution" section)
  - Update `comms_reserve` call in Worker Protocol to include `exclusive=true` as the default: `comms_reserve(paths=[...], exclusive=true, reason="...")`
  - Preserve existing Coordinator Protocol and Worker Protocol step sequences
  - Move `comms_release_all()` into the Coordinator-Only Operations section with a MUST-level access control statement

## 2. Sync the Mirror Copy

- [x] 2.1 Copy the updated content from `internal/agentkit/content/skills/forge-coordination/SKILL.md` to `.opencode/skills/forge-coordination/SKILL.md` so both files are byte-identical

## 3. Verification

- [x] 3.1 Run `make build` to verify the agentkit embed compiles with the updated skill file
- [x] 3.2 Run `make test` to verify no existing tests break
- [x] 3.3 Verify both skill files are byte-identical (diff returns no output)
- [x] 3.4 Verify constitution alignment: change is PASS/N/A on all four principles per proposal (no new code, no new tools, no behavioral changes — prompt hardening only)
<!-- spec-review: passed -->
<!-- code-review: passed -->
