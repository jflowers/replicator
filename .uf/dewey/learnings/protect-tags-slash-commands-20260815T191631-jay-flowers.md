---
tag: protect-tags-slash-commands
author: jay-flowers
category: pattern
created_at: 2026-08-15T19:16:31Z
identity: protect-tags-slash-commands-20260815T191631-jay-flowers
tier: draft
---

When implementing protect-tags-slash-commands (issue #74), the change was purely additive markup -- only `<protect>` and `</protect>` tags added to 5 embedded markdown files plus a single `"protectTags": true` line in opencode.json. No Go code, no tests, no MCP protocol changes. The key design decision was section-level granularity (one protect block per logical section, not per line), with a clear taxonomy: protect guardrails/MUST rules, workflows with numbered steps, and session procedures; leave explanatory/advisory content (examples, interpretations, usage patterns) unprotected so DCP can still compress them. The 5 reviewers in spec review and 3 in code review all returned APPROVE with zero blocking findings. The recurring review feedback was that the exact JSON path for `protectTags` in opencode.json should be documented with a concrete snippet -- this was addressed at implementation time by placing it at the top level after `$schema`. For content-only changes like this, the verification tasks (tag count grep, JSON validation, make build, git diff --stat for constitution alignment) are lightweight and effective.
