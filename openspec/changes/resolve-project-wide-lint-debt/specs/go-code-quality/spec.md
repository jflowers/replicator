## ADDED Requirements

### Requirement: Project-Wide Go Lint Cleanliness

The repository's default-build-tag Go packages selected by `golangci-lint run
./...` from the module root MUST pass the enabled checks without suppressing findings through weaker configuration, reduced
severity, broader exclusions, or disabled linters. Every baseline finding MUST
be traced in the change's durable inventory to a correction, regression test, or
explicit justification. Resource cleanup, SQL lifecycle, scan, and subprocess
errors MUST be handled without compromising transaction safety, public behavior,
or sensitive-data boundaries.

#### Scenario: Full Go lint scan succeeds

- **GIVEN** the repository's existing MegaLinter and golangci-lint configuration
- **WHEN** `golangci-lint run ./...` scans all default-build-tag packages under
  the module root
- **THEN** the command MUST exit successfully with no findings
- **AND** every baseline inventory row MUST have a resolved disposition
- **AND** the enabled linters and their severity settings MUST remain unchanged

#### Scenario: Resource or operation failure occurs

- **GIVEN** code that closes a resource, scans SQL data, manages a transaction,
  or invokes a subprocess or integration
- **WHEN** an operation or cleanup step returns an error
- **THEN** actionable scan, transaction work, commit, subprocess, and integration
  errors MUST be propagated or combined according to the operation's contract
- **AND** only expected, non-actionable cleanup errors MAY be explicitly
  discarded with a concise call-site justification
- **AND** transaction commit and rollback behavior MUST remain safe
- **AND** MCP fields, response shapes, error codes and categories, HTTP mappings,
  CLI exit codes, and parity fixtures MUST remain compatible
- **AND** public errors MUST NOT newly expose credentials, authorization headers,
  URL user information or query values, request or response bodies, untrusted
  subprocess output, or newly exposed absolute paths

#### Scenario: Mechanical lint correction is applied

- **GIVEN** a reproducing formatting, vet, static analysis, critic, unused-code,
  or ineffective-assignment diagnostic
- **WHEN** the finding is corrected
- **THEN** a production correction that changes behavior or control flow MUST
  begin with a failing Go regression test
- **AND** a strictly mechanical production edit or test-only correction MUST use
  the reproducing lint diagnostic as its failing automated check and MUST keep
  existing package tests green before and after the edit
- **AND** the smallest semantics-preserving correction MUST be used
- **AND** unrelated refactoring or dependency changes MUST NOT be introduced

#### Scenario: Corrected behavior is verified

- **GIVEN** a lint correction changes error propagation or control flow
- **WHEN** the affected package tests run
- **THEN** the durable test matrix MUST identify the failure fixture, expected
  error chain, state invariants, sensitive-data boundary, and success path
- **AND** focused regression tests MUST verify those success and failure paths
- **AND** database tests MUST use in-memory SQLite
- **AND** git and filesystem tests MUST use temporary directories
- **AND** HTTP tests MUST use `httptest.NewServer`
- **AND** no test MUST depend on shared mutable state or another test's side effects

#### Scenario: Response parity is verified

- **GIVEN** error-handling corrections are complete
- **WHEN** `go test -tags parity ./test/parity/... -count=1 -race` runs
- **THEN** all response-shape comparisons MUST pass
- **AND** response fixtures MUST remain unchanged unless an independently
  approved specification authorizes a contract change

#### Scenario: Repository quality gates run

- **GIVEN** all implementation steps and finding resolutions are complete
- **WHEN** the sole implementation task is considered for completion
- **THEN** formatting, direct lint, vet, race tests, coverage ratchets,
  vulnerability scanning, parity tests, and the production build MUST pass
- **AND** local MegaLinter execution MUST run when `mega-linter-runner` is
  available, otherwise its unavailability MUST be recorded
- **AND** after PR creation, the pushed change MUST obtain authoritative hosted
  evidence from the `Standardized CI / Run linters` check before merge; this
  post-PR evidence is not a task-completion prerequisite
- **AND** coverage thresholds, race flags, CI gates, and other protected quality
  values MUST remain unchanged

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
<!-- scaffolded by uf v0.17.0 -->
