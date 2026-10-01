# Tasks: Go Rewrite — Remaining Phases

**Input**: Design documents from `/specs/001-go-rewrite-phases/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, quickstart.md
**Branch**: `001-go-rewrite-phases`

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[US1–US6]**: Which user story this task belongs to
- Exact file paths included in every task description
- TDD: test tasks appear alongside or before implementation tasks
- **TDD ordering**: Within each `[P]` group, follow Red-Green-Refactor — write the failing test first, then the implementation. Test tasks (e.g., T006) and implementation tasks (e.g., T005) in the same `[P]` group are executed by the same worker in TDD order.

---

## Phase 1: Hive Completion + Swarm Mail

**User Stories**: US1 (Complete Hive Tool Suite), US2 (Agent Messaging)
**Requirements**: FR-001 through FR-007, FR-020, FR-022
**Gate**: `make check` passes — all 20 tools (4 existing + 7 hive + 9 swarmmail) have passing tests

### 1A — Database Migrations (shared infrastructure for Phase 1)

- [x] T001 [US1] Add `sessions` table migration in `internal/db/migrations.go` — columns: session_id TEXT PK, agent_name TEXT, project_path TEXT, started_at TEXT, ended_at TEXT, handoff_notes TEXT, active_cell_id TEXT. Register in `migrations` slice in `internal/db/db.go`.
- [x] T002 [US2] Add `messages` table migration in `internal/db/migrations.go` — columns: id INTEGER PK AUTOINCREMENT, from_agent TEXT, to_agents TEXT (JSON), subject TEXT, body TEXT, importance TEXT, thread_id TEXT, ack_required INTEGER, acknowledged INTEGER, created_at TEXT. Add index on from_agent, created_at.
- [x] T003 [US2] Add `reservations` table migration in `internal/db/migrations.go` — columns: id INTEGER PK AUTOINCREMENT, agent_name TEXT, project_path TEXT, path TEXT, exclusive INTEGER, reason TEXT, ttl_seconds INTEGER, created_at TEXT, expires_at TEXT. Add index on path, agent_name.
- [x] T004 Test migration idempotency: add test in `internal/db/db_test.go` verifying all new tables are created by `db.OpenMemory()` and that calling `migrate()` twice is safe.

### 1B — Remaining Hive Domain Logic (US1)

- [x] T005 [P] [US1] Implement `CreateEpic` in `internal/hive/epic.go` — accept `CreateEpicInput{EpicTitle, EpicDescription, Subtasks []SubtaskInput}`, use `sql.Tx` for atomic insert of epic + N subtask rows with `parent_id`. Add `CreateEpicInput` and `SubtaskInput` types.
- [x] T006 [P] [US1] Add tests in `internal/hive/epic_test.go` — `TestCreateEpic_Success` (epic + 3 subtasks, verify parent_id), `TestCreateEpic_AtomicRollback` (simulate failure, verify no partial rows), `TestCreateEpic_EmptySubtasks` (epic with zero subtasks succeeds).
- [x] T007 [P] [US1] Implement `StartCell` in `internal/hive/cells.go` — set `status='in_progress'`, update `updated_at`. Return error if cell not found.
- [x] T008 [P] [US1] Implement `ReadyCell` in `internal/hive/cells.go` — query cells where `status='open'` and no parent cell has `status` in ('open', 'in_progress', 'blocked'). Order by `priority DESC`. Return first match (single cell, not a list).
- [x] T009 [P] [US1] Add tests in `internal/hive/cells_test.go` — `TestStartCell_Success`, `TestStartCell_NotFound`, `TestReadyCell_UnblockedOnly` (create parent+child, verify only unblocked returned), `TestReadyCell_PriorityOrder`, `TestReadyCell_NoReady` (all blocked, returns nil).
- [x] T010 [US1] Implement `Sync` in `internal/hive/sync.go` — serialize all cells to JSON, write to `.uf/replicator/cells.json` in project dir, shell out to `git add .uf/replicator/ && git commit -m "hive sync"` via `os/exec`. Accept `projectPath string` parameter.
- [x] T011 [US1] Add tests in `internal/hive/sync_test.go` — `TestSync_WritesJSON` (use `t.TempDir()`, verify file contents), `TestSync_CommitsToGit` (init git repo in tempdir, run sync, verify commit exists via `git log`).
- [x] T012 [US1] Implement `SessionStart` and `SessionEnd` in `internal/hive/session.go` — `SessionStart(store, agentName, activeCellID)` creates session row, returns previous session's handoff_notes. `SessionEnd(store, handoffNotes)` updates current session with ended_at and handoff_notes.
- [x] T013 [US1] Add tests in `internal/hive/session_test.go` — `TestSessionStart_ReturnsHandoff` (create session with notes, start new, verify notes returned), `TestSessionStart_NoPrevious` (first session returns empty), `TestSessionEnd_SavesNotes`.

### 1C — Hive Tool Registration (US1)

- [x] T014 [US1] Register `hive_create_epic` tool in `internal/tools/hive/tools.go` — unmarshal `CreateEpicInput`, call `hive.CreateEpic()`, marshal result. Schema: `epic_title` (required string), `epic_description` (string), `subtasks` (array of objects with `title`, `priority`, `files`).
- [x] T015 [US1] Register `hive_query` tool in `internal/tools/hive/tools.go` — alias for existing `hive_cells` that matches the TypeScript tool name. Same schema and handler as `hiveCells()`.
- [x] T016 [US1] Register `hive_start` tool in `internal/tools/hive/tools.go` — unmarshal `{id: string}`, call `hive.StartCell()`. Schema: `id` (required string).
- [x] T017 [US1] Register `hive_ready` tool in `internal/tools/hive/tools.go` — no required args, call `hive.ReadyCell()`, marshal result. Return empty result if no ready cell.
- [x] T018 [US1] Register `hive_sync` tool in `internal/tools/hive/tools.go` — unmarshal `{auto_pull: boolean}`, call `hive.Sync()`. Schema: `auto_pull` (optional boolean).
- [x] T019 [US1] Register `hive_session_start` tool in `internal/tools/hive/tools.go` — unmarshal `{active_cell_id: string}`, call `hive.SessionStart()`. Schema: `active_cell_id` (optional string).
- [x] T020 [US1] Register `hive_session_end` tool in `internal/tools/hive/tools.go` — unmarshal `{handoff_notes: string}`, call `hive.SessionEnd()`. Schema: `handoff_notes` (required string).

### 1D — Swarm Mail Domain Logic (US2)

- [x] T021 [P] [US2] Implement `Init` in `internal/swarmmail/agent.go` — upsert agent into `agents` table (update `last_seen_at` if exists), accept `agentName`, `projectPath`, `taskDescription`. Return agent record. Add package-level types.
- [x] T022 [P] [US2] Add tests in `internal/swarmmail/agent_test.go` — `TestInit_NewAgent`, `TestInit_ExistingAgent` (verify upsert updates `last_seen_at`), `TestInit_DefaultRole`.
- [x] T023 [P] [US2] Implement `Send`, `Inbox`, `ReadMessage`, `Ack` in `internal/swarmmail/message.go` — `Send(store, msg)` inserts message row. `Inbox(store, agentName, limit, urgentOnly)` returns messages without body (max 5). `ReadMessage(store, messageID)` returns full message. `Ack(store, messageID)` sets `acknowledged=1`.
- [x] T024 [P] [US2] Add tests in `internal/swarmmail/message_test.go` — `TestSend_Persists`, `TestInbox_ExcludesBodies`, `TestInbox_MaxFive`, `TestInbox_UrgentOnly`, `TestReadMessage_FullBody`, `TestAck_MarksAcknowledged`, `TestReadMessage_NotFound`.
- [x] T025 [P] [US2] Implement `Reserve`, `Release`, `ReleaseAll`, `ReleaseAgent` in `internal/swarmmail/reservation.go` — `Reserve(store, agentName, paths, exclusive, reason, ttlSeconds)` checks for existing exclusive reservations (excluding expired) before inserting. `Release(store, paths, reservationIDs)` deletes by path or ID. `ReleaseAll(store, projectPath)` deletes all for project. `ReleaseAgent(store, agentName)` deletes all for agent.
- [x] T026 [P] [US2] Add tests in `internal/swarmmail/reservation_test.go` — `TestReserve_ExclusiveConflict` (agent A reserves, agent B fails), `TestReserve_TTLExpiry` (expired reservation allows re-reserve), `TestReserve_NonExclusive` (multiple non-exclusive succeed), `TestRelease_ByPath`, `TestRelease_ByID`, `TestReleaseAll`, `TestReleaseAgent`.

### 1E — Swarm Mail Tool Registration (US2)

- [x] T027 [US2] Create `internal/tools/swarmmail/tools.go` — implement `Register(reg, store)` function. Register all 9 swarmmail tools: `swarmmail_init`, `swarmmail_send`, `swarmmail_inbox`, `swarmmail_read_message`, `swarmmail_reserve`, `swarmmail_release`, `swarmmail_release_all`, `swarmmail_release_agent`, `swarmmail_ack`. Each tool: unmarshal JSON args, call domain function, marshal result. Match TypeScript argument schemas.
- [x] T028 [US2] Add `swarmmail_health` tool in `internal/tools/swarmmail/tools.go` — return database health status (ping SQLite, count agents/messages/reservations).

### 1F — Wire Up Phase 1 + Checkpoint

- [x] T029 Wire swarmmail tools into MCP server in `cmd/replicator/serve.go` — import `internal/tools/swarmmail`, call `swarmmail.Register(reg, store)` after `hive.Register()`.
- [x] T030 Phase 1 checkpoint: run `make check` (go vet + go test ./... -count=1 -race). All 20 tools registered, all tests pass. Verify tool count via `registry.Count() == 20`.

---

## Phase 2: Swarm Orchestration

**User Story**: US3 (Swarm Orchestration)
**Requirements**: FR-008 through FR-011
**Gate**: `make check` passes — all 36 tools (20 + 16 orchestration) have passing tests

### 2A — Git Utilities (shared infrastructure for Phase 2)

- [x] T031 [P] [US3] Create `internal/gitutil/git.go` — implement `Run(dir string, args ...string) (string, error)` using `os/exec.Command("git", args...)`. Set `cmd.Dir`, capture stdout/stderr, return stdout trimmed. On error, include stderr in wrapped error message.
- [x] T032 [P] [US3] Add `WorktreeAdd(projectPath, worktreePath, branch, startCommit)`, `WorktreeRemove(worktreePath)`, `WorktreeList(projectPath) ([]WorktreeInfo, error)` to `internal/gitutil/git.go`. Parse `git worktree list --porcelain` output into `WorktreeInfo{Path, Branch, Commit}` struct.
- [x] T033 [P] [US3] Add `CherryPick(projectPath, worktreeBranch, startCommit)`, `CurrentCommit(projectPath)` to `internal/gitutil/git.go`. `CherryPick` identifies commits on worktree branch since `startCommit` and cherry-picks each onto current branch. On conflict, return error listing conflicting files.
- [x] T034 [P] [US3] Add tests in `internal/gitutil/git_test.go` — guard with `if testing.Short() { t.Skip() }`. Tests: `TestRun_Success`, `TestRun_Failure` (bad command returns stderr), `TestWorktreeAdd_CreatesDirectory`, `TestWorktreeRemove`, `TestWorktreeList_ParsesPorcelain`, `TestCherryPick_AppliesCommits`, `TestCherryPick_DetectsConflict`, `TestCurrentCommit`. All use `t.TempDir()` with `git init`.

### 2B — Swarm Decomposition (US3)

- [x] T035 [P] [US3] Implement `Init` in `internal/swarm/init.go` — initialize swarm session, record event in `events` table with `type='swarm_init'`. Accept `projectPath`, `isolation` mode. Return session info.
- [x] T036 [P] [US3] Add tests in `internal/swarm/init_test.go` — `TestInit_RecordsEvent`, `TestInit_ReturnsSession`.
- [x] T037 [P] [US3] Implement `SelectStrategy`, `PlanPrompt`, `Decompose`, `ValidateDecomposition` in `internal/swarm/decompose.go` — these are prompt generators (no LLM calls). `SelectStrategy(task, context)` returns recommended strategy. `PlanPrompt(task, strategy, context, maxSubtasks)` returns structured prompt text. `Decompose(task, context, maxSubtasks)` returns decomposition prompt. `ValidateDecomposition(response)` validates JSON response against CellTree schema.
- [x] T038 [P] [US3] Add tests in `internal/swarm/decompose_test.go` — `TestSelectStrategy_ReturnsStrategy`, `TestPlanPrompt_ContainsSections`, `TestDecompose_IncludesFileList`, `TestValidateDecomposition_ValidJSON`, `TestValidateDecomposition_InvalidJSON`.

### 2C — Swarm Spawn + Subtask Management (US3)

- [x] T039 [P] [US3] Implement `SubtaskPrompt`, `SpawnSubtask`, `CompleteSubtask` in `internal/swarm/spawn.go` — `SubtaskPrompt(agentName, beadID, epicID, title, files, sharedContext)` generates prompt text for spawned agent. `SpawnSubtask(beadID, epicID, title, files, description, sharedContext)` prepares subtask metadata. `CompleteSubtask(beadID, taskResult, filesTouched)` handles subtask completion.
- [x] T040 [P] [US3] Add tests in `internal/swarm/spawn_test.go` — `TestSubtaskPrompt_IncludesContext`, `TestSubtaskPrompt_IncludesFiles`, `TestSpawnSubtask_ReturnsMetadata`, `TestCompleteSubtask_RecordsResult`.

### 2D — Swarm Progress + Status (US3)

- [x] T041 [P] [US3] Implement `Progress`, `Complete`, `Status`, `RecordOutcome` in `internal/swarm/progress.go` — `Progress(projectKey, agentName, beadID, status, progressPercent, message, filesTouched)` records progress event. `Complete(projectKey, agentName, beadID, summary, filesTouched, evaluation)` marks subtask complete (runs verification if not skipped). `Status(epicID, projectKey)` aggregates subtask statuses. `RecordOutcome(beadID, durationMs, success, strategy, filesTouched, errorCount, retryCount)` records outcome for feedback scoring.
- [x] T042 [P] [US3] Add tests in `internal/swarm/progress_test.go` — `TestProgress_RecordsEvent`, `TestComplete_MarksComplete`, `TestStatus_AggregatesSubtasks`, `TestRecordOutcome_PersistsData`.

### 2E — Swarm Worktree Operations (US3)

- [x] T043 [US3] Implement `WorktreeCreate`, `WorktreeMerge`, `WorktreeCleanup`, `WorktreeList` in `internal/swarm/worktree.go` — delegate to `internal/gitutil/`. `WorktreeCreate(projectPath, taskID, startCommit)` creates worktree at `{projectPath}/.worktrees/{taskID}`. `WorktreeMerge(projectPath, taskID, startCommit)` cherry-picks commits back. `WorktreeCleanup(projectPath, taskID, cleanupAll)` removes worktree (idempotent). `WorktreeList(projectPath)` returns active worktrees.
- [x] T044 [US3] Add tests in `internal/swarm/worktree_test.go` — guard with `if testing.Short() { t.Skip() }`. Tests: `TestWorktreeCreate_CreatesIsolatedDir`, `TestWorktreeMerge_CherryPicksCommits`, `TestWorktreeMerge_ConflictPreservesWorktree` (per spec edge case), `TestWorktreeCleanup_Idempotent`, `TestWorktreeList_ReturnsActive`.

### 2F — Swarm Review + Insights (US3)

- [x] T045 [P] [US3] Implement `Review`, `ReviewFeedback`, `AdversarialReview`, `EvaluationPrompt`, `Broadcast` in `internal/swarm/review.go` — prompt generators for review workflows. `Review(projectKey, epicID, taskID, filesTouched)` generates review prompt with diff. `ReviewFeedback(projectKey, taskID, workerID, status, issues, summary)` tracks review attempts (max 3). `AdversarialReview(diff, testOutput)` generates adversarial review prompt. `EvaluationPrompt(beadID, title, filesTouched)` generates self-evaluation prompt. `Broadcast(projectPath, agentName, epicID, message, importance, filesAffected)` sends context update.
- [x] T046 [P] [US3] Add tests in `internal/swarm/review_test.go` — `TestReview_IncludesDiff`, `TestReviewFeedback_TracksAttempts`, `TestReviewFeedback_FailsAfterThreeRejections`, `TestAdversarialReview_GeneratesPrompt`, `TestEvaluationPrompt_IncludesFiles`, `TestBroadcast_RecordsEvent`.

### 2G — Swarm Insights (US3)

- [x] T047 [P] [US3] Implement `GetStrategyInsights`, `GetFileInsights`, `GetPatternInsights` in `internal/swarm/insights.go` — query `events` table for past outcomes. `GetStrategyInsights(task)` returns success rates by strategy. `GetFileInsights(files)` returns file-specific gotchas. `GetPatternInsights()` returns top 5 failure patterns.
- [x] T048 [P] [US3] Add tests in `internal/swarm/insights_test.go` — `TestGetStrategyInsights_ReturnsRates`, `TestGetFileInsights_ReturnsGotchas`, `TestGetPatternInsights_ReturnsTopFive`, `TestInsights_EmptyDatabase`.

### 2H — Swarm Tool Registration (US3)

- [x] T049 [US3] Create `internal/tools/swarm/tools.go` — implement `Register(reg, store)`. Register all swarm tools: `swarm_init`, `swarm_select_strategy`, `swarm_plan_prompt`, `swarm_decompose`, `swarm_validate_decomposition`, `swarm_subtask_prompt`, `swarm_spawn_subtask`, `swarm_complete_subtask`, `swarm_progress`, `swarm_complete`, `swarm_status`, `swarm_record_outcome`, `swarm_worktree_create`, `swarm_worktree_merge`, `swarm_worktree_cleanup`, `swarm_worktree_list`, `swarm_review`, `swarm_review_feedback`, `swarm_adversarial_review`, `swarm_evaluation_prompt`, `swarm_broadcast`, `swarm_get_strategy_insights`, `swarm_get_file_insights`, `swarm_get_pattern_insights`. Each tool: unmarshal JSON args, call domain function, marshal result. Match TypeScript argument schemas.

### 2I — Wire Up Phase 2 + Checkpoint

- [x] T050 Wire swarm tools into MCP server in `cmd/replicator/serve.go` — import `internal/tools/swarm`, call `swarm.Register(reg, store)`.
- [x] T051 Phase 2 checkpoint: run `make check`. All 44+ tools registered, all tests pass. Git worktree integration tests pass (skip in `-short` mode).

---

## Phase 3: Memory and Context

**User Story**: US4 (Memory and Context)
**Requirements**: FR-012 through FR-014
**Gate**: `make check` passes — all 52+ tools have passing tests, Dewey proxy handles available and unavailable states

### 3A — Dewey Proxy Client (US4)

- [x] T052 [P] [US4] Create `internal/memory/proxy.go` — implement `Client` struct with `url string` and `http *http.Client`. Constructor `NewClient(deweyURL string) *Client` with 10s timeout. Method `Call(method string, params any) (json.RawMessage, error)` sends JSON-RPC 2.0 POST. Method `Health() error` pings Dewey endpoint. On connection failure, return structured error with code `DEWEY_UNAVAILABLE`.
- [x] T053 [P] [US4] Add tests in `internal/memory/proxy_test.go` — use `httptest.NewServer` to mock Dewey. Tests: `TestCall_ForwardsRequest`, `TestCall_ReturnsResponse`, `TestCall_DeweyUnavailable` (no server, verify error code), `TestHealth_Success`, `TestHealth_Failure`.

### 3B — Memory Tools (US4)

- [x] T054 [P] [US4] Implement `Store` and `Find` in `internal/memory/proxy.go` — `Store(client, information, tags)` proxies to Dewey `dewey_dewey_store_learning`, adds deprecation warning to response. `Find(client, query, collection, limit)` proxies to Dewey `dewey_dewey_semantic_search`, adds deprecation warning.
- [x] T055 [P] [US4] Add tests in `internal/memory/proxy_test.go` — `TestStore_ProxiesToDewey` (verify correct JSON-RPC method/params forwarded), `TestStore_IncludesDeprecationWarning`, `TestFind_ProxiesToDewey`, `TestFind_IncludesDeprecationWarning`.
- [x] T056 [P] [US4] Implement deprecated tool stubs in `internal/memory/deprecated.go` — `DeprecatedResponse(toolName, replacementTool string) string` returns `{"deprecated": true, "message": "...", "replacement": "dewey_xxx"}`. Map: `hivemind_get` → `dewey_get_page`, `hivemind_remove` → `dewey_delete_page`, `hivemind_validate` → (no equivalent), `hivemind_stats` → `dewey_health`, `hivemind_index` → `dewey_reload`, `hivemind_sync` → `dewey_reload`.
- [x] T057 [P] [US4] Add tests in `internal/memory/deprecated_test.go` — `TestDeprecatedResponse_IncludesReplacement` (for each of 6 deprecated tools, verify correct replacement name), `TestDeprecatedResponse_JSONShape`.

### 3C — Memory Tool Registration (US4)

- [x] T058 [US4] Create `internal/tools/memory/tools.go` — implement `Register(reg, client)` accepting `*memory.Client`. Register 8 tools: `hivemind_store`, `hivemind_find` (proxy tools), `hivemind_get`, `hivemind_remove`, `hivemind_validate`, `hivemind_stats`, `hivemind_index`, `hivemind_sync` (deprecated stubs). Match TypeScript argument schemas.

### 3D — Wire Up Phase 3 + Checkpoint

- [x] T059 Wire memory tools into MCP server in `cmd/replicator/serve.go` — create `memory.NewClient(cfg.DeweyURL)`, import `internal/tools/memory`, call `memory.Register(reg, memClient)`.
- [x] T060 Add `DeweyURL` field to `internal/config/config.go` — default `http://localhost:3333/mcp/`, read from `DEWEY_MCP_URL` env var.
- [x] T061 Phase 3 checkpoint: run `make check`. All 52+ tools registered, all tests pass. Dewey proxy tests use httptest mocks.

