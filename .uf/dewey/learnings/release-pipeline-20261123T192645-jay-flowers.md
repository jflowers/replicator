---
tag: release-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-23T19:26:45Z
identity: release-pipeline-20261123T192645-jay-flowers
tier: draft
as_of: "replicator@c548bda681d556ebc1d3a59096063444a1b7b265"
---

Pinning an external reusable workflow to a full 40-char commit SHA (CI-001/CI-002) mitigates *integrity* risk: an attacker cannot swap the workflow content behind a pinned SHA without breaking the pin, so the 'force-push the personal repo' scenario does not change what executes. It does not mitigate *availability* (deleting or transferring the repo breaks the pipeline) or *governance visibility* (org-level review and branch protection do not apply to out-of-org code). When assessing a supply-chain finding on a delegated workflow, separate the SHA-mitigated integrity risk from the residual availability/governance risk rather than treating the SHA pin as addressing all of it.
