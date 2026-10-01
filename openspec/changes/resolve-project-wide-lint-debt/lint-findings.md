# Lint Finding Inventory

## Baseline

- Command: `golangci-lint run ./...`
- Tool: golangci-lint 2.11.4, built with Go 1.26.5
- Captured: 2026-09-30
- Revision: `0355205b6ef999ce873418c228ec2ea25c534390`
- Module toolchain: Go 1.25.13
- Configuration: `.mega-linter.yml` SHA-256
  `88967e73ef11b5cc75b516b58f337a0d35d131877f1b4a3dca6e70ab3b1cc7fa`;
  no repository `.golangci.yml` exists
- Result: 54 findings (`errcheck`: 50, `staticcheck`: 3, `unused`: 1)
- Ownership: the package containing the finding

Issue #119 reported an earlier 46-finding snapshot. This inventory uses the
current reproducible scan as the implementation baseline so completion means all
default-build-tag packages selected by `./...` are clean rather than only the
stale subset. Build-tagged parity code is verified separately with `-tags parity`.

## Row Schema

Each implementation row MUST contain:

| Field | Meaning |
| --- | --- |
| ID | Stable `LNNN` identifier |
| Location | Repository-relative file and line from the baseline |
| Linter | Emitting linter |
| Class | `behavioral` or `mechanical` |
| Correction | Planned or applied correction |
| Test / justification | Regression test, lint-as-red evidence, or explicit cleanup rationale |
| Owner | Affected package |
| Disposition | `open` or `resolved` |

A row transitions to `resolved` only in the same edit that records its final
correction and concrete test, lint, or cleanup-justification evidence. Partial
updates retain `open` so no row claims completion without provenance.

## Findings

Rows MUST NOT be deleted; implementation updates the correction, evidence, and
disposition fields in place.

