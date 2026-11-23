---
tag: scaffold-pin-tests
author: jay-flowers
category: pattern
created_at: 2026-10-28T20:21:53Z
identity: scaffold-pin-tests-20261028T202153-jay-flowers
tier: draft
---

Pin-testing prompt-asset contracts (verified on the question-tool-presentation-contract change, PR #854 in replicator): when a normative prose contract is duplicated across mirrored markdown command assets, mechanical drift protection works as a five-layer pin test over the embedded copies — (1) assert shared contract fragments PRESENT RAW so each fragment must sit on a single line (wrap-safe: a fragment split across wrapped lines fails), (2) count each fragment per asset and pin the count to the gate count so a removed or added gate fails, (3) count the house-style option-array opener (backtick + `["`) to trip new gates, (4) assert the legacy phrasing is ABSENT both raw and whitespace-normalized so line-wrapped regressions fail, and (5) pin per-label occurrence totals in the whitespace-normalized asset using the quoted-label form (`"No -- skip"`) as a closing-quote anchor so prefix labels cannot collide with longer labels (`"No -- skip posting"`). Because the existing byte-parity drift test maps embedded assets to canonical files, pinning the embedded copies pins both. Known boundary (must be documented in the spec, as FR-010's last scenario does): a gate added in neither the house style nor the pinned legacy phrasing, or a non-inventory label inside an existing array, is outside mechanical detection and relies on review.
