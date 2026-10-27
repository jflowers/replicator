---
tag: spec-quality
author: jay-flowers
category: pattern
created_at: 2026-10-26T21:29:38Z
identity: spec-quality-20261026T212938-jay-flowers
tier: draft
---

In multi-iteration OpenSpec reviews, when a fix round settles a design decision that narrows an earlier acceptance-scenario gloss, the spec.md scenario must be reworded in the same round. Example from replicator change init-smart-sync (3rd review iteration): design D3 and task 2.2 specified that the new `updated` ScaffoldResult action reuses the warn style previously used for `overwritten`, but the spec.md scenario still required rendering `updated` "with a distinct style" while the requirement enumerated `overwritten` as a distinguished peer — an acceptance oracle the planned implementation could not satisfy. Every other cross-artifact claim in that change verified cleanly against repo state; only the stale scenario gloss survived as a spec-vs-plan contradiction. Acceptance scenarios are verification oracles: propagate decision-level wording changes into them immediately, or later iterations and implementation-time verification will fail on the contradiction.
