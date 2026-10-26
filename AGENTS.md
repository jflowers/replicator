# Replicator Agent Guide

## Overview

Replicator is the Go rewrite of cyborg-swarm. It provides multi-agent
coordination tools via the MCP protocol and a CLI for observability.

## Core Mission

- **Strategic Architecture**: Engineers shift from manual
  coding to directing an "infinite supply of junior
  developers" (AI agents).
- **Outcome Orientation**: Focus on conveying business
  value and user intent rather than low-level technical
  sub-tasks.
- **Intent-to-Context**: Treat specs and rules as the
  medium through which human intent is manifested into
  code.

## Command Resolution

When a workflow names a slash command such as
`/uf.review-council`, resolve it only from its exact
project command file: `.opencode/commands/uf.review-council.md`.

- Do not substitute a similarly named global skill, agent,
  MCP tool, or external command.
- Load a skill only when the resolved command explicitly
  instructs use of the `skill` tool.
- Before executing a referenced command, identify its exact
  resolved path. If the path is absent, stop and report the
  missing command.
- `/uf.unleash` Steps 6 and 8 MUST use
  `.opencode/commands/uf.review-council.md`; they MUST NOT
  load or use the global `review-council` skill.

## Language & Toolchain

- Go 1.26.9+
- SQLite via `modernc.org/sqlite` (pure Go, no CGo)
- CLI via `cobra`
- Tests via `go test` (stdlib)

## Critical Rules

### TDD Everything

All code changes follow Red-Green-Refactor:
1. Write a failing test
2. Write minimal code to pass
3. Refactor while tests stay green

### Database

Single global database at `~/.config/uf/replicator/replicator.db`.
Use in-memory databases for tests (`db.OpenMemory()`).

### MCP Protocol

Tools are registered via the `registry` package and served
over stdio JSON-RPC. Each tool has:
- A name (e.g., `org_cells`)
- A description
- A JSON schema for arguments
- An execute function

### Naming Convention

| Concept | Name |
| --- | --- |
| Work items | **Org** |
| Individual item | **Cell** |
| Agent coordination | **Forge** |
| Messaging | **Comms** |
| Parallel workers | **Workers** |
| Task orchestrator | **Coordinator** |
| File locks | **Reservations** |

## Constitution (Highest Authority)

The project constitution at `.specify/memory/constitution.md`
extends the Unbound Force org constitution (v1.1.0) with four
core principles:

1. **I. Autonomous Collaboration**: Tools are callable
   independently via MCP. Outputs are self-describing JSON.
   Inter-agent communication uses comms.
2. **II. Composability First**: The binary works standalone.
   Dewey integration degrades gracefully. Database schema is
   compatible with cyborg-swarm.
3. **III. Observable Quality**: All tool responses are JSON.
   Response shapes match the TypeScript version (parity tests).
   Doctor reports are machine-readable.
4. **IV. Testability**: Database tests use in-memory SQLite.
   Git tests use `t.TempDir()`. HTTP tests use `httptest`.
   No external services required.

Constitution violations are CRITICAL severity and
non-negotiable.

## Behavioral Constraints

- **Zero-Waste Mandate**: No orphaned code, unused
  dependencies, or dead functions. Every file must serve a
  purpose traceable to a spec or tool.
- **CI Parity Gate**: Before marking any task complete, run
  the CI-equivalent checks locally. Read
  `.github/workflows/` for the exact commands -- do not
  rely on memory. Any failure blocks the task. The CI
  `Build and Test` check includes `govulncheck` for
  vulnerability scanning of Go dependencies. The
  `Standardized CI / Run linters` check runs MegaLinter
  (via `complytime/org-infra` reusable workflow) with
  linters configured in `.mega-linter.yml` and PR title
  validation. Note: MegaLinter runs in CI only; there is
  no local equivalent. Use `mega-linter-runner` if local
  linting is needed (requires Docker).
- **Scanner Environment Verification**: Before treating a
  local `govulncheck` result as a repository regression,
  compare `go version` with the version declared in
  `go.mod` and used by CI's `setup-go`. Search existing
  issues for the reported finding. When the local scanner
  uses a different Go standard library and the finding is
  not reproducible under the repository toolchain, report
  an environment discrepancy; do not create, propose, or
  block an unrelated change on a follow-up issue.
- **Intent Drift Detection**: Implementation must faithfully
  capture the spec's intent. The parity test suite verifies
  response shapes match the TypeScript version.
- **Automated Governance**: Constitution alignment is
  verified via the Constitution Check gate at planning
  time, not ad-hoc review.

