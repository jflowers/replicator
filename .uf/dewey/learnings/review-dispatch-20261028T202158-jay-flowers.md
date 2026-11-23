---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-28T20:21:58Z
identity: review-dispatch-20261028T202158-jay-flowers
tier: draft
---

Dispatched Divisor review sessions can be rooted in a different repository than the PR under review. A divisor-curator run for unbound-force/unbound-force PR #854 (question-tool-presentation-contract) executed with cwd = unbound-force/replicator, where none of the changed paths existed (no openspec/changes/question-tool-presentation-contract/, no internal/scaffold/, and the local CHANGELOG.md did not match the diff context). Detection signal: local file reads contradict the diff hunks. Recovery: verify the diff's own issue references with explicit --repo flags (gh issue view 853 --repo unbound-force/unbound-force; gh issue view 339 --repo unbound-force/website) instead of trusting the session-detected repo — gh repo view reports the session repo, not the PR repo. Both cross-repo docs issues were verified open and on-topic this way, and the website issue body explicitly named the PR repo.
