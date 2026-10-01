## Why

Issue #119 records an initial 46 Go lint findings across 18 files after
MegaLinter was adopted; a fresh local scan currently reports 54 findings as the
codebase has continued to evolve. The current findings include ignored database,
transaction, resource cleanup, scan, and subprocess errors that can hide
correctness and reliability failures, plus static-analysis and unused-code
violations. The vet, formatting, critic, and ineffective-assignment counts in the
original issue are historical context and remain verification gates rather than
entries in the current baseline. The repository cannot satisfy its required project-wide lint gate
until this existing debt is resolved.

## What Changes

- Resolve the current `errcheck`, `staticcheck`, and `unused` findings without
  weakening any quality gate.
- Verify the historically reported `govet`, `gofmt`, `gocritic`, and
  `ineffassign` categories remain clean.
- Handle or explicitly justify cleanup errors while preserving safe SQL row and
  transaction lifecycle behavior.
- Preserve existing public behavior and package boundaries when propagating
  previously ignored subprocess and integration errors.
- Add focused regression coverage where an error-handling correction changes
  observable control flow.
- Verify the full repository with lint, vet, race tests, coverage ratchets,
  vulnerability scanning, formatting checks, and a production build.

## Capabilities

### New Capabilities

- `go-code-quality`: The default-build-tag Go packages selected by
  `golangci-lint run ./...` satisfy the enabled checks while retaining the current linter configuration, severity policy,
  coverage ratchets, race detection, and CI workflow gates.

### Modified Capabilities

None.

### Removed Capabilities

None.

## Impact

- Affects existing Go files in CLI commands, comms, database, forge, gitutil,
  memory, MCP server and client, org, query, parity helpers, and related tests.
- Adds a durable finding inventory under this change directory to trace every
  baseline finding to its correction, test, or explicit justification.
- Intentionally changes internal failure outcomes where actionable errors were
  previously discarded: affected functions return or combine those failures
  instead of reporting success. Runtime public APIs, MCP response contracts, CLI capabilities, dependencies,
  and data models remain unchanged; the test-only parity report helper MAY return
  a write error so report failures become testable.
- Does not modify `.golangci.yml`, `.mega-linter.yml`, CI workflow gates,
  coverage thresholds, severity definitions, or protected review settings.
- Requires the OpenSpec artifacts to be committed and pushed before
  implementation begins, per the repository's spec commit gate.

## Constitution Alignment

Assessed against the Replicator constitution at
`.specify/memory/constitution.md`, which extends the Unbound Force org
constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The change preserves independent MCP tool interfaces and self-describing JSON
responses. Correct error handling improves the reliability of artifacts passed
between autonomous agents without introducing runtime coupling.

### II. Composability First

**Assessment**: PASS

The work changes no required dependency or deployment topology. Replicator
remains independently buildable and usable, and optional integrations continue
to degrade through their existing structured error paths.

### III. Observable Quality

**Assessment**: PASS

The change restores reproducible lint and vet checks and preserves all existing
machine-verifiable quality gates. Previously hidden failures are handled or
reported rather than silently discarded.

### IV. Testability

**Assessment**: PASS

Behavioral corrections receive focused regression tests using the repository's
isolated database, filesystem, git, and HTTP test patterns. The full suite runs
with race detection and unchanged coverage ratchets without external services.
<!-- scaffolded by uf v0.17.0 -->
