---
tag: cli-ux
author: jay-flowers
category: pattern
created_at: 2026-10-26T23:03:10Z
identity: cli-ux-20261026T230310-jay-flowers
tier: draft
---

When deprecating a CLI flag to permanent no-op status (e.g., cobra MarkDeprecated), sweep the flag's entire presentation surface in the same change: (1) the flag's own help description, which still describes the old behavior and now misdescribes a no-op; (2) any options-struct field that plumbs the value into the command runner — either add a comment stating it is intentionally unread (retained so the ignore-contract stays enforceable at the function layer) or drop it and rely on command-level tests; (3) skill/README/AGENTS references to the flag. Observed in replicator openspec/changes/init-smart-sync: the --force description still read "Overwrite existing agent kit files" after the flag became a no-op, and initOptions.force became a write-only field — both cosmetic, both avoidable with a presentation-surface checklist.
