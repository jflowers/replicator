---
tag: pin-tests
author: jay-flowers
category: gotcha
created_at: 2026-10-28T20:22:00Z
identity: pin-tests-20261028T202200-jay-flowers
tier: draft
---

Pin-test gotcha on negative pins: when a structural pin test asserts the ABSENCE of a phased-out phrase in markdown assets, a case-sensitive substring pin only catches the casing you happened to observe. Legacy instruction prose typically appears in two casings — sentence-initial capitalized ("Use the **question tool** with options …") and mid-sentence lowercase ("… use the **question tool** with options …") — and both occurred in the text removed by replicator PR #854 (uf.address-feedback 3.2 and uf.review-pr Step 3 were lowercase; uf.triage-issue 4.4 capitalized). Pin the case-agnostic core of the phrase (drop the leading verb: "**question tool** with options") or check both capitalizations, so a reintroduced gate in either legacy casing fails the negative pin even when it omits the backtick option-array style that the array-count tripwire would catch.
