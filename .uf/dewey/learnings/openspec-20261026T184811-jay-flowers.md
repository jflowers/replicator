---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-10-26T18:48:11Z
identity: openspec-20261026T184811-jay-flowers
tier: draft
---

A change proposal's motivating inventory (e.g., "N of M mirrored files drifted") must be re-verified against the repository at authoring time. In replicator's init-smart-sync proposal, the claimed 6-of-15 drift (four commands missing <protect> tags, two agents missing delegation-contract updates) was stale — prior changes had already synced those pairs; the actual drift was 1 of 15 (the always-on-guidance skill), and that pair was already covered by TestSkillFiles_DriftDetection, contradicting the proposal's "CI cannot catch it" claim. Stale inventories make downstream dogfood-task acceptance criteria unachievable (task 4.2 expected 6 files reported "updated"; only 1 would be) and mischaracterize CI coverage. Always byte-verify every counted pair before writing counts into spec artifacts.