### Gatekeeping Value Protection

Agents MUST NOT modify values that serve as quality or
governance gates to make an implementation pass. The
following categories are protected:

1. **Coverage thresholds and CRAP scores** — minimum
   coverage percentages, CRAP score limits, coverage
   ratchets
2. **Severity definitions and auto-fix policies** —
   CRITICAL/HIGH/MEDIUM/LOW boundaries, auto-fix
   eligibility rules
3. **Convention pack rule classifications** —
   MUST/SHOULD/MAY designations on convention pack rules
   (downgrading MUST to SHOULD is prohibited)
4. **CI flags and linter configuration** — `-race`,
   `-count=1`, `govulncheck`, `golangci-lint` rules,
   pinned action SHAs
5. **Agent temperature and tool-access settings** —
   frontmatter `temperature`, `tools.write`, `tools.edit`,
   `tools.bash` restrictions
6. **Constitution MUST rules** — any MUST rule in
   `.specify/memory/constitution.md` or hero constitutions
7. **Review iteration limits and worker concurrency** —
   max review iterations, max concurrent Swarm workers,
   retry limits
8. **Workflow gate markers** — `<!-- spec-review: passed
   -->`, task completion checkboxes used as gates, phase
   checkpoint requirements

**What to do instead**: When an implementation cannot
meet a gate, the agent MUST stop, report which gate is
blocking and why, and let the human decide whether to
adjust the gate or rework the implementation. Modifying
a gate without explicit human authorization is a
constitution violation (CRITICAL severity).

### Workflow Phase Boundaries

Agents MUST NOT cross workflow phase boundaries:

- **Specify/Clarify/Plan/Tasks/Analyze/Checklist** phases:
  spec artifacts ONLY (`specs/NNN-*/` directory). No
  source code, test, agent, command, or config changes.
- **Implement** phase: source code changes allowed,
  guided by spec artifacts.
- **Review** phase: findings and minor fixes only. No new
  features.

A phase boundary violation is treated as a process error.
The agent MUST stop and report the violation rather than
proceeding with out-of-phase changes.

### Review Council as PR Prerequisite

Before submitting a pull request, agents **must** run
`/uf.review-council` and resolve all REQUEST CHANGES
findings until all reviewers return APPROVE. There must
be **minimal to no code changes** between the council's
APPROVE verdict and the PR submission — the council
reviews the final code, not a draft that changes
afterward.

Workflow:

1. Complete all implementation tasks
2. Run CI checks locally (build, test, vet)
3. Run `/uf.review-council` — fix any findings, re-run
   until APPROVE
4. Commit, push, and submit PR immediately after council
   APPROVE
5. Do NOT make further code changes between APPROVE and
   PR submission

Exempt from council review:

- Constitution amendments (governance documents, not code)
- Documentation-only changes (README, AGENTS.md, spec
  artifacts)
- Emergency hotfixes (must be retroactively reviewed)

## Coding Conventions

- **Formatting**: `gofmt` and `goimports` (enforced by
  golangci-lint).
- **Naming**: Standard Go conventions. PascalCase exported,
  camelCase unexported.
- **Comments**: GoDoc-style on all exported functions and
  types. Package-level doc comments on every package.
- **Error handling**: Return `error`. Wrap with
  `fmt.Errorf("context: %w", err)`. Use `errors.Is` for
  sentinel errors (not string comparison).
- **Import grouping**: Standard library, then third-party,
  then internal packages (separated by blank lines).
- **No global state**: Prefer dependency injection and
  functional style.
- **JSON tags**: Required on all struct fields intended for
  serialization.
- **Constants**: Use string-typed constants for enumerations.

## Testing Conventions

- **Framework**: Standard library `testing` package only.
  No testify, gomega, or external assertion libraries.
- **Assertions**: Use `t.Errorf` / `t.Fatalf` directly.
- **Test naming**: `TestXxx_Description` (e.g.,
  `TestCreateCell_Defaults`, `TestReadyCell_PriorityOrder`).
- **Test isolation**: `db.OpenMemory()` for database tests.
  `t.TempDir()` for filesystem/git tests.
  `httptest.NewServer` for HTTP tests.
- **No shared state**: Each test creates its own store and
  fixtures. No test depends on another test's side effects.
- **Parity tests**: Build tag `//go:build parity`. Compare
  Go response shapes against TypeScript fixtures.
- **Git tests**: Guard with `if testing.Short() { t.Skip() }`
  for tests that shell out to git.
- **Workflow tests**: Node.js is required to execute embedded
  workflow policy. Use `requireNode(t)` so short-mode and local
  runs without Node skip; CI fails when Node is unavailable.

