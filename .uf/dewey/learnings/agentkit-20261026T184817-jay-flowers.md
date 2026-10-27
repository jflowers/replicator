---
tag: agentkit
author: jay-flowers
category: gotcha
created_at: 2026-10-26T18:48:17Z
identity: agentkit-20261026T184817-jay-flowers
tier: draft
---

When reviewing or writing replicator OpenSpec proposals that cite mirrored-kit drift (`.opencode/` vs `internal/agentkit/content/`), re-verify the claim against the repo's embedded content at review time. Two failure modes produce phantom counts: (1) auditing with the locally installed binary, whose embedded content predates repo changes; (2) reusing an audit taken before a later sync commit. In the init-smart-sync review (2026-10-26) the proposal claimed 6/15 drifted pairs (four commands missing <protect> tags, two agents missing delegation-contract updates) while content-hash comparison showed 14/15 byte-identical; the single real drift (always-on-guidance) was in a pair already covered by TestSkillFiles_DriftDetection, so "CI cannot catch it" was false and the parity test was failing on the committed tree. Fast verification method: read both copies of each pair with limit=1 and compare the read tool's per-file content hash (validate the method against a known-identical CI-enforced pair such as forge-global).
