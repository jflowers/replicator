---
tag: documentation-completeness
author: jay-flowers
category: pattern
created_at: 2026-10-26T21:29:41Z
identity: documentation-completeness-20261026T212941-jay-flowers
tier: draft
---

When a spec change supersedes a previously documented behavior (flag semantics, default behavior, contract), the supersession sweep must cover every artifact that restates it, not just the primary spec file: normative FRs, acceptance scenarios, edge-case entries, quickstart/migration summaries in the same spec directory, and command summary tables (AGENTS.md CLI tables, skill Commands tables). In the replicator init-smart-sync review (2026-10-26), task 3.6 correctly enumerated all four restatements inside specs/003-rename-terminology/spec.md (FR-011 line 100, User Story 2 scenarios lines 38-39, Edge Cases line 78) but missed the quickstart.md:14 "New flag: replicator init --force" row in the same directory, and task 3.5 missed the AGENTS.md CLI Commands table row for `replicator init`. Pattern: grep the whole repo for the superseded flag/behavior string (e.g. `init --force`) across all .md files, then classify each hit as normative (annotate), reader-facing summary (annotate or update), or historical record (leave, with recorded decision) — and enumerate the chosen action per location in a doc-task bullet.
