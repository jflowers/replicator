---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-10-26T20:40:02Z
identity: spec-authoring-20261026T204002-jay-flowers
tier: draft
---

When a spec's tasks.md delegates GitHub issue filing to an implementation task (e.g., replicator's Website Documentation Gate filing into unbound-force/website), the task should encode the duplicate-issue search step — the exact `gh issue list --repo <target> --state open --search "<keyword>"` command — and name adjacent open issues to cross-link. Review-time duplicate verification cannot guarantee the issue-tracker state at implementation time, and an implementer executing the task verbatim files without searching. In the opsx/init-smart-sync spec review (2026-10-26), task 3.3 omitted the search; live checking confirmed no duplicate existed but surfaced adjacent website issue #231 (replicator scaffold docs) that the new issue should reference.
