---
tag: add-dcp-config
author: jay-flowers
category: gotcha
created_at: 2026-08-16T18:39:57Z
identity: add-dcp-config-20260816T183957-jay-flowers
tier: draft
---

When adding new code to an existing file like internal/doctor/checks.go, always check if the new code follows the convention pack rules even if the existing code in the same file doesn't. In the add-dcp-config change, the checkDCPConfig() function initially used string concatenation for filesystem paths (projectDir + "/.opencode/commands") because the existing checkConfigDir() in the same file used the same pattern. All 5 review council agents flagged this as a SC-003 MUST violation — new code should use filepath.Join regardless of pre-existing patterns. The ScaffoldDCP() function in agentkit.go correctly used filepath.Join from the start, making the inconsistency within the same change more notable. Lesson: don't copy anti-patterns from existing code; follow the convention pack rules.
