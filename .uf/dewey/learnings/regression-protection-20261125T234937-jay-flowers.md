---
tag: regression-protection
author: jay-flowers
category: gotcha
created_at: 2026-11-25T23:49:37Z
identity: regression-protection-20261125T234937-jay-flowers
tier: draft
as_of: "replicator@ddcb67791f79260a6be500fa34a1f5ef413b76ba"
---

Rewriting a regression test to match a new architecture silently deletes spec-mandated coverage when the pinning spec stays live. When a live, unarchived OpenSpec/Speckit requirement mandates what a test asserts (e.g. CLS-FR-006 requiring the workflowtest suite to pin .github/settings.yml required_status_checks.contexts), moving the enforced contract elsewhere — such as branch protection moving to safe-settings in an org .github repo — requires the same PR to open MODIFIED requirement deltas that re-point the enforcement source and define what the repo still verifies. Without the delta, make check passes while the spec claims regression protection that no longer exists, and the in-repo regression net for the moved contract is gone with nothing failing. Before inverting or deleting a spec-cited assertion, grep the live spec corpus (openspec/changes/*/specs/) and the design docs for the test's name and the asserted tokens.