| ID | Location | Linter | Class | Correction | Test / justification | Owner | Disposition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| L001 | `cmd/replicator/docs.go:60` | errcheck | behavioral | Added `writeDocsAndClose`; write and close failures are context-wrapped and joined | `TestWriteDocsAndClose_PreservesWriteAndCloseErrors`; `go test ./cmd/replicator` | cmd/replicator | resolved |
| L002 | `cmd/replicator/serve.go:30` | errcheck | behavioral | Added `runAndClose`; serve and log-close failures are preserved together | `TestRunAndClose_PreservesServeAndCloseErrors`; `go test ./cmd/replicator` | cmd/replicator | resolved |
| L003 | `cmd/replicator/serve_test.go:20` | errcheck | mechanical | Working-directory restoration now reports cleanup failure | `golangci-lint run ./...`: 0 issues; `go test ./cmd/replicator` | cmd/replicator | resolved |
| L004 | `cmd/replicator/serve_test.go:24` | errcheck | mechanical | Logger close is checked in test cleanup | `golangci-lint run ./...`: 0 issues; `go test ./cmd/replicator` | cmd/replicator | resolved |
| L005 | `cmd/replicator/serve_test.go:45` | errcheck | mechanical | Working-directory restoration now reports cleanup failure | `golangci-lint run ./...`: 0 issues; `go test ./cmd/replicator` | cmd/replicator | resolved |
| L006 | `cmd/replicator/serve_test.go:49` | errcheck | mechanical | Directory setup now fails the test on error | `golangci-lint run ./...`: 0 issues; `go test ./cmd/replicator` | cmd/replicator | resolved |
| L007 | `cmd/replicator/serve_test.go:51` | errcheck | mechanical | Marker-file setup now fails the test on error | `golangci-lint run ./...`: 0 issues; `go test ./cmd/replicator` | cmd/replicator | resolved |
| L008 | `internal/comms/reservation.go:37` | errcheck | behavioral | Added one deferred rollback; only post-commit `sql.ErrTxDone` is ignored and pre-commit failures are joined | `TestRunReservationTransaction_FinalizesExactlyOnce`; `go test ./internal/comms` | internal/comms | resolved |
| L009 | `internal/db/db.go:32` | errcheck | behavioral | Ping failure is joined with close failure through `closeOnOpenError` | `TestCloseOnOpenError_PreservesPrimaryAndCloseErrors`; `go test ./internal/db` | internal/db | resolved |
| L010 | `internal/db/db.go:38` | errcheck | behavioral | Disk migration failure is joined with close failure | `TestCloseOnOpenError_PreservesPrimaryAndCloseErrors`; `go test ./internal/db` | internal/db | resolved |
| L011 | `internal/db/db.go:53` | errcheck | behavioral | In-memory migration failure is joined with close failure | `TestCloseOnOpenError_PreservesPrimaryAndCloseErrors`; `go test ./internal/db` | internal/db | resolved |
| L012 | `internal/db/db_test.go:10` | errcheck | mechanical | Store cleanup now checks `Close` | `golangci-lint run ./...`: 0 issues; `go test ./internal/db` | internal/db | resolved |
| L013 | `internal/db/db_test.go:35` | errcheck | mechanical | Explicit store close is asserted | `golangci-lint run ./...`: 0 issues; `go test ./internal/db` | internal/db | resolved |
| L014 | `internal/forge/worktree.go:98` | errcheck | behavioral | Cleanup now returns contextual removal failure while retaining missing-path idempotence | `TestWorktreeCleanupWithRemover_ReturnsRemovalError`; `go test ./internal/forge` | internal/forge | resolved |
| L015 | `internal/forge/worktree_test.go:89` | errcheck | mechanical | File fixture write is checked | `golangci-lint run ./...`: 0 issues; `go test ./internal/forge` | internal/forge | resolved |
| L016 | `internal/forge/worktree_test.go:90` | errcheck | mechanical | Git add fixture step is checked | `golangci-lint run ./...`: 0 issues; `go test ./internal/forge` | internal/forge | resolved |
| L017 | `internal/forge/worktree_test.go:91` | errcheck | mechanical | Git commit fixture step is checked | `golangci-lint run ./...`: 0 issues; `go test ./internal/forge` | internal/forge | resolved |
| L018 | `internal/gitutil/git_test.go:128` | errcheck | mechanical | Worktree creation is asserted | `golangci-lint run ./...`: 0 issues; `go test ./internal/gitutil` | internal/gitutil | resolved |
| L019 | `internal/gitutil/git_test.go:148` | errcheck | mechanical | Cherry-pick worktree creation is asserted | `golangci-lint run ./...`: 0 issues; `go test ./internal/gitutil` | internal/gitutil | resolved |
| L020 | `internal/gitutil/git_test.go:152` | errcheck | mechanical | Cherry-pick fixture write is checked | `golangci-lint run ./...`: 0 issues; `go test ./internal/gitutil` | internal/gitutil | resolved |
| L021 | `internal/gitutil/git_test.go:153` | errcheck | mechanical | Cherry-pick git add is checked | `golangci-lint run ./...`: 0 issues; `go test ./internal/gitutil` | internal/gitutil | resolved |
| L022 | `internal/mcp/server_test.go:191` | errcheck | mechanical | Server execution is asserted | `golangci-lint run ./...`: 0 issues; `go test ./internal/mcp` | internal/mcp | resolved |
| L023 | `internal/mcpclient/client.go:160` | errcheck | mechanical | Post-consumption response close is explicitly discarded with rationale | `golangci-lint run ./...`: 0 issues; `go test ./internal/mcpclient` | internal/mcpclient | resolved |
| L024 | `internal/mcpclient/client.go:220` | errcheck | mechanical | Post-consumption response close is explicitly discarded with rationale | `golangci-lint run ./...`: 0 issues; `go test ./internal/mcpclient` | internal/mcpclient | resolved |
| L025 | `internal/mcpclient/client_test.go:103` | errcheck | mechanical | Handler JSON encode failure is reported | `golangci-lint run ./...`: 0 issues; `go test ./internal/mcpclient` | internal/mcpclient | resolved |
| L026 | `internal/mcpclient/client_test.go:207` | errcheck | mechanical | Handler response write failure is reported | `golangci-lint run ./...`: 0 issues; `go test ./internal/mcpclient` | internal/mcpclient | resolved |
| L027 | `internal/mcpclient/client_test.go:300` | errcheck | mechanical | Request decode failure is reported and stops the handler | `golangci-lint run ./...`: 0 issues; `go test ./internal/mcpclient` | internal/mcpclient | resolved |
| L028 | `internal/mcpclient/client_test.go:483` | errcheck | mechanical | Request decode failure is reported and stops the handler | `golangci-lint run ./...`: 0 issues; `go test ./internal/mcpclient` | internal/mcpclient | resolved |
| L029 | `internal/mcpclient/client_test.go:554` | errcheck | mechanical | Request decode failure is reported and stops the handler | `golangci-lint run ./...`: 0 issues; `go test ./internal/mcpclient` | internal/mcpclient | resolved |
| L030 | `internal/memory/proxy_test.go:221` | errcheck | mechanical | Argument unmarshal is asserted | `golangci-lint run ./...`: 0 issues; `go test ./internal/memory` | internal/memory | resolved |
| L031 | `internal/memory/proxy_test.go:262` | errcheck | mechanical | Received argument unmarshal is asserted | `golangci-lint run ./...`: 0 issues; `go test ./internal/memory` | internal/memory | resolved |
| L032 | `internal/memory/proxy_test.go:288` | errcheck | mechanical | Argument unmarshal is asserted | `golangci-lint run ./...`: 0 issues; `go test ./internal/memory` | internal/memory | resolved |
| L033 | `internal/memory/proxy_test.go:418` | errcheck | mechanical | Response-body cleanup is checked | `golangci-lint run ./...`: 0 issues; `go test ./internal/memory` | internal/memory | resolved |
| L034 | `internal/org/epic.go:31` | errcheck | behavioral | Added one deferred rollback; only post-commit `sql.ErrTxDone` is ignored and pre-commit failures are joined | `TestRunEpicTransaction_FinalizesExactlyOnce`; `go test ./internal/org` | internal/org | resolved |
| L035 | `internal/org/session_test.go:31` | errcheck | mechanical | Session count scan is asserted | `golangci-lint run ./...`: 0 issues; `go test ./internal/org` | internal/org | resolved |
| L036 | `internal/query/presets.go:62` | errcheck | behavioral | Row close failure is joined with scan/iteration failure and suppresses partial output | `TestReadTableRows_PreservesScanAndCloseErrors`; `go test ./internal/query` | internal/query | resolved |
| L037 | `internal/query/presets.go:98` | errcheck | behavioral | Row close failure is joined with scan/iteration failure and suppresses partial output | `TestReadTableRows_PreservesScanAndCloseErrors`; `go test ./internal/query` | internal/query | resolved |
| L038 | `internal/query/presets.go:130` | errcheck | behavioral | Total-count scan failure is context-wrapped | `TestScanForgeCounts_ReturnsEachScanError`; `go test ./internal/query` | internal/query | resolved |
| L039 | `internal/query/presets.go:131` | errcheck | behavioral | Completed-count scan failure is context-wrapped | `TestScanForgeCounts_ReturnsEachScanError`; `go test ./internal/query` | internal/query | resolved |
| L040 | `internal/query/presets.go:154` | errcheck | behavioral | Row close failure is joined with scan/iteration failure and suppresses partial output | `TestReadTableRows_PreservesScanAndCloseErrors`; `go test ./internal/query` | internal/query | resolved |
| L041 | `internal/query/presets_test.go:17` | errcheck | mechanical | Store cleanup now checks `Close` | `golangci-lint run ./...`: 0 issues; `go test ./internal/query` | internal/query | resolved |
| L042 | `internal/query/presets_test.go:72` | errcheck | mechanical | First fixture insert is checked | `golangci-lint run ./...`: 0 issues; `go test ./internal/query` | internal/query | resolved |
| L043 | `internal/query/presets_test.go:73` | errcheck | mechanical | Second fixture insert is checked | `golangci-lint run ./...`: 0 issues; `go test ./internal/query` | internal/query | resolved |
| L044 | `internal/query/presets_test.go:74` | errcheck | mechanical | Third fixture insert is checked | `golangci-lint run ./...`: 0 issues; `go test ./internal/query` | internal/query | resolved |
| L045 | `test/parity/report.go:27` | errcheck | behavioral | Header write errors are returned with context | `TestGenerateReport_ReturnsWriterFailures`; parity-tagged package test passed | test/parity | resolved |
| L046 | `test/parity/report.go:28` | errcheck | behavioral | Separator write errors are returned with context | `TestGenerateReport_ReturnsWriterFailures`; parity-tagged package test passed | test/parity | resolved |
| L047 | `test/parity/report.go:33` | errcheck | behavioral | Successful-row write errors are returned with context | `TestGenerateReport_ReturnsWriterFailures`; parity-tagged package test passed | test/parity | resolved |
| L048 | `test/parity/report.go:39` | errcheck | behavioral | First-difference write errors are returned with context | `TestGenerateReport_ReturnsWriterFailures`; parity-tagged package test passed | test/parity | resolved |
| L049 | `test/parity/report.go:42` | errcheck | behavioral | Subsequent-difference write errors are returned with context | `TestGenerateReport_ReturnsWriterFailures`; parity-tagged package test passed | test/parity | resolved |
| L050 | `test/parity/report.go:49` | errcheck | behavioral | Footer/summary errors are returned and the caller checks them | `TestGenerateReport_ReturnsWriterFailures`; `go test -tags parity ./test/parity/...` | test/parity | resolved |
| L051 | `internal/forge/decompose.go:50` | staticcheck | mechanical | Replaced formatting-through-`Sprintf` with `Fprintf`; impossible builder errors are explicit | `golangci-lint run ./...`: 0 issues; `go test ./internal/forge` | internal/forge | resolved |
| L052 | `internal/forge/decompose.go:51` | staticcheck | mechanical | Replaced formatting-through-`Sprintf` with `Fprintf`; impossible builder errors are explicit | `golangci-lint run ./...`: 0 issues; `go test ./internal/forge` | internal/forge | resolved |
| L053 | `internal/forge/decompose.go:52` | staticcheck | mechanical | Replaced formatting-through-`Sprintf` with `Fprintf`; impossible builder errors are explicit | `golangci-lint run ./...`: 0 issues; `go test ./internal/forge` | internal/forge | resolved |
| L054 | `internal/mcp/server.go:34` | unused | mechanical | Removed unused `nextID` and `sync/atomic` import | `golangci-lint run ./...`: 0 issues; `go test ./internal/mcp` | internal/mcp | resolved |

