---
tag: ci-pinning
author: jay-flowers
category: pattern
created_at: 2026-11-23T23:30:39Z
identity: ci-pinning-20261123T233039-jay-flowers
tier: draft
as_of: "replicator@ce555cf869d9e801ac5994ca662905c4b424e868"
---

For Dependabot PRs that bump a pinned third-party GitHub reusable workflow SHA, the SRE review is verification-only: (1) the new 40-char SHA must equal the commit an annotated release tag dereferences to (tag object target, not the tag SHA itself), (2) the version comment must name that tag, (3) the pinned workflow path must exist at the commit, and (4) `with:` inputs in the consumer workflow must still match the upstream input contract. Points (1)-(3) are checkable at review time; (4) only fails at run time because reusable-workflow input validation is evaluated when the workflow executes, not when CI parses the changed file — so a green PR check suite does not prove input compatibility. The compensating controls for that residual gap are the dependency-review release-age gate and post-merge observation of the next dispatch run, not additional PR-time checks.
