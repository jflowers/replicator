---
tag: go-toolchain
author: jay-flowers
category: pattern
created_at: 2026-11-11T14:37:51Z
identity: go-toolchain-20261111T143751-jay-flowers
tier: draft
---

Replicator remediates Go standard-library CVEs with a one-line `go` directive bump in go.mod (1.26.10 to 1.26.11 for GO-2026-6859) plus a CHANGELOG "### Security" entry — no workflow changes, because ci.yml selects Go via `go-version-file: go.mod` and the pinned reusable release workflow holds the same contract; no GOTOOLCHAIN is set anywhere in .github/. The bump operates inside the "Go 1.26.9 or later" baseline that openspec/changes/fix-go-toolchain-vulnerabilities established, whose design explicitly treats 1.26.9 as a minimum ("not a ban on later patched releases") and whose task 1.2 prescribes the bump pattern (`go` directive only, no separate `toolchain` directive, no dependency changes) — so a patch-floor raise needs no new spec. Verification is CI's pinned govulncheck step passing on the head SHA; stdlib CVE findings reproducible under the go.mod-declared toolchain are repository regressions, not scanner environment discrepancies.
