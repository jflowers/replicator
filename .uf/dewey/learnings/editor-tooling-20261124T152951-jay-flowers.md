---
tag: editor-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-24T15:29:51Z
identity: editor-tooling-20261124T152951-jay-flowers
tier: draft
as_of: "replicator@a3390914a2d4a8e6ca5a0a93ecfdb68452d59f62"
---

The edit tool can return "Could not find oldString in the file" for an edit whose content nonetheless lands on disk (observed editing .github/workflows/ci_checks.yml in the replicator repo, 2026-11-24): the failed-looking call had applied both the intended old→new replacement. Blindly retrying the same edit after that error would fail again (oldString genuinely gone) or risk duplicating content. Always re-read the file and diff/validate the on-disk state before retrying or assuming failure; verify with git diff, actionlint, and targeted tests rather than trusting the tool's error message alone.
