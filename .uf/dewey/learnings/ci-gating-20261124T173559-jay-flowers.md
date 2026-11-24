---
tag: ci-gating
author: jay-flowers
category: gotcha
created_at: 2026-11-24T17:35:59Z
identity: ci-gating-20261124T173559-jay-flowers
tier: draft
as_of: "replicator@a3390914a2d4a8e6ca5a0a93ecfdb68452d59f62"
---

Structural regression tests for GitHub Actions workflow files must pin the safety-critical literal tokens of inline job scripts, not just the job-graph wiring (needs/if/aggregate names). Wiring assertions cannot detect a weakened detection script: dropping `git diff --no-renames`, replacing a fail-closed `if ! cmd` guard with `|| true`, or removing an event short-circuit leaves the job graph looking correct while the gate silently fails open, and no test fails. Assert the script tokens (flags, fail-closed guards, path filters, exit codes) with the same section-scoped regex matching already used for the wiring assertions. Replicator precedent: bump-org-infra-reviewer-pins requires its workflow regression test to match the pinned orgInfraSHA and forbids loosening the regex to pass.
