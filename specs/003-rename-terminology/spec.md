# Feature Specification: Terminology Rename + Agent Kit

**Feature Branch**: `003-rename-terminology`  
**Created**: 2026-04-06  
**Status**: Ready  
**Input**: User description: "Rename hive→org, swarmmail→comms, swarm→forge. Ship agent kit (commands, skills, agents) via replicator init."  
**References**: [Issue #9](https://github.com/unbound-force/replicator/issues/9), [upstream swarm-tools commands](https://github.com/joelhooks/swarm-tools/tree/main/packages/opencode-swarm-plugin/claude-plugin)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - New Tool Names (Priority: P1)

An AI coding agent connects to the replicator MCP server and discovers tools named with the new terminology: `org_cells`, `org_create`, `comms_send`, `comms_inbox`, `forge_decompose`, `forge_worktree_create`, etc. All tools function identically to their predecessors -- only the names have changed. The agent's existing workflows continue working after updating tool name references.

**Why this priority**: Tool names are the public API surface. Every agent prompt, every tutorial, every config that references `hive_cells` or `swarm_decompose` must change. This is the foundation that all other stories depend on.

**Independent Test**: Call `tools/list` on the MCP server and verify all 45 non-deprecated tools use the new prefixes (`org_`, `comms_`, `forge_`). Call each tool and verify identical behavior to the old name.

**Acceptance Scenarios**:

1. **Given** the MCP server is running, **When** `tools/list` is called, **Then** no tool names start with `hive_`, `swarmmail_`, or `swarm_`. All use `org_`, `comms_`, or `forge_` prefixes. The 8 deprecated `hivemind_*` tools retain their names.
2. **Given** the old tool `hive_create` existed, **When** `org_create` is called with the same arguments, **Then** the response shape and behavior are identical.
3. **Given** the `replicator docs` command exists, **When** it is run, **Then** the generated tool reference uses the new category names (Org, Comms, Forge, Memory) and all tool names match.

---

### User Story 2 - Agent Kit Scaffolding (Priority: P1)

A developer runs `replicator init` in a project directory and receives a complete agent kit: 5 slash commands, 7 skills, and 3 agent definitions installed to `.opencode/`. The commands reference the new tool names and provide a lightweight orchestration workflow.

**Why this priority**: Without the agent kit, replicator provides raw MCP tools but no higher-level workflows. The commands and skills are what make the tools usable by AI agents in practice.

**Independent Test**: Run `replicator init` in an empty directory. Verify 15 files are created across `.opencode/command/`, `.opencode/skills/`, and `.opencode/agents/`. Open `/forge` command and verify it references `forge_*` tools.

**Acceptance Scenarios**:

1. **Given** a directory without `.opencode/`, **When** `replicator init` is run, **Then** `.uf/replicator/cells.json` is created AND 15 agent kit files are created under `.opencode/`.
2. **Given** `.opencode/command/forge.md` already exists, **When** `replicator init` is run without `--force`, **Then** the existing file is NOT overwritten and a message indicates it was skipped.
3. **Given** `replicator init --force` is run, **When** agent kit files already exist, **Then** all files are overwritten with the latest versions from the binary.
   > **Scenarios 2-3 superseded 2026-10-26** by `openspec/changes/init-smart-sync/` -- smart-sync overwrites drifted kit files by default; `--skip-existing` restores the old skip behavior.
4. **Given** the agent kit is installed, **When** an AI agent loads the project in OpenCode, **Then** `/forge`, `/org`, `/inbox`, `/forge:status`, and `/handoff` are available as slash commands.

---

### User Story 3 - Lightweight Forge Orchestrator (Priority: P2)

An AI coding agent invokes `/forge` with a task description. The command decomposes the task into subtasks using `forge_decompose`, creates an epic with `org_create_epic`, spawns workers, monitors progress via `comms_inbox`, reviews results, and completes the forge. The orchestrator is lightweight -- it delegates to MCP tools without the full Socratic planning flow.

**Why this priority**: The forge orchestrator is the flagship workflow that ties all tools together. It's P2 because the individual tools (P1) must work first.

**Independent Test**: Run `/forge "add a health check endpoint"` and verify it decomposes the task, creates an epic, and provides worker prompts.

**Acceptance Scenarios**:

1. **Given** an agent invokes `/forge` with a task, **When** the command executes, **Then** it calls `forge_decompose`, creates an epic via `org_create_epic`, and produces worker spawn prompts.
2. **Given** the forge is active, **When** the agent checks status, **Then** `/forge:status` shows active workers, pending cells, and recent messages.
3. **Given** a forge completes, **When** `/handoff` is invoked, **Then** reservations are released, state is synced, and a handoff note is generated.

---

### User Story 4 - Consistent Naming Across Documentation (Priority: P3)

All documentation (README, AGENTS.md, constitution, tool reference, spec artifacts) uses the new terminology consistently. The naming convention table reflects the updated metaphor. No stale references to the old names remain in active documentation.

**Why this priority**: Documentation consistency is important but doesn't block functionality. It depends on all other stories being complete.

**Independent Test**: Search all active documentation for `hive_`, `swarmmail_`, `swarm_` (as tool prefixes). Zero matches outside of historical attribution.

**Acceptance Scenarios**:

1. **Given** the rename is complete, **When** AGENTS.md is read, **Then** the naming convention table shows Org/Comms/Forge terminology and the project structure reflects renamed packages.
2. **Given** the rename is complete, **When** `replicator docs` is run, **Then** the output groups tools under Org, Comms, Forge, and Memory categories.
3. **Given** the rename is complete, **When** all `.go` source files are searched for old tool name strings, **Then** zero matches for `"hive_"`, `"swarmmail_"`, or `"swarm_"` as tool name prefixes (excluding `hivemind_` which is intentionally unchanged).

---

### Edge Cases

- What happens if a project already has `.opencode/command/forge.md` from a previous `replicator init`? The file is skipped unless `--force` is passed.
  > **Superseded 2026-10-26** by `openspec/changes/init-smart-sync/` -- smart-sync overwrites drifted kit files by default; `--skip-existing` restores the old skip behavior.
- What happens if an agent sends a request using an old tool name (`hive_cells`)? The MCP server returns "Unknown tool" error. There are no backward-compatible aliases.
- What happens to parity test fixtures that reference old tool names? The fixtures are updated to use the new names. Parity testing compares response shapes, not tool names.
- What happens to the `hivemind_*` deprecated tools? They keep their names. They are Dewey proxy stubs with deprecation messages and are not part of the org/comms/forge namespace.

## Requirements *(mandatory)*

### Functional Requirements

#### Terminology Rename
- **FR-001**: All 11 `hive_*` MCP tools MUST be renamed to `org_*` (e.g., `hive_cells` → `org_cells`, `hive_create` → `org_create`).
- **FR-002**: All 10 `swarmmail_*` MCP tools MUST be renamed to `comms_*` (e.g., `swarmmail_send` → `comms_send`).
- **FR-003**: All 24 `swarm_*` MCP tools MUST be renamed to `forge_*` (e.g., `swarm_decompose` → `forge_decompose`).
- **FR-004**: The 8 `hivemind_*` tools MUST retain their current names (deprecated Dewey proxy).
- **FR-005**: Tool behavior, argument schemas, and response shapes MUST remain identical -- only the `name` field changes.
- **FR-006**: The `replicator docs` command MUST group tools under Org, Comms, Forge, and Memory categories using the new prefixes.

#### Agent Kit
- **FR-007**: `replicator init` MUST create 5 command files in `.opencode/command/`: `forge.md`, `org.md`, `inbox.md`, `forge-status.md`, `handoff.md`.
- **FR-008**: `replicator init` MUST create 7 skill files in `.opencode/skills/`: `always-on-guidance/SKILL.md`, `forge-coordination/SKILL.md`, `replicator-cli/SKILL.md`, `testing-patterns/SKILL.md`, `system-design/SKILL.md`, `learning-systems/SKILL.md`, `forge-global/SKILL.md`.
- **FR-009**: `replicator init` MUST create 3 agent files in `.opencode/agents/`: `coordinator.md`, `worker.md`, `background-worker.md`.
- **FR-010**: Agent kit files MUST be embedded in the binary and written from embedded content on `init`.
- **FR-011**: `replicator init` MUST NOT overwrite existing agent kit files unless `--force` is passed.
  > **Superseded 2026-10-26** by `openspec/changes/init-smart-sync/` -- smart-sync overwrites drifted kit files by default; `--skip-existing` restores the old skip behavior.
- **FR-012**: All command files MUST reference the new tool names (`org_*`, `comms_*`, `forge_*`).

#### Naming Convention
- **FR-013**: The naming convention table in AGENTS.md MUST be updated: Work items=Org, Individual item=Cell, Agent coordination=Forge, Messaging=Comms.
- **FR-014**: All active documentation (README, AGENTS.md, constitution) MUST use the new terminology consistently.

### Key Entities

- **Org**: The work item management domain (formerly Hive). Contains cells (individual work items), epics, sessions, and sync operations.
- **Comms**: The inter-agent communication domain (formerly Swarm Mail). Contains messages, reservations, and agent registration.
- **Forge**: The orchestration domain (formerly Swarm). Contains task decomposition, worker spawning, worktrees, progress tracking, review, and insights.
- **Agent Kit**: The set of commands, skills, and agents that `replicator init` scaffolds into a project. These files teach AI agents how to use replicator's MCP tools effectively.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `tools/list` returns exactly 53 tools: 11 `org_*`, 10 `comms_*`, 24 `forge_*`, 8 `hivemind_*`. Zero tools with `hive_`, `swarmmail_`, or `swarm_` prefixes.
- **SC-002**: `replicator init` in an empty directory creates exactly 16 files: 1 cells.json + 5 commands + 7 skills + 3 agents.
- **SC-003**: All 53 tools pass their existing test suites with zero failures after the rename (behavior unchanged).
- **SC-004**: The parity test suite shows 100% shape match for all tools (response shapes unchanged despite name changes).
- **SC-005**: Zero active source files or documentation contain `"hive_"`, `"swarmmail_"`, or `"swarm_"` as tool name prefixes (verified by grep, excluding `hivemind_*`).
- **SC-006**: The `/forge` command file references at least `forge_decompose`, `org_create_epic`, `comms_inbox`, and `forge_status` tools.

## Assumptions

- The 8 `hivemind_*` tools keep their names. They are deprecated Dewey proxy stubs and will be removed in a future version, not renamed.
- No backward-compatible aliases are provided for old tool names. This is a clean break.
- The agent kit files are adapted from the upstream `joelhooks/swarm-tools` Claude Code plugin, rewritten for OpenCode conventions and the new terminology.
- The `ralph` command and `ralph-supervisor` skill are excluded (Codex-specific, not applicable to OpenCode).
- The `cli-builder`, `queue`, `skill-creator`, and `skill-generator` global skills are deferred to a future spec.
- Agent kit files are delivered to `.opencode/` in the project directory (not `~/.config/`).

## Dependencies

- **Spec 002 (charm-ux)**: Must be merged first. The `replicator docs` command and lipgloss styling are prerequisites.
- **Issue #9**: This spec fully addresses all requirements and comments in the issue. Issue #9 can be closed when this spec is merged.
