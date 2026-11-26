---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-25T20:56:49Z
identity: ci-cd-20261125T205649-jay-flowers
tier: draft
as_of: "replicator@3e10327f3f0208c60ad3ad0571bb26e0a0542fa0"
---

A Dependabot bump of a reusable workflow SHA consumed by a `pull_request_target` workflow cannot be validated by the bumping PR's own CI: `pull_request_target` runs the base-branch copy of the workflow, so CI exercises the OLD pinned version. Treat such bumps as post-merge-validated and plan a smoke check (watch the first real dispatch, keep a revert-the-pin recovery path) instead of relying on green CI as proof the new version works.
