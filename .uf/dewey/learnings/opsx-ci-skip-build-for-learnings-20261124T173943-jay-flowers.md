---
tag: opsx-ci-skip-build-for-learnings
author: jay-flowers
category: gotcha
created_at: 2026-11-24T17:39:43Z
identity: opsx-ci-skip-build-for-learnings-20261124T173943-jay-flowers
tier: draft
as_of: "replicator@a3390914a2d4a8e6ca5a0a93ecfdb68452d59f62"
---

On branch opsx/ci-skip-build-for-learnings (2026-11-24), dispatch_review_runs failed non-retryable input validation because session_metadata.plan entries omitted the nullable model field: the plan-entry schema requires every field (agent, decision, reason_code, reason, source, sequence, read_only, tier, model, variant, validation_errors) to be PRESENT, and a JSON null must be passed explicitly as "model": null rather than the key being dropped from the object. The top-level plan_entries array (which did include model: null) validated fine; only the session_metadata.plan copy failed. A failed validation dispatches nothing (correlation_id none, zero runs), so correcting the parameter serialization and re-issuing the same dispatch is safe and does not count as a second dispatch.
