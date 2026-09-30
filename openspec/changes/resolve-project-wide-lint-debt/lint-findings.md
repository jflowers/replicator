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

| ID | Location | Linter | Class | Planned correction | Test / justification | Owner | Disposition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| L001 | `cmd/replicator/docs.go:60` | errcheck | behavioral | Return a contextual file-close error without replacing an earlier write error | Failing close fixture; see M1 | cmd/replicator | open |
| L002 | `cmd/replicator/serve.go:30` | errcheck | behavioral | Preserve the serve result and surface an actionable log-close error through the existing command error path | Failing closer fixture; see M2 | cmd/replicator | open |
| L003 | `cmd/replicator/serve_test.go:20` | errcheck | mechanical | Fail cleanup when restoring the working directory fails | Lint red; existing serve test remains green | cmd/replicator | open |
| L004 | `cmd/replicator/serve_test.go:24` | errcheck | mechanical | Check the test closer error | Lint red; existing serve test remains green | cmd/replicator | open |
| L005 | `cmd/replicator/serve_test.go:45` | errcheck | mechanical | Fail cleanup when restoring the working directory fails | Lint red; existing serve test remains green | cmd/replicator | open |
| L006 | `cmd/replicator/serve_test.go:49` | errcheck | mechanical | Fail setup on directory creation error | Lint red; existing serve test remains green | cmd/replicator | open |
| L007 | `cmd/replicator/serve_test.go:51` | errcheck | mechanical | Fail setup on marker write error | Lint red; existing serve test remains green | cmd/replicator | open |
| L008 | `internal/comms/reservation.go:37` | errcheck | behavioral | Ignore only `sql.ErrTxDone`; combine a pre-commit rollback failure with the primary operation error | Transaction finalizer regression; see M8 | internal/comms | open |
| L009 | `internal/db/db.go:32` | errcheck | behavioral | Join ping failure with an actionable database-close failure | Failing closer seam; see M3 | internal/db | open |
| L010 | `internal/db/db.go:38` | errcheck | behavioral | Join disk migration failure with an actionable database-close failure | Failing closer seam; see M3 | internal/db | open |
| L011 | `internal/db/db.go:53` | errcheck | behavioral | Join in-memory migration failure with an actionable database-close failure | Failing closer seam; see M3 | internal/db | open |
| L012 | `internal/db/db_test.go:10` | errcheck | mechanical | Check store cleanup with `t.Cleanup` | Lint red; database tests remain green | internal/db | open |
| L013 | `internal/db/db_test.go:35` | errcheck | mechanical | Assert explicit store close succeeds | Lint red; database tests remain green | internal/db | open |
| L014 | `internal/forge/worktree.go:98` | errcheck | behavioral | Return contextual worktree-removal failure instead of reporting success | Temp git failure fixture; see M4 | internal/forge | open |
| L015 | `internal/forge/worktree_test.go:89` | errcheck | mechanical | Fail setup on worktree file write error | Lint red; worktree test remains green | internal/forge | open |
| L016 | `internal/forge/worktree_test.go:90` | errcheck | mechanical | Fail setup on git add error | Lint red; worktree test remains green | internal/forge | open |
| L017 | `internal/forge/worktree_test.go:91` | errcheck | mechanical | Fail setup on git commit error | Lint red; worktree test remains green | internal/forge | open |
| L018 | `internal/gitutil/git_test.go:128` | errcheck | mechanical | Assert worktree creation succeeds | Lint red; git test remains green | internal/gitutil | open |
| L019 | `internal/gitutil/git_test.go:148` | errcheck | mechanical | Assert worktree creation succeeds | Lint red; git test remains green | internal/gitutil | open |
| L020 | `internal/gitutil/git_test.go:152` | errcheck | mechanical | Fail setup on file write error | Lint red; git test remains green | internal/gitutil | open |
| L021 | `internal/gitutil/git_test.go:153` | errcheck | mechanical | Fail setup on git add error | Lint red; git test remains green | internal/gitutil | open |
| L022 | `internal/mcp/server_test.go:191` | errcheck | mechanical | Assert server request handling succeeds | Lint red; server test remains green | internal/mcp | open |
| L023 | `internal/mcpclient/client.go:160` | errcheck | mechanical | Explicitly discard non-actionable response-body close after full consumption | HTTP body already consumed; client tests remain green | internal/mcpclient | open |
| L024 | `internal/mcpclient/client.go:220` | errcheck | mechanical | Explicitly discard non-actionable response-body close after full consumption | HTTP body already consumed; client tests remain green | internal/mcpclient | open |
| L025 | `internal/mcpclient/client_test.go:103` | errcheck | mechanical | Fail handler test on JSON encode error | Lint red; client test remains green | internal/mcpclient | open |
| L026 | `internal/mcpclient/client_test.go:207` | errcheck | mechanical | Fail handler test on response write error | Lint red; client test remains green | internal/mcpclient | open |
| L027 | `internal/mcpclient/client_test.go:300` | errcheck | mechanical | Fail handler test on request decode error | Lint red; client test remains green | internal/mcpclient | open |
| L028 | `internal/mcpclient/client_test.go:483` | errcheck | mechanical | Fail handler test on request decode error | Lint red; client test remains green | internal/mcpclient | open |
| L029 | `internal/mcpclient/client_test.go:554` | errcheck | mechanical | Fail handler test on request decode error | Lint red; client test remains green | internal/mcpclient | open |
| L030 | `internal/memory/proxy_test.go:221` | errcheck | mechanical | Fail test on argument unmarshal error | Lint red; proxy test remains green | internal/memory | open |
| L031 | `internal/memory/proxy_test.go:262` | errcheck | mechanical | Assert received argument unmarshal succeeds | Lint red; proxy test remains green | internal/memory | open |
| L032 | `internal/memory/proxy_test.go:288` | errcheck | mechanical | Fail test on argument unmarshal error | Lint red; proxy test remains green | internal/memory | open |
| L033 | `internal/memory/proxy_test.go:418` | errcheck | mechanical | Check response-body cleanup with `t.Cleanup` | Lint red; proxy test remains green | internal/memory | open |
| L034 | `internal/org/epic.go:31` | errcheck | behavioral | Ignore only `sql.ErrTxDone`; combine a pre-commit rollback failure with the primary operation error | Transaction finalizer regression; see M8 | internal/org | open |
| L035 | `internal/org/session_test.go:31` | errcheck | mechanical | Assert session count scan succeeds | Lint red; session test remains green | internal/org | open |
| L036 | `internal/query/presets.go:62` | errcheck | behavioral | Propagate actionable row-close failure while preserving query or scan errors | Failing rows seam; see M5 | internal/query | open |
| L037 | `internal/query/presets.go:98` | errcheck | behavioral | Propagate actionable row-close failure while preserving query or scan errors | Failing rows seam; see M5 | internal/query | open |
| L038 | `internal/query/presets.go:130` | errcheck | behavioral | Return contextual total-count scan failure through the two-count scanner seam | Injected first-call failure; see M6 | internal/query | open |
| L039 | `internal/query/presets.go:131` | errcheck | behavioral | Return contextual completed-count scan failure through the two-count scanner seam | Injected second-call failure; see M6 | internal/query | open |
| L040 | `internal/query/presets.go:154` | errcheck | behavioral | Propagate actionable row-close failure while preserving query or scan errors | Failing rows seam; see M5 | internal/query | open |
| L041 | `internal/query/presets_test.go:17` | errcheck | mechanical | Check store cleanup with `t.Cleanup` | Lint red; preset tests remain green | internal/query | open |
| L042 | `internal/query/presets_test.go:72` | errcheck | mechanical | Fail fixture setup on insert error | Lint red; preset test remains green | internal/query | open |
| L043 | `internal/query/presets_test.go:73` | errcheck | mechanical | Fail fixture setup on insert error | Lint red; preset test remains green | internal/query | open |
| L044 | `internal/query/presets_test.go:74` | errcheck | mechanical | Fail fixture setup on insert error | Lint red; preset test remains green | internal/query | open |
| L045 | `test/parity/report.go:27` | errcheck | behavioral | Return report header write errors | Failing writer regression; see M7 | test/parity | open |
| L046 | `test/parity/report.go:28` | errcheck | behavioral | Return report separator write errors | Failing writer regression; see M7 | test/parity | open |
| L047 | `test/parity/report.go:33` | errcheck | behavioral | Return successful-row write errors | Failing writer regression; see M7 | test/parity | open |
| L048 | `test/parity/report.go:39` | errcheck | behavioral | Return first-difference write errors | Failing writer regression; see M7 | test/parity | open |
| L049 | `test/parity/report.go:42` | errcheck | behavioral | Return subsequent-difference write errors | Failing writer regression; see M7 | test/parity | open |
| L050 | `test/parity/report.go:49` | errcheck | behavioral | Return footer write errors and update callers | Failing writer regression; see M7 | test/parity | open |
| L051 | `internal/forge/decompose.go:50` | staticcheck | mechanical | Use `fmt.Fprintf` and explicitly discard its impossible `strings.Builder` write error | Lint red; builder writes return nil errors; decomposition tests remain green | internal/forge | open |
| L052 | `internal/forge/decompose.go:51` | staticcheck | mechanical | Use `fmt.Fprintf` and explicitly discard its impossible `strings.Builder` write error | Lint red; builder writes return nil errors; decomposition tests remain green | internal/forge | open |
| L053 | `internal/forge/decompose.go:52` | staticcheck | mechanical | Use `fmt.Fprintf` and explicitly discard its impossible `strings.Builder` write error | Lint red; builder writes return nil errors; decomposition tests remain green | internal/forge | open |
| L054 | `internal/mcp/server.go:34` | unused | mechanical | Remove unused `nextID` field and import if no other use remains | Lint red; server tests remain green | internal/mcp | open |

