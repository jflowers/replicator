---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-24T22:54:30Z
identity: ci-cd-20261124T225430-jay-flowers
tier: draft
as_of: "replicator@b0457f9b84ab0ec527be5c132f9ebe6a5d4b4e1b"
---

A GitHub Actions reusable workflow call (`uses:`) is all-or-nothing for job-level `if:` gating: skipping the calling job skips every check bundled inside the called workflow. When gating a reusable call for path-based CI skipping, enumerate what the call bundles (linters, secret scanners, PR-title/commitlint validation) and confirm each is either genuinely safe to skip for the skipped path set or re-homed to an ungated job. Otherwise a required aggregate check reports green while a real validation never ran.
