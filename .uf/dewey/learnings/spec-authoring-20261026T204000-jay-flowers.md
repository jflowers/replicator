---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-10-26T20:40:00Z
identity: spec-authoring-20261026T204000-jay-flowers
tier: draft
---

When authoring OpenSpec tasks.md in replicator, close the loop with two audits before finalizing: (1) enumerate every `#### Scenario:` block in the change's spec.md and confirm each maps to at least one named test task — flag-combination edge cases (e.g., `--force --skip-existing` precedence in opsx/init-smart-sync) are the ones that fall between per-flag test tasks; (2) inventory every existing test whose asserted behavior the change inverts or deletes, and give each an explicit disposition (rewrite/remove/replace) at every layer where it lives — unit and CLI — not just one. In opsx/init-smart-sync iteration 2, these two audits surfaced the remaining gaps after 13 prior findings were fixed: the combined-flags precedence scenario had no task, and the unit-layer TestScaffold_SkipsExisting/TestScaffold_ForceOverwrites dispositions were unspecified while their cmd-level twins (TestRunInit_AgentKitSkipsExisting/TestRunInit_ForceOverwrites) got explicit rewrite instructions in task 2.1.
