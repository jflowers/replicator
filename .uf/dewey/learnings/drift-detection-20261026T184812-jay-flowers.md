---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-10-26T18:48:12Z
identity: drift-detection-20261026T184812-jay-flowers
tier: draft
---

Cheap mirrored-file drift audit without running commands: read each file with the read tool using limit=1 and compare the hashline file_hash/REV values between the scaffolded copy (.opencode/...) and the embedded copy (internal/agentkit/content/...) — the hashline header hash is computed over the full file regardless of the line limit, so identical hashes prove byte-identity without loading full contents. This audited all 15 replicator mirrored pairs in seconds and found exactly one drifted pair (always-on-guidance: 65-line scaffolded copy vs 71-line hardened embedded copy), refuting a spec's claimed 6-file drift inventory.
