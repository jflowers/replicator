## Context

The org-infra CI adoption enabled project-wide Go linting and issue #119 recorded
46 pre-existing findings across 18 files. A fresh scan with golangci-lint 2.11.4
reports 54 findings: 50 `errcheck`, 3 `staticcheck`, and 1 `unused`. The vet,
formatting, critic, and ineffective-assignment counts in issue #119 belong to an
older snapshot and remain clean-verification targets rather than current inventory
rows. Most current findings are ignored errors from resource cleanup, SQL
operations, transaction rollback, scans, and subprocesses. Issue #119 defines the
repository-wide verification target.

This design follows the proposal's constitution assessment: preserve independent
and self-describing MCP behavior, retain standalone operation and graceful
integration degradation, improve observable failures, and verify changes with
isolated tests.

## Human-Approved Interpretations

On 2026-09-30, the maintainer approved these interpretations to resolve review
circularities:

1. Behavioral production changes require a failing Go test. Strictly mechanical
   production edits and test-only lint corrections may use the reproducing lint
   diagnostic as the failing automated check while package tests remain green.
2. The implementation uses one atomic completion checkbox because the full lint
   gate remains red until all findings are fixed; the 54-row inventory is the
   granular progress and resumability record.
3. Local CI-equivalent checks gate task completion. Hosted CI is a post-PR,
   pre-merge gate and cannot be task-completion evidence.

OpenSpec is expressly authorized by the repository's two-tier specification
framework and stores tactical artifacts under `openspec/changes/`. The older
phase-boundary wording that names only `specs/NNN-*/` is treated as applying to
the equivalent OpenSpec artifact directory; this change does not amend that
governance text. No production or test code may change until review passes and
these OpenSpec artifacts are committed and pushed.

## Goals / Non-Goals

### Goals

- Make all default-build-tag Go packages selected by `golangci-lint run ./...`
  pass the currently enabled lint checks; verify parity-tagged code separately.
- Preserve transaction correctness and the most useful error when operation and
  cleanup failures occur together.
- Make previously discarded actionable failures observable through existing
  error-return and structured response paths.
- Use focused regression tests for behavioral changes and run every existing
  quality gate unchanged.
- Keep corrections local to the affected package and function unless an existing
  project abstraction already fits.

### Non-Goals

- Changing linter configuration, exclusions, severity, CI workflow gates,
  coverage ratchets, race flags, or review settings.
- Adding dependencies, runtime public APIs, MCP tools, CLI capabilities, database
  schema, or response fields. The test-only parity report helper may return an
  error to make writer failures observable.
- Broad refactoring, package moves, or opportunistic cleanup unrelated to the
  reported findings.
- Splitting issue #119 into separate delivery tracks.

## Decisions

### 1. Establish and preserve a finding inventory

Maintain `lint-findings.md` in this change directory as the durable inventory.
Record the command and tool version at the top. Give every finding a stable ID
and record its file and line, linter, behavioral or mechanical classification,
planned correction, required test or explicit justification, package owner, and
disposition (`open` or `resolved`). Every baseline finding must have one row and
every correction must trace to that row. This prevents scope drift and provides
an objective zero-finding completion condition.

### 2. Correct findings in place

Use the smallest local correction that satisfies the existing contract. Do not
introduce cross-package helpers or generic cleanup abstractions solely to remove
lint findings. Existing package boundaries and error conventions remain the
design boundary, supporting Composability First and the zero-waste mandate.

### 3. Preserve primary errors while handling cleanup failures

For resources and SQL operations, return or combine actionable cleanup errors
when the function's contract permits it. When an operation error already exists,
preserve it as the primary context and combine a distinct cleanup failure only
when callers can act on it. Expected, non-actionable cleanup outcomes, such as
`sql.ErrTxDone` after a successful commit or response-body close after full
consumption, may be explicitly discarded at the call site with a concise
justification. A rollback failure before commit must be combined with the primary
operation error. Scan,
transaction work, commit, subprocess, integration, and other actionable operation
errors must be returned through the existing contract. Errors must not be hidden
merely to satisfy `errcheck`.

Transaction changes use one coherent sequence: begin, immediately defer one
rollback attempt, perform work, and commit only after all required operations
succeed. The deferred rollback ignores only `sql.ErrTxDone` after a successful
commit. After work or commit failure, a rollback failure is combined with the
primary error. Tests cover success, work failure, commit failure, and both
work/rollback and commit/rollback dual failures.

### 4. Propagate operation and integration errors through existing contracts

