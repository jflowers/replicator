---
tag: spec-review
author: jay-flowers
category: gotcha
created_at: 2026-10-26T23:03:12Z
identity: spec-review-20261026T230312-jay-flowers
tier: draft
---

Recorded review learnings do not self-execute. When a review iteration stores a learning that flags a specific artifact defect, a later verification pass must re-open that artifact and confirm compliance directly, because the learning's existence does not guarantee the fix was applied. Observed in replicator openspec/changes/init-smart-sync: learning spec-quality-20261026T212938-jay-flowers flagged the spec.md "Machine-parseable result" scenario ("the CLI renders it with a distinct style") as contradicting design D3's warn-style reuse, yet the final spec.md still carries the un-reworded gloss. Final-verification checklists should include: re-read every learning filed against this change and confirm each named artifact now complies.
