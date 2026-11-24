---
tag: ci-workflow-gating
author: jay-flowers
category: pattern
created_at: 2026-11-24T17:35:57Z
identity: ci-workflow-gating-20261124T173557-jay-flowers
tier: draft
as_of: "replicator@a3390914a2d4a8e6ca5a0a93ecfdb68452d59f62"
---

To skip expensive CI jobs for a restricted path set (e.g. learnings-only PRs) without leaving required checks unreported, use job-level `if` gating plus an always-running aggregate job (`if: always()`, `needs:` covering the detector and every gated job, failing on `failure`/`cancelled`). Never use workflow-level `paths`/`paths-ignore`, which leaves required checks in "Expected — Waiting" forever. Two gotchas: (1) a skipped reusable-workflow call never emits its nested check, so the aggregate is mandatory; (2) any event type not explicitly handled (e.g. `merge_group`) falls into the pull_request branch with empty SHAs and fails closed — add an explicit handler (treat as code=true) before adopting a merge queue.
