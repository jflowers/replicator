---
tag: ci-gating
author: jay-flowers
category: pattern
created_at: 2026-11-24T22:24:55Z
identity: ci-gating-20261124T222455-jay-flowers
tier: draft
as_of: "replicator@b0457f9b84ab0ec527be5c132f9ebe6a5d4b4e1b"
---

Structure-pinning regression tests for CI gating must pin the load-bearing tokens of any inline detection script, not just the job-graph wiring (needs/if/aggregate). The wiring determines whether jobs run; the script determines whether the gate fails open or closed. A gate whose decision logic silently regresses — a widened path prefix in a case pattern, a dropped `git diff --no-renames`, a renamed step id that empties the gating output (skipped counts as success for `if: always()` aggregates) — passes wiring-only tests. Pin the tokens with the same regex assertions used for the wiring, or extract the detection into a script whose decision paths can be fixture-tested offline (the replicate workflowtest package's ci_dependencies_test.go precedent: structure assertions plus decision-path fixtures). When a design document cites a regression test as the mitigation for a named risk, verify the test actually asserts the risk's mechanism — the path filter, not just the jobs that consume it.
