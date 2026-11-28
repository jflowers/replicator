---
tag: ci-gating
author: jay-flowers
category: pattern
created_at: 2026-11-28T00:50:46Z
identity: ci-gating-20261128T005046-jay-flowers
tier: draft
as_of: "replicator@1417eab74bb6c10984026580a858fc0c2537cf1f"
---

Delegating a CI gate's decision logic to a pinned reusable workflow converts the consumer repo's regression net from behavioral to wiring-only. The pin freezes behavior and the aggregate verdict can reject malformed outputs, but nothing local detects a classifier producing well-formed wrong output (labeling a code PR a chore skips heavy jobs while the always() aggregate stays green). Close the gap with evidence, not comments: declare the owning repo in .uf/sibling-repos.yaml with the workflow and its tests as contract paths so acquire_sibling_evidence can verify coverage at review time, or record the relied-upon contract surface (output value set, event semantics like push yields 'true', fail-closed behavior) in a doc a structure test cross-references — replicator's TestFullsendZizmorException_IsNarrowAndDocumented pattern.
