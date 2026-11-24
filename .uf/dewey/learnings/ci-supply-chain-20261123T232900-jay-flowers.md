---
tag: ci-supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-23T23:29:00Z
identity: ci-supply-chain-20261123T232900-jay-flowers
tier: draft
as_of: "replicator@ce555cf869d9e801ac5994ca662905c4b424e868"
---

In repos that scaffold fullsend-managed GitHub workflows (files with a "managed by fullsend / do not edit" header, e.g. .github/workflows/fullsend.yaml and prioritize.yml), dependabot bumps the in-repo pinned SHAs, but the canonical source lives in fullsend's scaffold template (internal/scaffold/fullsend-repo/.github/workflows/). A future fullsend re-scaffold can overwrite the in-repo pin, and each scaffolded workflow is bumped independently, so reusable-dispatch.yml and reusable-prioritize.yml can drift to different fullsend versions. Treat in-repo pin bumps as transient; land durable version changes in the fullsend scaffold and align coupled workflow pins to one release, then smoke-test one prioritize run after merge.
