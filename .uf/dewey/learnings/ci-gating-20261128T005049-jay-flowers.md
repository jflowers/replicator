---
tag: ci-gating
author: jay-flowers
category: gotcha
created_at: 2026-11-28T00:50:49Z
identity: ci-gating-20261128T005049-jay-flowers
tier: draft
as_of: "replicator@1417eab74bb6c10984026580a858fc0c2537cf1f"
---

A pinact ignore rule for a pre-release pin is self-expiring but not self-narrowing: when the pin is later bumped to a tagged release the rule stops matching and normal verification resumes, but broadening the rule (dropping the ActionVersion condition, widening ActionName, adding rules) still exits 0 — pinact verifies the pins it checks, never the ones it was told to skip. Pin the rule's shape with a workflowtest that binds the ignore rule's ActionName and ActionVersion to the exact SHA in the workflow's uses: line (parse the SHA from the workflow so the two files cannot drift), the same treatment replicator gives its zizmor exception (TestFullsendZizmorException_IsNarrowAndDocumented).
