---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-11T14:37:48Z
identity: supply-chain-20261111T143748-jay-flowers
tier: draft
---

When remediating a Go stdlib advisory (GO-YYYY-NNNN) via a go.mod `go` directive bump, the fix's real blast radius equals the toolchain-propagation surface. Before accepting the bump as remediation, verify five things: (1) CI's setup-go uses `go-version-file: go.mod`; (2) the release pipeline — including delegated/reusable workflows — holds the same contract with no caller-side go-version override; (3) no `GOTOOLCHAIN` override (e.g. `local`) exists in any workflow, which would pin an older toolchain despite the go.mod bump; (4) no parallel version pin exists elsewhere (`.go-version`, Dockerfile `golang:` image, `go.work`); (5) the pinned `govulncheck` gate actually ran and passed on the PR head SHA under the new toolchain. In unbound-force/replicator PR #163 (GO-2026-6859, net/http MaxBytesReader header race, fixed in go1.26.11), all five held: ci.yml pins setup-go by SHA with `go-version-file: go.mod`, release.yml delegates to SHA-pinned org-infra reusable workflows with the same contract, and a green govulncheck run on the head SHA is the authoritative closure signal. Note that issue bodies can carry false environment claims (issue #162 asserted CI sets GOTOOLCHAIN=local; no such setting exists) — verify pipeline mechanics from the workflow files, not from issue prose.
