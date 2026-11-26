---
tag: release-pipeline
author: jay-flowers
category: reference
created_at: 2026-11-26T16:31:08Z
identity: release-pipeline-20261126T163108-jay-flowers
tier: draft
as_of: "replicator@52657820a642302abd53e18a04e45ba918350b16"
---

In replicator, `replicator --version` is a valid release smoke-test command: the root cobra command sets `Version: Version` (cmd/replicator/main.go), so cobra auto-registers a `--version` flag on the root that prints and exits 0 (alongside the explicit `version` subcommand). When specifying ship's `tap_install_check`, `\"$HOMEBREW_PREFIX/bin/replicator\" --version` is safe because ship runs the check with `HOMEBREW_PREFIX=$(brew --prefix) bash -euo pipefail -c \"$CHECK\"`, so the variable is in the environment and a non-zero exit fails the check. Because the job runs after publish-cask, a failing check makes an already-published release's run red — which is intended post-release verification, not a rollback.
