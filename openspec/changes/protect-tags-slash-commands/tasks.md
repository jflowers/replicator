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

## 1. Add protect tags to command files

Each file is independent -- all five tasks touch different files and can
run in parallel.

- [x] 1.1 [P] Add `<protect>` tags to `internal/agentkit/content/commands/forge.md` — wrap Critical Invariants section, Workflow section (steps 1-8), and Rules section. Leave title, description, and `$ARGUMENTS` placeholder unprotected.
- [x] 1.2 [P] Add `<protect>` tags to `internal/agentkit/content/commands/handoff.md` — wrap the entire Workflow section including the ordering constraint, steps 0-5, dependency explanations, and handoff note structure. Leave title and description unprotected.
- [x] 1.3 [P] Add `<protect>` tags to `internal/agentkit/content/commands/inbox.md` — wrap the Workflow section (3 steps). Leave title, description, Filtering, and Sending sections unprotected.
- [x] 1.4 [P] Add `<protect>` tags to `internal/agentkit/content/commands/forge-status.md` — wrap the Workflow section (3 steps). Leave title, description, Interpreting Results, and Quick Health Check sections unprotected.
- [x] 1.5 [P] Add `<protect>` tags to `internal/agentkit/content/commands/org.md` — wrap Common Actions section and Sessions section. Leave title, description, Usage, and Epics sections unprotected.

## 2. Enable DCP configuration

- [x] 2.1 Add `"protectTags": true` to `opencode.json` at the top level

## 3. Verification

- [x] 3.1 Verify each command file has correctly placed `<protect>` / `</protect>` tags with no nesting errors
- [x] 3.2 Verify `opencode.json` is valid JSON after modification
- [x] 3.3 Run `make build` to confirm embedded files compile without issues
- [x] 3.4 Verify constitution alignment: confirm no Go code changes, no test changes, no MCP protocol changes (Autonomous Collaboration PASS, Composability First PASS, Testability PASS)
<!-- spec-review: passed -->
<!-- code-review: passed -->
<!-- scaffolded by uf vdev -->
