---
tag: ci-gating
author: jay-flowers
category: pattern
created_at: 2026-11-26T16:11:02Z
identity: ci-gating-20261126T161102-jay-flowers
tier: draft
as_of: "replicator@719e81c289339c0fde0d429a37cec45caaa0718f"
---

When a regression test hand-rolls a line parser for a declarative config file (replicator's classicProtectionRetired over .github/settings.yml in internal/workflowtest/ci_gating_test.go), the parser's accepted grammar silently becomes part of the quality gate while being narrower than the real format: only some null spellings, one list-item layout, first-match-wins entry lookup. Pin those boundaries in the table test — cases for the null spellings you reject, the bare-key form, duplicate entries — so a future failure distinguishes 'the config regressed' from 'the config was reformatted', and keep every divergence fail-closed (unrecognized forms produce errors, never silent passes). Unpinned grammar boundaries invite either false passes (a duplicate entry satisfying the first match) or false failures that pressure maintainers to weaken the gate — the outcome AGENTS.md's gatekeeping-value protection exists to prevent.
