---
tag: uf-unleash
author: jay-flowers
category: pattern
created_at: 2026-10-28T15:28:57Z
identity: uf-unleash-20261028T152857-jay-flowers
tier: draft
---

When /uf.unleash review councils cannot use the task tool or the dispatch MCP tools (wrong-repo binding or missing task tool in the session), run them manually inline: derive the dispatch plan from .uf/review-matrix.yaml and reviewer-capabilities.yaml in the target repo, read each included reviewer's agent definition, embody each persona against the saved diff file, record plan/findings/consolidated verdict in session notes, and relay per-finding decision gates through the orchestrator's GATE:<name> end-of-turn protocol (details, per-option reasoning, recommendation printed above the gate block; recommended option first labeled "(Recommended)"). Verified on the question-tool-presentation-contract change in unbound-force: spec review found 2 real findings (1 MEDIUM design-internal D2/D4 wrap contradiction, 1 LOW docs-impact gap) via 4 reviewers; code review returned 5/5 APPROVE with 4 skipped.
