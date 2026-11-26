---
tag: ci-gating
author: jay-flowers
category: gotcha
created_at: 2026-11-25T23:02:53Z
identity: ci-gating-20261125T230253-jay-flowers
tier: draft
as_of: "replicator@005c9ebe5c0bf4f8854b4a3b24c38d2d28083b3b"
---

When a config contract enforced by a structure test moves to another repository or system, the test rewrite and the spec mandate it implements must move together. A requirement worded like "the test MUST fail when this wiring is removed or weakened" (replicator CLS-FR-006, ci-skip-build-for-learnings) keeps claiming regression coverage after the wiring legitimately migrates cross-repo, unless a spec change retires or re-points the mandate in the same PR. Observed on PR #205 (retire-classic-protection): settings.yml branch protection moved to safe-settings in unbound-force/.github, the structure test was rewritten to pin only the new in-repo state, and the live spec still mandates the deleted assertion — the repo's regression net for main's required checks silently became empty while the spec still asserts it exists. Pattern: on such a migration, open the spec change first (a MODIFIED requirement naming the new enforcement source and the replacement verification the repo can still perform), then rewrite the test to pin exactly that.