## Knowledge Retrieval

Agents SHOULD prefer Dewey MCP tools over grep/glob/read
for cross-repo context, design decisions, and architectural
patterns.

### Tool Selection Matrix

| Query Intent             | Dewey Tool                        | When to Use                    |
| ------------------------ | --------------------------------- | ------------------------------ |
| Conceptual understanding | `dewey_semantic_search`           | "How does X work?"             |
| Keyword lookup           | `dewey_search`                    | Known terms, FR numbers        |
| Read specific page       | `dewey_get_page`                  | Known document path            |
| Relationship discovery   | `dewey_find_connections`          | "How are X and Y related?"     |
| Similar documents        | `dewey_similar`                   | "Find specs like this one"     |
| Filtered semantic        | `dewey_semantic_search_filtered`  | Search within source type      |
| Graph navigation         | `dewey_traverse`                  | Dependency chain walking       |

### Graceful Degradation (3-Tier)

**Tier 3 (Full Dewey)**: `dewey_semantic_search`,
`dewey_search`, `dewey_traverse`, and
`dewey_semantic_search_filtered` for comprehensive
cross-repo context.

**Tier 2 (Graph-only, no embedding model)**: `dewey_search`
and `dewey_traverse` for keyword and structural queries.

**Tier 1 (No Dewey)**: Direct file operations -- Read tool,
Grep tool, convention packs at
`.opencode/unbound/packs/`.

## Spec-First Development

All changes that modify production code, test code, agent
prompts, embedded assets, or CI configuration **must** be
preceded by a spec workflow. The constitution
(`.specify/memory/constitution.md`) is the highest-
authority document in this project -- all work must align
with it.

**What requires a spec** (no exceptions without explicit
user override):

- New features or capabilities
- Refactoring that changes function signatures, extracts
  helpers, or moves code between packages
- Test additions or assertion strengthening across
  multiple functions
- Agent prompt changes
- CI workflow modifications
- Data model changes (new struct fields, schema updates)

**What is exempt** (may be done directly):

- Constitution amendments (governed by the constitution's
  own Governance section)
- Typo corrections, comment-only changes, single-line
  formatting fixes
- Emergency hotfixes for critical production bugs (must
  be retroactively documented)

When an agent is unsure whether a change is trivial, it
**must** ask the user rather than proceeding without a
spec. The cost of an unnecessary spec is minutes; the
cost of an unplanned change is rework, drift, and broken
CI.

## Specification Framework

This project uses a two-tier specification framework:

| Tier      | Tool     | When to Use              | Location              |
| --------- | -------- | ------------------------ | --------------------- |
| Strategic | Speckit  | 3+ stories, architecture | `specs/NNN-*/`        |
| Tactical  | OpenSpec | <3 stories, bug fix      | `openspec/changes/`   |

### Speckit Pipeline

```text
constitution -> specify -> clarify -> plan -> tasks
  -> analyze -> checklist -> implement
```

### OpenSpec Workflow

```text
propose -> design -> specs -> tasks -> apply -> archive
```

### Ordering Constraints

1. Constitution must exist before specs.
2. Spec before plan. Plan before tasks.
3. Tasks before implementation.
4. All checklists must pass before implementation.

### Task Completion Bookkeeping

Mark `- [ ]` to `- [x]` immediately after each task. Do
not batch completions.

### Documentation Validation Gate

Before marking any task complete, check whether changes
require updates to:

- `README.md` -- commands, flags, architecture
- `AGENTS.md` -- conventions, packages, patterns
- GoDoc comments -- exported functions and types
- Spec artifacts under `specs/`

### Website Documentation Gate

When a change affects user-facing behavior, hero
capabilities, CLI commands, or workflows, a GitHub issue
**MUST** be created in the `unbound-force/website`
repository to track required documentation or website
updates. The issue must be created before the
implementing PR is merged.

```bash
gh issue create --repo unbound-force/website \
  --title "docs: <brief description of what changed>" \
  --body "<what changed, why it matters, which pages
          need updating>"
```

**Exempt changes** (no website issue needed):
- Internal refactoring with no user-facing behavior
  change
- Test-only changes
- CI/CD pipeline changes
- Spec artifacts (specs are internal planning documents)

**Examples requiring a website issue**:
- New CLI command or flag added
- Hero capabilities changed (new agent, removed feature)
- Installation steps changed (`uf setup` flow)
- New convention pack added
- Breaking changes to any user-facing workflow

### Spec Commit Gate

All spec artifacts MUST be committed and pushed before
implementation begins.