---

## Phase 4: CLI Commands

**User Story**: US5 (CLI Operations)
**Requirements**: FR-015 through FR-017
**Gate**: `make check` passes — all CLI commands execute correctly, `replicator doctor` completes in <2s, binary starts in <50ms

### 4A — Doctor Command (US5)

- [x] T062 [P] [US5] Create `internal/doctor/checks.go` — define `CheckResult{Name, Status, Message, Duration}` and `Options{Writer io.Writer, Config *config.Config, GitChecker, DeweyChecker}` types. Implement `Run(opts Options) ([]CheckResult, error)`. Individual check functions: `checkGit()` (run `git --version`, verify exit 0), `checkDatabase()` (open and ping SQLite), `checkDewey()` (HTTP GET to health endpoint), `checkConfigDir()` (verify `~/.config/uf/replicator/` exists). Use interface injection for external checks.
- [x] T063 [P] [US5] Add tests in `internal/doctor/checks_test.go` — `TestDoctor_AllPass` (mock all checks passing, verify output contains ✓), `TestDoctor_DeweyDown` (mock Dewey failure, verify ✗ with message), `TestDoctor_GitMissing`, `TestDoctor_CompletesInTwoSeconds`. Use `bytes.Buffer` as writer.
- [x] T064 [US5] Create `cmd/replicator/doctor.go` — implement `doctorCmd() *cobra.Command` and `runDoctor()` that creates `doctor.Options` and calls `doctor.Run()`. Format output as table to stdout.

