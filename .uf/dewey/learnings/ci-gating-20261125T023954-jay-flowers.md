---
tag: ci-gating
author: jay-flowers
category: gotcha
created_at: 2026-11-25T02:39:54Z
identity: ci-gating-20261125T023954-jay-flowers
tier: draft
as_of: "replicator@247dfce68b5a4fed8fe299a785991ebc8eb44ab4"
---

A GitHub Actions aggregate job used as the single required branch-protection check only covers jobs listed in its `needs`; adding a new CI job later does not auto-extend the gate, so a failing new job can leave the protected aggregate green. Guard this with a structure test asserting `ci-ok.needs` covers every job in the workflow, and document that new gated jobs must be appended to the aggregate's `needs`.
