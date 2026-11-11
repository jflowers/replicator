---
tag: ci-configuration
author: jay-flowers
category: gotcha
created_at: 2026-11-11T14:37:54Z
identity: ci-configuration-20261111T143754-jay-flowers
tier: draft
---

Issue bodies can misstate CI configuration: replicator issue #162 claimed CI sets GOTOOLCHAIN=local, but a grep of .github/ finds no GOTOOLCHAIN anywhere — CI derives the toolchain from go.mod via actions/setup-go's `go-version-file: go.mod`. Verify CI-environment claims in issue text against the actual workflow files before planning a fix; here the correction changed nothing (setup-go installs the go.mod-declared version either way), but the same class of error could send a fix down the wrong path when it matters.
