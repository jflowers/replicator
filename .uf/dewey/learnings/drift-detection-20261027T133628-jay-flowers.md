---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-10-27T13:36:28Z
identity: drift-detection-20261027T133628-jay-flowers
tier: draft
---

Mirrored-file byte-identity at PR head is provable from diff metadata alone, without reading files or running the drift tests: when a PR modifies both sides of a mirrored pair, each side's hunk header `index <old>..<new>` names the post-change git blob, and matching `new` blobs across the pair's two hunks prove the pair converges byte-identically at head. Used in the replicator init-smart-sync code review (2026-10-27) to confirm the always-on-guidance pair (both hunks ending at blob f28b27d) and the replicator-cli pair (both at 2d99f9d). For pairs a PR does not touch, head state equals base, so base-side verification suffices — read both copies with limit=1 and compare the read tool's full-file hashline hash, validating the method against a CI-enforced pair (forge-global) first. Limits: differing new-blobs expose one-sided sync; matching blobs prove content convergence only, not file-set convergence (smart-sync leaves orphaned scaffolded files in place by design).
