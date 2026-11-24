---
tag: opsx-ci-skip-build-for-learnings
author: jay-flowers
category: pattern
created_at: 2026-11-24T17:40:49Z
identity: opsx-ci-skip-build-for-learnings-20261124T174049-jay-flowers
tier: draft
as_of: "replicator@a3390914a2d4a8e6ca5a0a93ecfdb68452d59f62"
---

On branch opsx/ci-skip-build-for-learnings (2026-11-24), the /uf.unleash Step 6 code review council returned APPROVE WITH ADVISORIES (7/7 reviewers succeeded) while the compact consolidation result carried the scalar canonical verdict ESCALATED — the artifact's verdict_reason showed one MEDIUM finding noted as an advisory (the ci_gating_test.go regression test pins job-graph wiring but not the detection script's safety tokens). The correct handling, matching the escalated_advisory_unrelayed enforcement rule, was to stop before proceeding to the retrospective and relay the escalated advisory to the human through the question tool with the full advisory substance, the minimal-changes-between-APPROVE-and-PR tradeoff, and three options (record and proceed / apply fix and re-review / pause). The human chose to record and proceed, preserving the council's approving verdict with no post-APPROVE code changes; the advisory is carried into the final report and PR body with the test-strengthening deferred to a follow-up change.
