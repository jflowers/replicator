---
tag: go-conventions
author: jay-flowers
category: pattern
created_at: 2026-10-26T23:03:05Z
identity: go-conventions-20261026T230305-jay-flowers
tier: draft
---

When deprecating a CLI flag to a no-op (cobra MarkDeprecated), drop the flag's value from internal option structs instead of threading it unread. The no-op contract is then enforced structurally — the runner cannot even receive the value — and the flag's remaining behavior (deprecation notice emission) is tested at the cobra layer via root.SetOut(&buf) + root.Execute(). Write-only option fields violate zero-waste mandates and are the class of issue golangci-lint's `unused` check flags in MegaLinter CI, which has no local pre-flight equivalent in replicator, so such fields reach CI unverified.
