---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-28T00:50:43Z
identity: ci-cd-20261128T005043-jay-flowers
tier: draft
as_of: "replicator@1417eab74bb6c10984026580a858fc0c2537cf1f"
---

When a CI classifier job is delegated to an external reusable workflow, the in-repo push short-circuit (and its behavioral test) disappears. Verify the external classifier returns its 'run everything' value on push-to-main, because the release preflight requires specific check names (e.g. 'Build and Test', 'Standardized CI / Run linters') that only run when the classifier says so; a non-'true' result on push fails the aggregate and blocks releases.