## Git & Workflow

- **Commit format**: Conventional Commits --
  `type: description` (feat, fix, docs, chore, refactor).
- **Branching**: Feature branches required. Speckit:
  `NNN-<name>`. OpenSpec: `opsx/<name>`.
- **Code review**: Required before merge.
- **Semantic versioning**: For releases.

## Commands

```bash
make build    # Build binary
make test     # Run all tests
make vet      # Go vet
make check    # Vet + test
make serve    # Build and run MCP server
make release  # GoReleaser dry-run
make coverage       # Run tests with race detection and print coverage report
make check-coverage # Enforce coverage ratchets locally (mirrors CI step)
make install        # Install to GOPATH/bin
```

### CLI Commands

| Command | Purpose |
| --- | --- |
| `replicator init` | Per-repo setup: creates `.uf/replicator/` + scaffolds agent kit (smart-sync: updates drifted files; `--skip-existing` to preserve) |
| `replicator setup` | Per-machine setup: creates `~/.config/uf/replicator/` + SQLite DB |
| `replicator serve` | Start MCP JSON-RPC server on stdio |
| `replicator cells` | List org cells (work items) |
| `replicator doctor` | Check environment health |
| `replicator stats` | Display activity summary |
| `replicator query` | Run preset SQL analytics queries |
| `replicator docs` | Generate MCP tool reference (markdown) |
| `replicator version` | Print version, commit, build date |

## Project Structure

```text
cmd/replicator/       CLI entrypoint (cobra)
internal/
  agentkit/           Embedded agent kit (commands, skills, agents)
  config/             Configuration
  db/                 SQLite + migrations (7 tables)
  org/                Cell domain logic (CRUD, epics, sessions, sync)
  comms/              Agent messaging + file reservations
  forge/              Orchestration (decompose, spawn, worktree, review, insights)
  mcpclient/          MCP Streamable HTTP client (shared transport)
  memory/             Dewey proxy + deprecated tool stubs
  gitutil/            Git worktree operations (os/exec)
  doctor/             Health check engine
  stats/              Database statistics
  query/              Preset SQL queries
  mcp/                MCP JSON-RPC server + structured logging
  ui/                 Centralized lipgloss styles + table helpers
  workflowtest/       Workflow policy regression tests
  tools/
    registry/         Tool registration framework
    org/              Org tool handlers (11 tools)
    comms/            Comms tool handlers (10 tools)
    forge/            Forge tool handlers (24 tools)
    memory/           Memory tool handlers (8 tools)
test/parity/          Shape comparison engine + fixtures
```

## Drift Detection (Mirrored Files)

The `replicator init` command scaffolds agent kit files into
`.opencode/` via `embed.FS` assets stored under
`internal/agentkit/content/`. The scaffolded copies and
embedded copies MUST be byte-identical. Changes to either
location MUST be synchronized to the other.

`replicator init` smart-syncs the agent kit by default and is
the mechanical repair path for drifted pairs: it creates
missing files, skips files byte-identical to the embedded
content, and updates drifted files to the embedded content
(embedded is canonical, compared via `bytes.Equal`).
`cells.json` and DCP scaffolding are unchanged; `--force` is
a deprecated no-op. Warning: smart-sync overwrites drifted
kit files by default -- repositories that intentionally
customize kit files must pass `--skip-existing`, which
preserves all existing kit files while still creating
missing ones. In this repo, the `.opencode/` copies are
synced by running `replicator init` in the repo root
(dogfood task 4.2 in `openspec/changes/init-smart-sync/`).

Run `make check` to confirm drift detection tests pass.

### Mirrored File Pairs

**Skills** (7 files):

| Scaffolded Path | Embedded Path |
|----------------|---------------|
| `.opencode/skills/always-on-guidance/SKILL.md` | `internal/agentkit/content/skills/always-on-guidance/SKILL.md` |
| `.opencode/skills/forge-coordination/SKILL.md` | `internal/agentkit/content/skills/forge-coordination/SKILL.md` |
| `.opencode/skills/forge-global/SKILL.md` | `internal/agentkit/content/skills/forge-global/SKILL.md` |
| `.opencode/skills/learning-systems/SKILL.md` | `internal/agentkit/content/skills/learning-systems/SKILL.md` |
| `.opencode/skills/replicator-cli/SKILL.md` | `internal/agentkit/content/skills/replicator-cli/SKILL.md` |
| `.opencode/skills/system-design/SKILL.md` | `internal/agentkit/content/skills/system-design/SKILL.md` |
| `.opencode/skills/testing-patterns/SKILL.md` | `internal/agentkit/content/skills/testing-patterns/SKILL.md` |

