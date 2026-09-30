<!--
  This change has one completion checkbox because repository policy requires the
  full CI-equivalent gate before any task is marked complete, while the lint gate
  remains red until all packages are corrected. Per-finding progress, ownership,
  tests, and resumability live in lint-findings.md. Ordered steps below are
  implementation guidance, not independently completable task checkboxes.
-->

## 1. Complete The Project-Wide Lint Remediation

Before source or test edits:

1. Confirm `lint-findings.md` contains all 54 baseline rows and eight concrete
   matrix entries covering every behavioral row.
2. Confirm `tasks.md` contains `<!-- spec-review: passed -->` from an approved
   review and `.unleash-checklist.md` records the completed spec-review step.
3. Commit and push the reviewed proposal, design, delta spec, tasks, baseline-complete
   `lint-findings.md` (54 open rows and eight matrices), and finalized workflow
   checklist on `opsx/resolve-project-wide-lint-debt`.
4. Do not edit production or test code until that commit and push are complete.

Implementation sequence:

1. Add focused failing tests, then correct SQL scan, row lifecycle, transaction
   commit, and rollback findings in org, comms, database, and query packages.
2. Add focused failing tests, then correct subprocess, filesystem, git, HTTP,
   and integration findings in CLI, forge, gitutil, memory, and MCP client
   packages. Preserve public status, exit, and structured error contracts and do
   not expose sensitive values or newly expose absolute paths.
3. Add focused failing tests where observable behavior changes, then correct
   operation, cleanup, and writer findings in MCP, tests, and parity helpers
   without changing JSON response shapes or fixtures. Update `GenerateReport`
   GoDoc and all callers if its test-only signature returns an error.
4. For every production behavior or control-flow change, add a failing Go test
   first and apply the smallest implementation that makes it pass. For strictly
   mechanical production edits and test-only corrections, retain green package
   tests and use the reproducing diagnostic as the failing automated check.
5. Resolve every inventory row and confirm a final `golangci-lint run ./...`
   reports zero findings.

Completion gate:

1. Run `gofmt` and `goimports` on changed Go files and confirm
   `gofmt -l $(git ls-files '*.go')` produces no output.
2. Run `go vet ./...`.
3. Run `go test ./... -count=1 -race -coverprofile=coverage.out`.
4. Run `go test -tags parity ./test/parity/... -count=1 -race` and confirm fixtures and
   response shapes remain unchanged.
5. Run `make check-coverage`, then separately enforce the missing CI ratchet
   `internal/mcpclient >= 80%` from the same `coverage.out` with a
   fail-closed reproduction of CI's per-function percentage average:

   ```bash
   set -o pipefail
   PKG_COV=$(go tool cover -func=coverage.out | awk '/github.com\/unbound-force\/replicator\/internal\/mcpclient/ { value = $3; sub(/%$/, "", value); total += value; count++ } END { if (count == 0) { print "no internal/mcpclient coverage records" > "/dev/stderr"; exit 2 } printf "%.1f", total / count }') || exit 1
   printf 'internal/mcpclient: %s%%\n' "$PKG_COV"
   awk -v cov="$PKG_COV" 'BEGIN { if (cov + 0 < 80) exit 1 }'
   ```

   Do not modify either protected threshold source as part of this change.
6. Install govulncheck at commit `3e6f44f962742443c11ae2261f02e0c917aeb2bc`,
   then run `govulncheck ./...`, `go build -o bin/replicator ./cmd/replicator`,
   `.github/scripts/patch-homebrew-cask_test.sh`, and `make check` with Go 1.25.13.
7. Run `mega-linter-runner` against `.mega-linter.yml` when available; otherwise
   record its unavailability. Treat the hosted `Standardized CI / Run linters`
   result as a post-PR, pre-merge gate rather than task-completion evidence.
8. Verify constitution alignment: independent MCP calls, self-describing and
   shape-compatible responses, standalone operation, graceful degradation, and
   isolated testability all remain intact.
9. Confirm all required README, AGENTS, GoDoc, main-spec, and website documentation
   updates are complete, including `GenerateReport` GoDoc if its signature changes.
   Also confirm no protected gate, dependency, runtime public API, or unrelated
   file changed. If the diff gains user-facing behavior, stop and satisfy the
   website documentation gate before completion.

- [ ] 1.1 Complete all ordered remediation steps, resolve the durable finding
  inventory, and pass every completion gate before marking this task complete.
<!-- scaffolded by uf v0.17.0 -->
<!-- spec-review: passed -->
