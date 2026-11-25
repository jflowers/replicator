---
tag: ci-security-gating
author: jay-flowers
category: gotcha
created_at: 2026-11-24T22:51:02Z
identity: ci-security-gating-20261124T225102-jay-flowers
tier: draft
as_of: "replicator@b0457f9b84ab0ec527be5c132f9ebe6a5d4b4e1b"
---

Gating a reusable CI workflow (e.g., complytime/org-infra reusable_ci.yml) at the workflow-call granularity skips everything bundled inside it, not just the expensive job the gating was meant to avoid. In replicator, gating `call-reusable-ci` on a `changes` output also removed security scanning: `.mega-linter.yml` enables REPOSITORY_BETTERLEAKS and REPOSITORY_KICS with no `.uf/` exclusion (only MARKDOWN_MARKDOWNLINT/YAML_YAMLLINT exclude `.uf/`), so learnings-only PRs merged without secret scanning. When adding a path- or content-based CI skip, enumerate every scanner/check bundled in the gated unit and confirm each is either acceptable to skip or split into its own always-on job.