**Commands** (5 files):

| Scaffolded Path | Embedded Path |
|----------------|---------------|
| `.opencode/commands/forge.md` | `internal/agentkit/content/commands/forge.md` |
| `.opencode/commands/forge-status.md` | `internal/agentkit/content/commands/forge-status.md` |
| `.opencode/commands/handoff.md` | `internal/agentkit/content/commands/handoff.md` |
| `.opencode/commands/inbox.md` | `internal/agentkit/content/commands/inbox.md` |
| `.opencode/commands/org.md` | `internal/agentkit/content/commands/org.md` |

**Agents** (3 files):

| Scaffolded Path | Embedded Path |
|----------------|---------------|
| `.opencode/agents/coordinator.md` | `internal/agentkit/content/agents/coordinator.md` |
| `.opencode/agents/worker.md` | `internal/agentkit/content/agents/worker.md` |
| `.opencode/agents/background-worker.md` | `internal/agentkit/content/agents/background-worker.md` |

Not all `.opencode/` files have embedded copies. Files
that appear only in `.opencode/` (e.g., Divisor agents,
Speckit commands, OpenSpec skills) are scaffolded by the
Unbound Force CLI, not by `replicator init`, and are not
subject to drift detection.

### Test Coverage

`TestSkillFiles_DriftDetection` in
`internal/agentkit/agentkit_test.go` enforces byte
parity for `always-on-guidance` and `forge-global`
skills. `TestForgeCoordinationSkill_StructuralHardening`
checks `forge-coordination` inline. See issue #70 for
planned expansion of drift detection coverage to all 15
mirrored files. If these test names or their coverage
scope change, update this section to match.

### When Reviewing or Implementing

When a PR modifies a file in either column of the tables
above, verify the corresponding copy is updated. If only
one copy is changed, the drift detection tests will fail
in CI (`make check`) for the three skills currently
covered (see Test Coverage above). For the remaining
pairs, run `replicator init` (or the drift tests) instead
of manual byte comparison -- smart-sync updates drifted
copies to the embedded content; commit the synced files.
Issue #70 tracks expanding test coverage to all 15
mirrored pairs.

## Credits

Go rewrite of [cyborg-swarm](https://github.com/unbound-force/cyborg-swarm),
originally by [Joel Hooks](https://github.com/joelhooks).

## Active Technologies
- Go 1.26.9+ + `cobra` (CLI), `modernc.org/sqlite` (pure Go SQLite), stdlib `encoding/json` (MCP JSON-RPC), stdlib `os/exec` (git operations) (001-go-rewrite-phases)
- SQLite at `~/.config/uf/replicator/replicator.db` (WAL mode) (001-go-rewrite-phases)
- Go 1.26.9+ + `charmbracelet/lipgloss v1.1.0`, `charmbracelet/log v1.0.0`, `muesli/termenv v0.16.0`, `charmbracelet/lipgloss/table` (sub-package of lipgloss) (002-charm-ux)
- SQLite via `modernc.org/sqlite` (unchanged) (002-charm-ux)
- Go 1.26.9+ + cobra (CLI), modernc.org/sqlite (pure Go SQLite), embed (stdlib) (003-rename-terminology)

## Recent Changes
- 101-adopt-org-infra-ci: Added MegaLinter CI workflow via `complytime/org-infra` reusable workflow, added `.mega-linter.yml` config, added `Standardized CI / Run linters` to branch protection and release preflight gating
- 428-adopt-org-infra-release-workflows: Adopted org-infra reusable workflows for release pipeline, added govulncheck to CI, added per-package coverage ratchets, added CI convention pack, added SECURITY.md and CODEOWNERS, migrated commands to `uf.*` namespace
- 001-go-rewrite-phases: Added Go 1.25+ + `cobra` (CLI), `modernc.org/sqlite` (pure Go SQLite), stdlib `encoding/json` (MCP JSON-RPC), stdlib `os/exec` (git operations)

## Convention Packs

This repository uses convention packs scaffolded by
unbound-force. Agents MUST read the applicable pack(s)
before writing or reviewing code.

- `.opencode/uf/packs/default.md`
- `.opencode/uf/packs/default-custom.md`
- `.opencode/uf/packs/severity.md`
- `.opencode/uf/packs/content.md`
- `.opencode/uf/packs/content-custom.md`
- `.opencode/uf/packs/ci.md`
- `.opencode/uf/packs/ci-custom.md`
- `.opencode/uf/packs/go.md`
- `.opencode/uf/packs/go-custom.md`