## Behavioral Test Matrix

| Matrix | Finding IDs | Fixture | Success path | Failure path and expected chain | State invariants | Sensitive-data boundary |
| --- | --- | --- | --- | --- | --- | --- |
| M1 | L001 | Package-local `io.WriteCloser` seam configurable for write-only, close-only, and dual sentinel failures | Exact documentation bytes are written and close returns nil | Write-only returns `errWrite`; close-only returns `errClose`; dual failure preserves `errors.Is(err, errWrite)` and `errors.Is(err, errClose)` with `write docs`/`close docs` context | No successful result on any failure; successful file is complete | No file contents or newly exposed absolute path in the error |
| M2 | L002 | Injected log writer/closer configurable for serve-only, close-only, and dual sentinel failures | Existing serve return and successful close produce the current exit behavior | Serve-only returns `errServe`; close-only returns `errClose`; dual failure preserves both sentinels with static context | Cleanup always runs once; existing exit-code mapping remains | No log contents or newly exposed absolute path |
| M3 | L009-L011 | Table-driven package-local `closeOnOpenError(primary, closer)` seam using contextual ping, disk-migration, and memory-migration sentinels plus `errClose`; existing integration cases verify successful disk and in-memory opens | Nil primary does not invoke error cleanup; successful opens remain usable | Each contextual stage sentinel remains discoverable alone and together with `errClose` | Failed-open cleanup returns the primary error; fake closer is called exactly once | Static stage name only; no DSN or path values |
| M4 | L014 | Inject a package-local worktree remover function into the delete helper; run the git-backed success case under `t.TempDir()` with `testing.Short()` guard and return `errRemove` from the injected failure case | Real temporary worktree removal returns success | Injected `errRemove` is wrapped as `remove worktree` and remains discoverable | Branch cleanup is not reported successful; remover called exactly once | No subprocess output or newly exposed absolute path |
| M5 | L036-L037, L040 | Package-local query/rows seam configurable for scan-only, close-only, and dual sentinel failures | Exact existing table rows and order are returned | Scan-only returns `errScan`; close-only returns `errClose`; dual failure preserves both with `scan rows`/`close rows` context | No partial table is returned; close called exactly once | No SQL text or stored event payload in errors |
| M6 | L038-L039 | Extract a package-local two-count scanner seam; table cases return first-call `errTotal` and second-call `errCompleted` independently | Exact total and completed counts are returned | L038 case returns `errTotal` after one call; L039 case returns first count then `errCompleted` after exactly two calls | No fabricated counts or partial success | No SQL text, payload, or database path in errors |
| M7 | L045-L050 | Writer that fails after configurable byte counts | Existing parity report text and summary remain unchanged | Writer sentinel is returned from `GenerateReport` for header, row, difference, and footer failures | No false successful report is returned; parity fixtures and shapes are unchanged | No fixture body is added to error text |
| M8 | L008, L034 | Package-local transaction finalizer with injected `commit` and `rollback` functions plus an isolated `db.OpenMemory()` constraint-failure integration case; rollback is deferred exactly once immediately after begin | Success sequence is work, commit, rollback; counts are work 1, commit 1, rollback 1; rollback returns `sql.ErrTxDone`, which is ignored; exactly one row persists and no error is returned | Work failure sequence: work, rollback; counts 1/0/1, zero rows, returns `errWork`. Work plus rollback failure has the same counts and preserves `errWork` and `errRollback`. Commit failure sequence: work, commit, rollback; counts 1/1/1, zero rows, returns `errCommit`. Commit plus rollback failure has the same counts and preserves `errCommit` and `errRollback` | Commit is never called after work failure; rollback is always called exactly once after begin; only successful commit persists one row; every failure case persists zero rows | Static operation context only; no SQL or payload values |

## Verification Notes

- `mega-linter-runner` was unavailable locally; direct `golangci-lint run ./...` completed with zero findings. Hosted MegaLinter remains the authoritative post-PR gate.
