---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-25T21:25:59Z
identity: ci-cd-20261125T212559-jay-flowers
tier: draft
as_of: "replicator@156f81fc7d207fdd7e6b085703734a88dbae11dc"
---

When a repo pins two paired reusable workflows (e.g. fullsend's reusable-dispatch.yml as caller and reusable-prioritize.yml as callee) and Dependabot bumps only one, the resulting version skew is unvalidatable by the bumping PR's CI: the callee is workflow_dispatch-only so no PR check runs it, and per-file SHA/tag verification does not check cross-file pairing. GitHub validates a reusable-workflow `with:` block against the callee's declared inputs, so a newer callee with a required/changed input breaks the older caller's dispatch — and the reverse order breaks too. Bump paired pins atomically in one PR rather than in either single order.