### 4B — Stats Command (US5)

- [x] T065 [P] [US5] Implement stats queries in `internal/stats/stats.go` — `Run(store *db.Store, w io.Writer) error`. Query `events` table: count by type, recent activity (last 24h), active agents, cell counts by status. Format as human-readable table.
- [x] T066 [P] [US5] Add tests in `internal/stats/stats_test.go` — `TestStats_EmptyDatabase`, `TestStats_WithEvents` (insert sample events, verify output contains counts), `TestStats_FormatTable`.
- [x] T067 [US5] Create `cmd/replicator/stats.go` — implement `statsCmd() *cobra.Command` and `runStats()`.

### 4C — Query Command (US5)

- [x] T068 [P] [US5] Implement preset queries in `internal/query/presets.go` — define preset SQL queries as Go constants: `AgentActivity24h`, `CellsByStatus`, `SwarmCompletionRate`, `RecentEvents`. Implement `Run(store *db.Store, presetName string, w io.Writer) error` that executes the named query and formats results.
- [x] T069 [P] [US5] Add tests in `internal/query/presets_test.go` — `TestQuery_AgentActivity`, `TestQuery_CellsByStatus`, `TestQuery_UnknownPreset` (returns error), `TestQuery_FormatOutput`.
- [x] T070 [US5] Create `cmd/replicator/query.go` — implement `queryCmd() *cobra.Command` with preset name as positional arg. List available presets with `--list` flag.

