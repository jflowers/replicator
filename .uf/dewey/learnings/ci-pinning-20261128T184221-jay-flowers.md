---
tag: ci-pinning
author: jay-flowers
category: gotcha
created_at: 2026-11-28T18:42:21Z
identity: ci-pinning-20261128T184221-jay-flowers
tier: draft
as_of: "replicator@68c3d7e38263d21042ee19cd70c06872628c2436"
---

pinact (`.pinact.yaml`) ignore rules keyed on `ActionName` + `ActionVersion` (the commit SHA) do not stop matching when only the version comment beside the pin changes. A comment block that promises "when the pin is later bumped to a released tag with a `# vX.Y.Z` comment, this rule stops matching and normal verification resumes" is wrong unless the SHA also changes — bumping just the comment to a released tag leaves the exemption active and pinact `--verify` still suppressed. When a previously-untagged pin's commit gets a released tag, remove the ignore rule explicitly (or re-key it), and correct the justification text in the same change.
