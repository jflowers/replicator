---
tag: ci-gating
author: jay-flowers
category: gotcha
created_at: 2026-11-25T01:14:02Z
identity: ci-gating-20261125T011402-jay-flowers
tier: draft
as_of: "replicator@051e352f6a2e89c06fc17d86cfe94f22e62e9bc0"
---

A structure-pinning regression test for CI gating must pin the decision-propagation statements, not just the guards and filters: the step id paired with the job's outputs mapping, the decision variable's initialization and set-true branch, the final GITHUB_OUTPUT write, and the boolean operator joining failure-propagation contains() checks (pin the combined condition as one unit, not independent fragments). Omitting the output-propagation pins creates a self-concealing fail-open: a PR that deletes or disconnects the output write touches the workflow file, so the detection classifies it as code — but the emptied output skips the very build-and-test job that runs the regression test, and every check including the always() aggregate stays green while the gate is permanently open for all subsequent PRs. Deeper alternative when the inline script has real decision logic: extract it and fixture-test its decision paths offline (replicator's TestCIDependenciesWorkflow_ApprovalScriptEvaluatesPolicyFixtures precedent).