### 4D — Setup Command (US5)

- [x] T071 [US5] Create `cmd/replicator/setup.go` — implement `setupCmd() *cobra.Command` and `runSetup()`. Create config dir `~/.config/uf/replicator/` if not exists, initialize database, verify git is installed. Print status for each step.

### 4E — Wire Up Phase 4 + Checkpoint

- [x] T072 Register new commands in `cmd/replicator/main.go` — add `root.AddCommand(doctorCmd())`, `root.AddCommand(statsCmd())`, `root.AddCommand(queryCmd())`, `root.AddCommand(setupCmd())`.
- [x] T073 Phase 4 checkpoint: run `make check`. Verify `time ./bin/replicator version` completes in <50ms. Verify `./bin/replicator doctor` runs all checks. All CLI commands have tests.

---

## Phase 5: Parity Testing

**User Story**: US6 (Parity Verification)
**Requirements**: FR-018, FR-019
**Gate**: Parity report shows 100% shape match for all implemented tools (SC-003)

### 5A — Shape Comparison Engine (US6)

- [x] T074 [P] [US6] Create `test/parity/shape.go` — implement `ShapeMatch(expected, actual json.RawMessage) (bool, []Difference)`. `Difference{Path, ExpectedType, ActualType}`. Recursive JSON shape comparison: objects compare keys and recurse on values, arrays compare first element shape, primitives compare type (string/number/bool/null). Path uses JSONPath notation (e.g., `$.content[0].text`).
- [x] T075 [P] [US6] Add tests in `test/parity/shape_test.go` — `TestShapeMatch_IdenticalObjects`, `TestShapeMatch_TypeMismatch`, `TestShapeMatch_MissingField`, `TestShapeMatch_ExtraField`, `TestShapeMatch_NestedDifference`, `TestShapeMatch_ArrayElementShape`, `TestShapeMatch_EmptyObjects`.

