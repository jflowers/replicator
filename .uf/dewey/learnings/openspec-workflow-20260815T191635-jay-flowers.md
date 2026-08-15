---
tag: openspec-workflow
author: jay-flowers
category: gotcha
created_at: 2026-08-15T19:16:35Z
identity: openspec-workflow-20260815T191635-jay-flowers
tier: draft
---

The OpenSpec workflow for content-only changes (no Go code, no tests) follows a streamlined path: propose creates all artifacts, spec review validates the design, implementation is direct file edits, and code review confirms tag placement and JSON validity. The key gotcha is that review markers (`<!-- spec-review: passed -->` and `<!-- code-review: passed -->`) must be written to tasks.md immediately after reviews pass -- if the session ends before markers are written, the unleash pipeline cannot detect completion on resume and must re-run reviews. Always write filesystem markers before proceeding to the next step.
