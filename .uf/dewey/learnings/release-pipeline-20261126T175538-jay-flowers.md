---
tag: release-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-26T17:55:38Z
identity: release-pipeline-20261126T175538-jay-flowers
tier: draft
as_of: "replicator@c8203e0d68ce6de23e813fec1060c3ed1cc9bcca"
---

Post-release Homebrew cask checks wired through a reusable workflow's string input (e.g. jflowers/ship v0.10.0 tap_install_check, run as `bash -euo pipefail -c \"$CHECK\"` with only HOMEBREW_PREFIX guaranteed in the environment) can assert installability (binary runs, exit 0) but cannot reliably assert version identity, because the release tag is not exposed to the check's shell. Treat such a check as a liveness smoke test; version-match assertions require reading `brew info --cask <name>` version or passing the tag through the workflow's env, which must be added in the reusable workflow, not the consumer.
