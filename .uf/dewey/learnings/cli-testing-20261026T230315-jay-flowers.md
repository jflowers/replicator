---
tag: cli-testing
author: jay-flowers
category: gotcha
created_at: 2026-10-26T23:03:15Z
identity: cli-testing-20261026T230315-jay-flowers
tier: draft
---

cobra's MarkDeprecated notice is emitted by the cobra execution layer, not by the command's RunE: tests that call the run function directly bypass flag parsing and never see the notice. Asserting the notice requires executing through a cobra root with the deprecated flag actually passed and output captured via root.SetOut(&buf), as in replicator's TestInitCmd_ForceDeprecationNotice; the behavioral half (flag is a no-op) must be covered by separate direct-run tests.