### 5B — Fixture Capture + Parity Harness (US6)

- [x] T076 [US6] Create `test/parity/fixtures/` directory structure — one JSON file per tool family: `hive.json`, `swarmmail.json`, `swarm.json`, `memory.json`. Each file contains `{toolName: {request: {...}, typescript_response: {...}}}` captured from the TypeScript cyborg-swarm server.
- [x] T077 [US6] Create `test/parity/parity_test.go` — build tag `//go:build parity`. Iterate over fixture files, for each tool: start Go MCP server in-process, send fixture request, compare response shape against TypeScript fixture using `ShapeMatch()`. Report pass/fail per tool.
- [x] T078 [US6] Implement parity report generation in `test/parity/report.go` — `GenerateReport(results []ToolResult, w io.Writer) error`. Output table: tool name, status (✓/✗), differences (if any). Summary line: `X/Y tools match (Z%)`.

### 5C — Phase 5 Checkpoint

- [x] T079 Phase 5 checkpoint: run `go test -tags parity ./test/parity/... -count=1`. Parity report shows shape match for all implemented tools. Generate report to `test/parity/report.txt`.
- [x] T080 Final validation: run `make check` (all unit/integration tests), verify binary size < 20MB (`ls -lh bin/replicator`), verify tool count matches expected total.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: No external dependencies — extends existing patterns
- **Phase 2**: Depends on Phase 1 (swarm tools use hive cells and swarmmail)
- **Phase 3**: Depends on Phase 1 (memory tools wire into same server). Can run in parallel with Phase 2 if desired.
- **Phase 4**: Depends on Phase 1 (doctor checks database). Can run in parallel with Phase 2/3.
- **Phase 5**: Depends on all other phases (parity tests cover all tools)

