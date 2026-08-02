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

## 1. Harden forge-global skill

All tasks modify the same file (`internal/agentkit/content/skills/forge-global/SKILL.md`) -- no parallel execution.

- [x] 1.1 Replace "When to Forge / Don't forge" parallel lists with a decision table using Signal / Forge / Skip columns. Table MUST cover all 6 original criteria (3 forge, 3 skip). (design decision D1)
  **Files**: `internal/agentkit/content/skills/forge-global/SKILL.md`
- [x] 1.2 Inline `ttl_seconds` parameter into reservation step 1 (`comms_reserve(paths=[...], ttl_seconds=300)`) and remove standalone TTL bullet (design decision D2)
  **Files**: `internal/agentkit/content/skills/forge-global/SKILL.md`
- [x] 1.3 Add explicit temporal ordering language ("FIRST", "THEN", "FINALLY") to File Reservation Protocol steps (design decision D3)
  **Files**: `internal/agentkit/content/skills/forge-global/SKILL.md`

## 2. Harden always-on-guidance skill

All tasks modify the same file (`internal/agentkit/content/skills/always-on-guidance/SKILL.md`) -- no parallel execution.

- [x] 2.1 Add `## Critical Safety` section after the title, move "Never force push to main" into it (design decision D5)
  **Files**: `internal/agentkit/content/skills/always-on-guidance/SKILL.md`
- [x] 2.2 Move "Check `hivemind_find` before solving problems from scratch" from last to first position in Tool Usage (design decision D4)
  **Files**: `internal/agentkit/content/skills/always-on-guidance/SKILL.md`
- [x] 2.3 Split Code Quality list (5 items) into sub-grouped lists of 2-3 items under descriptive sub-headers. Total rule count MUST equal original 5. (design decision D6)
  **Files**: `internal/agentkit/content/skills/always-on-guidance/SKILL.md`
- [x] 2.4 Split Error Handling list (4 items) into sub-grouped lists of 2-3 items under descriptive sub-headers. Total rule count MUST equal original 4. (design decision D6)
  **Files**: `internal/agentkit/content/skills/always-on-guidance/SKILL.md`
- [x] 2.5 Split Testing list (5 items) into sub-grouped lists of 2-3 items under descriptive sub-headers. Total rule count MUST equal original 5. (design decision D6)
  **Files**: `internal/agentkit/content/skills/always-on-guidance/SKILL.md`

## 3. Verification

- [x] 3.1 [P] Run `make build` to verify embedded skill files compile without errors
- [x] 3.2 [P] Run `make test` to verify no regressions (existing agentkit tests verify file existence, count, and frontmatter)
- [x] 3.3 Verify semantic preservation: diff before/after content of both skill files and confirm every rule present in the original appears in the restructured version with equivalent meaning. No rules added, removed, or weakened.
- [x] 3.4 Verify constitution alignment: confirm no new imports added, no Go source files modified outside `internal/agentkit/content/skills/`, and `go.sum` unchanged. Verify AGENTS.md and README.md do not require updates (skill names and descriptions unchanged).
<!-- spec-review: passed -->
<!-- code-review: passed -->
