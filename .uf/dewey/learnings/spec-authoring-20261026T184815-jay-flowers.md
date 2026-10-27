---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-10-26T18:48:15Z
identity: spec-authoring-20261026T184815-jay-flowers
tier: draft
---

When authoring an OpenSpec/Speckit change proposal in replicator, derive the documentation-impact inventory by searching AGENTS.md and specs/ for the workflows the change modifies, rather than by recalling standard doc surfaces. The opsx/init-smart-sync proposal listed CHANGELOG.md, README.md, the embedded replicator-cli skill, and the website-gate issue, but omitted two surfaces its own text touches: AGENTS.md's 'Drift Detection (Mirrored Files)' section (whose 'manual verification during review is the only safeguard' guidance the smart-sync default supersedes) and specs/003-rename-terminology/spec.md FR-011, a live MUST requirement ('replicator init MUST NOT overwrite existing agent kit files unless --force') that the change reverses with no annotation task. Rule of thumb: when a design.md says a prior spec's scenario is 'superseded', add a task to annotate the superseded requirement (or record an explicit decision that historical specs stay frozen point-in-time records), and any AGENTS.md section naming the changed command belongs in the Impact list.