## Behavioral Test Matrix

| Matrix | Finding IDs | Fixture | Success path | Failure path and expected chain | State invariants | Sensitive-data boundary |
| --- | --- | --- | --- | --- | --- | --- |
| M1 | L001 | Package-local `io.WriteCloser` seam configurable for write-only, close-only, and dual sentinel failures | Exact documentation bytes are written and close returns nil | Write-only returns `errWrite`; close-only returns `errClose`; dual failure preserves `errors.Is(err, errWrite)` and `errors.Is(err, errClose)` with `write docs`/`close docs` context | No successful result on any failure; successful file is complete | No file contents or newly exposed absolute path in the error |
| M2 | L002 | Injected log writer/closer configurable for serve-only, close-only, and dual sentinel failures | Existing serve return and successful close produce the current exit behavior | Serve-only returns `errServe`; close-only returns `errClose`; dual failure preserves both sentinels with static context | Cleanup always runs once; existing exit-code mapping remains | No log contents or newly exposed absolute path |
| M3 | L009-L011 | Table-driven package-local `closeOnOpenError(primary, closer)` seam using `errPing`, `errDiskMigration`, or `errMemoryMigration` plus `errClose`; existing integration cases retain real ping, disk-migration, and in-memory-migration stage coverage | Nil primary does not invoke error cleanup; successful opens remain usable | Each stage sentinel remains discoverable alone and together with `errClose` | Failed open returns no store; fake closer is called exactly once | Static stage name only; no DSN or path values |
| M4 | L014 | Inject a package-local worktree remover function into the delete helper; run the git-backed success case under `t.TempDir()` with `testing.Short()` guard and return `errRemove` from the injected failure case | Real temporary worktree removal returns success | Injected `errRemove` is wrapped as `remove worktree` and remains discoverable | Branch cleanup is not reported successful; remover called exactly once | No subprocess output or newly exposed absolute path |
| M5 | L036-L037, L040 | Package-local query/rows seam configurable for scan-only, close-only, and dual sentinel failures | Exact existing table rows and order are returned | Scan-only returns `errScan`; close-only returns `errClose`; dual failure preserves both with `scan rows`/`close rows` context | No partial table is returned; close called exactly once | No SQL text or stored event payload in errors |
| M6 | L038-L039 | Extract a package-local two-count scanner seam; table cases return first-call `errTotal` and second-call `errCompleted` independently | Exact total and completed counts are returned | L038 case returns `errTotal` after one call; L039 case returns first count then `errCompleted` after exactly two calls | No fabricated counts or partial success | No SQL text, payload, or database path in errors |
| M7 | L045-L050 | Writer that fails after configurable byte counts | Existing parity report text and summary remain unchanged | Writer sentinel is returned from `GenerateReport` for header, row, difference, and footer failures | No false successful report is returned; parity fixtures and shapes are unchanged | No fixture body is added to error text |
| M8 | L008, L034 | Package-local transaction finalizer with injected `commit` and `rollback` functions plus an isolated `db.OpenMemory()` constraint-failure integration case; rollback is deferred exactly once immediately after begin | Success sequence is work, commit, rollback; counts are work 1, commit 1, rollback 1; rollback returns `sql.ErrTxDone`, which is ignored; exactly one row persists and no error is returned | Work failure sequence: work, rollback; counts 1/0/1, zero rows, returns `errWork`. Work plus rollback failure has the same counts and preserves `errWork` and `errRollback`. Commit failure sequence: work, commit, rollback; counts 1/1/1, zero rows, returns `errCommit`. Commit plus rollback failure has the same counts and preserves `errCommit` and `errRollback` | Commit is never called after work failure; rollback is always called exactly once after begin; only successful commit persists one row; every failure case persists zero rows | Static operation context only; no SQL or payload values |
