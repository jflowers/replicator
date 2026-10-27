---
tag: spec-review
author: jay-flowers
category: gotcha
created_at: 2026-10-26T20:40:05Z
identity: spec-review-20261026T204005-jay-flowers
tier: draft
---

Supersession annotations scoped to a single requirement line leave sibling restatements of the old behavior contradictory. Speckit specs state one behavior in several places — FR requirements, user-story acceptance scenarios, and edge-case notes. In replicator's specs/003-rename-terminology/spec.md the init skip-without-force behavior appears at lines 38, 39, 78, and 100; a supersession task that annotates only FR-011 (line 100) leaves three unannotated contradictions with the new default. When writing or reviewing a supersession task, grep the old spec for every restatement of the superseded behavior and cover all locations.