Check scan, row iteration, subprocess, git, HTTP, and integration results at the
point where their failure becomes meaningful. Wrap returned errors with static
operation context and preserve sentinel matching with `%w`. Error paths must not
include credentials, authorization headers, URL user information or query values,
request or response bodies, untrusted subprocess output, or newly exposed absolute
paths. MCP JSON fields, error codes and categories, HTTP status mapping, CLI exit
codes, and parity fixtures remain unchanged. Internal error strings may gain
static operation context; public boundaries must continue mapping failures through
their existing structured error contracts. This protects Autonomous Collaboration
and Observable Quality without changing response shapes or exposing secrets.

### 5. Keep mechanical fixes semantics-neutral

Every production change that alters behavior or control flow starts with a
failing Go regression test, followed by the minimum implementation needed to make
it pass. Strictly mechanical production edits and test-only assertion or cleanup
corrections do not create a new behavior to express as a failing Go test; for
those edits, the reproducing lint diagnostic is the failing automated check and
the existing package tests must remain green before and after the edit. Apply
`gofmt` and direct source corrections for vet, staticcheck, gocritic, unused code,
and ineffective assignments, then rerun the specific linter. Do not mix these
corrections with unrelated renaming, control-flow redesign, or API changes.

### 6. Use TDD for observable behavioral corrections

Before changing error propagation or transaction control flow, add or strengthen
a focused failing test that demonstrates the required behavior. Database tests
use `db.OpenMemory()`, git and filesystem tests use `t.TempDir()`, and HTTP tests
use `httptest`. The standard library testing package remains the only assertion
framework.

Before source edits, add a test matrix to `lint-findings.md` for behavioral rows:

| Failure class | Fixture mechanism | Required invariants |
| --- | --- | --- |
| SQL scan/work | malformed fixture data, constraint failure, or closed test DB using `db.OpenMemory()` | error chain is preserved; no partial success is returned |
| Commit/rollback | deferred rollback runs exactly once after begin; only post-commit `sql.ErrTxDone` is ignored, while pre-commit rollback failures are combined with the work or commit error | commit occurs only after successful work; all failures preserve primary and rollback errors; failed work is not persisted |
| Row/resource close | package-local failing closer or writer fixture introduced before the implementation change; non-actionable post-consumption closes are explicitly justified | no leak on normal paths; actionable cleanup failure is returned without replacing a primary error |
| Subprocess/git/filesystem | failing executable, invalid repository operation, or permission/path fixture under `t.TempDir()` | existing exit/return contract remains; untrusted output and absolute paths are not exposed |
| HTTP/integration | `httptest.Server` and failing request/body fixtures | existing status/error category remains; credentials, headers, URLs, and bodies are sanitized |
| MCP/report writer | failing reader/writer fixture | JSON response shape and parity fixture remain unchanged; write failure is observable where the contract supports it |
| Transaction rollback | package-local finalizer seam with injected operation and rollback sentinels | `sql.ErrTxDone` after commit is ignored; pre-commit rollback failure is combined with the primary error |

### 7. Verify all protected gates without modification

Completion requires clean formatting and lint output plus the existing vet, race,
coverage, vulnerability, build, and build-tagged parity checks. Local lint evidence
uses `golangci-lint run ./...` with the tool version recorded in
`lint-findings.md`. Local verification uses Go 1.25.13 from `go.mod`,
golangci-lint 2.11.4, and govulncheck installed at commit
`3e6f44f962742443c11ae2261f02e0c917aeb2bc`. If `mega-linter-runner` is available, run it against the
repository configuration; otherwise record that it is unavailable and use the
direct Go linter locally. The hosted `Standardized CI / Run linters` check from
`complytime/org-infra` SHA `0c784711926c9864f027ec565fd7c06a382d80f8`
(v0.7.1) remains the authoritative post-PR MegaLinter evidence before merge. No
implementation task may alter a protected threshold or configuration to obtain a
passing result.

### 8. Track implementation as one atomic completion unit

The repository requires the CI-equivalent gate before any task checkbox is marked
complete. The lint gate is intentionally red until findings across all affected
packages are resolved, so package-level checkboxes could not truthfully be
completed as work progresses. `tasks.md` therefore has one atomic completion
checkbox, while `lint-findings.md` provides per-finding progress, ownership,
evidence, and disposition. The checkbox is marked only after all 54 rows are
resolved and every completion gate passes. This preserves both the
CI-before-completion rule and detailed resumability.

## Risks / Trade-offs

- Surfacing previously ignored errors can alter internal control flow. Focused
  failure-path tests and preservation of existing public contracts reduce this
  risk.
- Combining cleanup and operation errors can change exact error text. Tests should
  assert with `errors.Is` or stable contextual fragments rather than full strings.
- A single project-wide issue touches many packages, increasing review load. A
  strict finding inventory and package-local commits keep the work reviewable
  without creating excessive child issues.
- Some cleanup failures are not actionable. Explicitly justified discards are
  preferable to artificial propagation that obscures the primary result.
<!-- scaffolded by uf v0.17.0 -->
