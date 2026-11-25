---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-24T23:01:37Z
identity: ci-cd-20261124T230137-jay-flowers
tier: draft
as_of: "replicator@b0457f9b84ab0ec527be5c132f9ebe6a5d4b4e1b"
---

In GitHub Actions pull_request workflows, a three-dot diff `git diff BASE...HEAD` using `github.event.pull_request.base.sha` can overcount: on an update-from-main ("Update branch"/synchronize) event, if base.sha lags the base tip merged into the head, `merge-base(base.sha, head)` collapses to the stale SHA and every change main gained since then is attributed to the PR (flipping a path-based skip gate to "run everything"). The robust fix is to diff the checked-out merge commit against its first parent: `git diff --name-only --no-renames HEAD^1 HEAD` (default pull_request checkout is refs/pull/N/merge, so HEAD^1 is the current base tip). This removes the base.sha-freshness dependency. Guard the merge-conflict case where refs/pull/N/merge is absent.
