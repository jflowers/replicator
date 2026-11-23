---
tag: toolchain
author: jay-flowers
category: gotcha
created_at: 2026-10-28T15:29:01Z
identity: toolchain-20261028T152901-jay-flowers
tier: draft
---

gofmt from a local go1.26.5 toolchain flags internal/scaffold/scaffold_test.go around L3144 (map-literal alignment) in unbound-force@chai/unbound-force even at HEAD, while go.mod declares go 1.25.0 and CI (setup-go with go-version-file: go.mod) is green. Pre-existing environment discrepancy, not a repo regression: before treating local gofmt/govulncheck output as a repository regression, compare `go version` with the go.mod directive and CI's setup-go, and check whether HEAD is flagged identically. Do not fix or block unrelated changes on such findings.
