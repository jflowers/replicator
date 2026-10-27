---
tag: scaffolding
author: jay-flowers
category: pattern
created_at: 2026-10-26T18:48:19Z
identity: scaffolding-20261026T184819-jay-flowers
tier: draft
---

When two scaffolding tools can write the same `.opencode/` path (Unbound Force CLI scaffolding vs `replicator init`'s embedded kit), an embedded-canonical smart-sync policy creates a clobber ping-pong: each `uf` update rewrites the path with newer uf-generation content, each subsequent `replicator init` overwrites it back with embedded content, and byte-parity tests flap on every version bump. Before declaring one tool's copy canonical by default overwrite, verify per-path ownership. The tell: a drifted pair whose on-disk copy contains tool-scaffolded content newer than the embedded copy (observed for `.opencode/skills/always-on-guidance/SKILL.md` at the uf v0.21.0 base commit during the init-smart-sync review, 2026-10-26).
