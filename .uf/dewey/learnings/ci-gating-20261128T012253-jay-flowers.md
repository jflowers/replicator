---
tag: ci-gating
author: jay-flowers
category: pattern
created_at: 2026-11-28T01:22:53Z
identity: ci-gating-20261128T012253-jay-flowers
tier: draft
as_of: "replicator@f5baa677d5075e11b904d5dcc14fc18b686fea22"
---

When an aggregate CI verdict validates that each job "ended as the classifier chose" (success when chosen, skipped when not), execution-testing the verdict must cover both directions for every gated job: a chosen job failing AND a non-chosen job succeeding (or a chosen job skipping). If the verdict only fails on failure/cancelled results, a classifier regression that misroutes a PR (e.g., returns code='false' for a code PR) skips the build and still passes the aggregate — the gate launders the misclassification into a green required check. replicator's TestCIGatingWorkflows_AggregateVerdictBehavior pins this with an 18-case truth table over scope/code/build/lint/gitleaks, including "gitleaks succeeded when it should have been skipped" and "build skipped when it should have run" — the anti-laundering cases.
