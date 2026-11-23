---
tag: pin-tests
author: jay-flowers
category: pattern
created_at: 2026-10-28T15:29:04Z
identity: pin-tests-20261028T152904-jay-flowers
tier: draft
---

Pin-test design for markdown command assets (scaffold_test.go, unbound-force): (a) pin wrap-safe contract clauses with RAW substring strings.Contains so a clause split across markdown lines fails the pin — each fragment MUST sit on a single line; (b) pin option inventories with whitespace-normalized strings.Contains over QUOTED labels — strings.Join(strings.Fields(s), " ") collapses line wrapping, and the closing quote anchors the match so a prefix label like "No -- skip" cannot false-positive against "No -- skip posting"; (c) pin gate counts with strings.Count of a marker that appears exactly once per gate (e.g. "Per-option reasoning:") so any added or removed gate fails until the pin is deliberately updated. Pattern: TestQuestionToolPresentationContractGates (fragments 8x8/6x8 across three assets, inventories 14/16/19 labels, counts 8/6/8), verified by single-occurrence-deletion mutation testing.
