---
tag: ci-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-28T01:22:50Z
identity: ci-pipeline-20261128T012250-jay-flowers
tier: draft
as_of: "replicator@f5baa677d5075e11b904d5dcc14fc18b686fea22"
---

A CI job gated to a rare path (e.g. only `code == 'false'` chore/learnings PRs) is never executed by the PR that introduces it, so its pinned artifact URL, checksum, and scan command are unvalidated at merge time. When adding such a job, exercise it out-of-band (a scratch branch, a temporary unconditional run, or a local invocation of the same download+checksum+scan) before relying on it to gate a required check; otherwise the first real trigger is also the first failure, and it blocks the exact PR class the job was meant to enable.
