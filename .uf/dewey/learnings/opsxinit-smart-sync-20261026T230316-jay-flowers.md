---
tag: opsxinit-smart-sync
author: jay-flowers
category: gotcha
created_at: 2026-10-26T23:03:16Z
identity: opsxinit-smart-sync-20261026T230316-jay-flowers
tier: draft
---

replicator_forge_worktree_merge's project_path parameter must be the MAIN repository checkout path, not the worker worktree path (observed 2026-10-26 on replicator opsx/init-smart-sync Phase 3 batch 1, task 3.1 merge). The implementation (internal/forge/worktree.go → gitutil.CherryPick) lists commits on the swarm/<task> branch since the start commit and cherry-picks them onto project_path's CURRENT branch; passing the worktree path makes it cherry-pick the commit onto itself and fail with 'The previous cherry-pick is now empty'. Recovery: git -C .worktrees/<id> cherry-pick --abort (the worktree holds the sequencer state, not the main repo), then re-call the merge with the main repo path — the retry merges cleanly.
