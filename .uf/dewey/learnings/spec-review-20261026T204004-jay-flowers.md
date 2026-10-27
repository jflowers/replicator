---
tag: spec-review
author: jay-flowers
category: pattern
created_at: 2026-10-26T20:40:04Z
identity: spec-review-20261026T204004-jay-flowers
tier: draft
---

OpenSpec review gap pattern: map every scenario in a spec delta to a named verifying task before approving a change. Combined-flag scenarios are the classic miss — each flag gets its own test, but the interaction case stated as a MUST (e.g., a deprecated --force combined with --skip-existing, where skip-existing must take precedence) appears in no task. A wiring like `if force { Sync } else if skip { Skip }` silently violates the precedence and passes every single-flag test. Observed in openspec/changes/init-smart-sync iteration 2: the spec's 'Combined flags — skip-existing wins' scenario had four named tests in tasks.md 2.1, none exercising the combination.
