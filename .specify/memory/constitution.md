<!--
  SYNC IMPACT REPORT
  ==================
  Version change: 1.0.0 → 1.1.0 (MINOR: Spec-First Development
    replaced by Tier-Routed Workflow)
  Amendment date: 2026-11-25

  Modified sections:
    - Development Workflow: "Spec-First Development" (every
      non-trivial change needs Speckit or OpenSpec) replaced by
      "Tier-Routed Workflow" — Speckit, OpenSpec, or Quick chosen
      per change, matching /uf.unleash tier routing. Risk floors
      keep security, auth, data-loss, and API-breaking changes out
      of Quick. Branch naming gains `quick/<name>`.

  Unchanged principles: all

  Previous version history:
    - 1.0.0 (2026-04-05): Initial ratification
-->

# Replicator Constitution

**parent_constitution**: unbound-force/unbound-force v1.1.0

## Core Principles

### I. Autonomous Collaboration

Heroes MUST collaborate through well-defined artifacts --
files, reports, and schemas -- rather than runtime coupling
or synchronous interaction.

- Every tool MUST be callable independently via the MCP
  protocol without requiring other tools to be present.
- Tool outputs MUST be self-describing JSON with enough
  context for any consumer to interpret without consulting
  the producing tool.
- Inter-agent communication MUST use the comms
  messaging system, not ad-hoc coordination.

**Rationale**: A swarm of autonomous agents cannot rely on
real-time negotiation. MCP tools are stateless request/
response handlers; agents coordinate through messages and
shared database state, not direct coupling.

### II. Composability First

Replicator MUST be independently installable and usable
without any other hero being present. Optional integrations
MUST degrade gracefully.

- The binary MUST deliver core value (org, comms,
  forge) when deployed alone with no external
  services running.
- Dewey integration MUST degrade gracefully: when Dewey
  is unavailable, memory tools return structured
  `DEWEY_UNAVAILABLE` errors instead of crashing.
- The database schema MUST be compatible with cyborg-swarm
  so both can operate on the same SQLite file during
  migration.

**Rationale**: Replicator replaces a TypeScript monorepo.
It must work standalone from day one while coexisting with
the system it replaces during the transition period.

### III. Observable Quality

Every tool MUST produce machine-parseable output. All
quality claims MUST be backed by automated, reproducible
tests.

- All MCP tool responses MUST return JSON content blocks
  per the MCP specification.
- Tool response shapes MUST match the TypeScript version's
  shapes (verified by the parity test suite).
- The `doctor` command MUST report environment health in
  a structured, machine-readable format.

**Rationale**: AI agents parse tool responses
programmatically. Inconsistent or unparseable output breaks
the agent workflow. Parity with the TypeScript version
ensures agents can switch backends transparently.

### IV. Testability

Every component MUST be testable in isolation without
requiring external services, network access, or shared
mutable state.

- Database tests MUST use `db.OpenMemory()` (in-memory
  SQLite) -- never touch the global database file.
- Git tests MUST use `t.TempDir()` with `git init` --
  never touch the user's repositories.
- HTTP tests MUST use `httptest.NewServer` -- never call
  live Dewey or Zen endpoints.
- Coverage strategy MUST be defined before implementation.
- Coverage regressions MUST be treated as test failures.

**Rationale**: A 15MB binary with 53 tools and 190+ tests
must be verifiable in seconds on any machine. External
service dependencies make tests flaky, slow, and
environment-dependent.

## Development Workflow

- **Tier-Routed Workflow**: Every change MUST follow a workflow
  tier chosen for that change's scope and risk; no single tier is
  mandatory. Speckit (`specs/NNN-<name>/`) for strategic work: 3+
  user stories, cross-cutting or cross-repo. OpenSpec
  (`openspec/changes/<name>/`) for tactical work: 1-2 stories,
  bounded blast radius. Quick (no spec artifacts) for trivial
  fixes, chores, and docs changes where ceremony exceeds value.
  `/uf.unleash` routes a change with no spec through the
  `tier-routing` rubric; manual work applies the same rubric.
  Risk floors escalate only: security-sensitive code,
  authentication or authorization flows, data-loss paths, or a
  public API break MUST use OpenSpec at minimum, never Quick.
  When signals conflict between adjacent tiers, choose the
  heavier one. Every tier, Quick included, keeps the full
  quality gates (CI, code review). Constitution amendments follow
  the Governance section.
- **Branching**: All work MUST occur on feature branches.
  Speckit: `NNN-<name>`. OpenSpec: `opsx/<name>`.
  Quick: `quick/<name>`.
- **Code Review**: Every PR MUST receive review. When the
  review council is available, agents MUST run
  `/uf.review-council` and receive APPROVE before PR
  submission.
- **CI Parity Gate**: Before marking any task complete,
  agents MUST run the CI-equivalent commands locally
  (`make check`). Derive commands from
  `.github/workflows/`, not from memory.
- **Continuous Integration**: CI MUST pass before merge.
- **Releases**: Semantic versioning. `workflow_dispatch`
  with tag input triggers preflight validation then
  GoReleaser.
- **Commit Messages**: Conventional commits
  (`type: description`).
- **Task Completion**: Mark `- [ ]` to `- [x]` immediately
  after each task, not in a batch.
- **Documentation Gate**: Before marking a task complete,
  check whether AGENTS.md, README.md, or GoDoc needs
  updating.

## Governance

This constitution extends the Unbound Force org constitution
(v1.1.0). On matters where this document and the org
constitution conflict, the org constitution prevails.

- **Amendments**: Proposed via PR, reviewed, approved.
- **Versioning**: semver (MAJOR/MINOR/PATCH).
- **Compliance Review**: Constitution Check gate at
  planning phase. Violations are CRITICAL severity.

**Version**: 1.1.0 | **Ratified**: 2026-04-05 | **Last Amended**: 2026-11-25
