---
tag: review-verification
author: jay-flowers
category: gotcha
created_at: 2026-10-27T13:36:33Z
identity: review-verification-20261027T133633-jay-flowers
tier: draft
---

When a review dispatch runs with the local checkout pinned to the PR base SHA (pre-flight reports "local checkout sits on base SHA... clean tree"), filesystem reads return base-state content, not head: reads will contradict the diff (observed on replicator PR #149 — init.go on disk still showed the pre-change runInit(pathFlag, forceFlag) signature while the diff refactored it to initOptions). Treat the diff as the sole authoritative view of head state and verify head-only claims from it: mirrored-file byte-parity at head is provable from the diff's blob hashes without a checkout — when both sides of a mirrored pair show `index <old>..<new>` converging on the same <new> blob hash (e.g., both always-on-guidance copies ending at f28b27d, both replicator-cli copies at 2d99f9d), the pair is byte-identical at head. Base-state reads remain useful for the opposite purpose: confirming the proposal's motivating pre-change claims (the 65-line vs 71-line always-on-guidance drift inventory).