### Within-Phase Parallelism

- **Phase 1**: Tasks T005–T006 (epic) ∥ T007–T009 (start/ready) ∥ T021–T026 (swarmmail domain) — different packages
- **Phase 2**: Tasks T031–T034 (gitutil) ∥ T035–T038 (decompose) ∥ T039–T040 (spawn) ∥ T041–T042 (progress) — different packages
- **Phase 3**: Tasks T052–T053 (proxy) ∥ T056–T057 (deprecated) — different files in same package
- **Phase 4**: Tasks T062–T063 (doctor) ∥ T065–T066 (stats) ∥ T068–T069 (query) — different packages

### Cross-Phase Dependencies

- T029 (wire swarmmail) depends on T027 (swarmmail tools registration)
- T050 (wire swarm) depends on T049 (swarm tools registration)
- T059 (wire memory) depends on T058 (memory tools registration) + T060 (config)
- T072 (wire CLI) depends on T064, T067, T070, T071 (CLI commands)
- T076–T079 (parity) depends on all tool implementations

### Parallel Opportunities (Swarm Workers)

```
Worker A: T005–T006 (hive/epic.go)
Worker B: T007–T009 (hive/cells.go additions)
Worker C: T021–T026 (swarmmail/ package)
Worker D: T010–T013 (hive/sync.go + session.go)
```

After domain logic completes, tool registration (T014–T020, T027–T028) and wiring (T029) are sequential.

<!-- spec-review: passed -->
