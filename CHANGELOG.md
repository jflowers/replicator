# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/) and
this project adheres to [Semantic Versioning](https://semver.org/).

## Unreleased

### Fixed
- Homebrew cask published with checksums on the wrong stanza. The
  `publish-cask` job scanned forward from the `darwin_arm64` marker for the
  next `sha256` line, but the cask emits each stanza's `sha256` before its
  `url` — so darwin's checksum was written onto the `linux_amd64` line while
  darwin kept its stale pre-signing value. This broke `brew install` on both
  macOS arm64 and Linux amd64 in v0.5.0. A tested integrity script now
  requires exactly one darwin URL immediately preceded by its `sha256`,
  cross-checks the release manifest, fails closed on layout drift, and runs
  its regression fixtures in pull-request CI.
  (Fixes [#87](https://github.com/unbound-force/replicator/issues/87))
- TOCTOU gap between macOS code signing and cask publication. Extracted cask
  publishing from `sign-macos` into a dedicated `publish-cask` job that
  computes the SHA from the final release artifact. Shipped in v0.5.0; note
  that Homebrew installs remained broken until the stanza-targeting fix
  above.
  (Fixes [#81](https://github.com/unbound-force/replicator/issues/81))

### Added
- SECURITY.md with vulnerability reporting policy (GitHub Security
  Advisories preferred, email fallback)
- CODEOWNERS for governance file protection
- Dependabot configuration for automated weekly dependency update PRs
  for Go modules and GitHub Actions
- Standardized Dependabot dependency-review reporting and guarded approval
  for safe updates
- `govulncheck` security scanning in CI pipeline (pinned to v1.5.0)
- Per-package coverage ratchets in CI with 11 package thresholds
- `make coverage` and `make check-coverage` Makefile targets for
  local coverage enforcement before pushing
- CI convention pack (`ci.md`) for workflow standards
- CLAUDE.md for automatic convention pack loading in Claude Code
- New agentkit slash commands scaffolded by `replicator init`:
  `/forge`, `/forge:status`, `/org`, `/inbox`, `/handoff`
- New agentkit agents: `coordinator`, `worker`, `background-worker`
- New agentkit skills: `always-on-guidance`, `forge-coordination`,
  `forge-global`, `learning-systems`, `replicator-cli`,
  `system-design`, `testing-patterns`

### Changed
- Release pipeline now uses shared org-infra reusable workflows for
  preflight validation and GoReleaser execution. Releases gain supply
  chain artifacts (cosign signatures, SBOMs), smarter re-run resilience,
  and semver ordering validation. GoReleaser config gains
  `release.extra_files` for Homebrew cask upload. macOS code signing
  and notarization stays inline.
  (Part of unbound-force/unbound-force#428)
- Slash commands migrated to `uf.*` namespace (e.g., `/unleash` is
  now `/uf.unleash`, `/cobalt-crush` is now `/uf.cobalt-crush`)

### Security
- Go bumped to 1.25.12 for crypto/tls vulnerability fix
- All CI actions and reusable workflows pinned to commit SHAs for
  supply chain integrity (`actions/checkout`, `actions/setup-go`,
  `complytime/org-infra`, `govulncheck`)

## [0.2.0] - 2026-04-06

### Added
- 53 MCP tools across 4 categories: Hive (11), Swarm Mail (10),
  Swarm (24), Memory (8)
- Swarm orchestration with git worktree isolation
- Agent messaging with file reservations
- Dewey memory proxy with graceful degradation
- CLI commands: init, doctor, stats, query, setup
- Parity testing engine (100% shape match vs TypeScript)
- macOS code signing and notarization
- Homebrew distribution via `brew install unbound-force/tap/replicator`
- GoReleaser v2 release pipeline (darwin-arm64, linux-amd64, linux-arm64)
- Dewey MCP tool name update (dewey#28 prefix drop)
- `replicator init` command for per-repo setup
- Constitution and expanded AGENTS.md

### Changed
- Version command now displays commit hash and build date
- Makefile: added release, install targets

## [0.1.0] - 2026-04-04

### Added
- Initial release: Phase 0 scaffold
- MCP JSON-RPC server (stdio transport)
- SQLite database via `modernc.org/sqlite` (pure Go, no CGo)
- Tool registry framework
- 4 hive tools: `hive_cells`, `hive_create`, `hive_close`, `hive_update`
- CLI: `replicator serve`, `replicator cells`, `replicator version`
- 16 tests across 3 packages
- CI workflow (go vet + go test + go build)
- MIT LICENSE with Joel Hooks attribution

[0.2.0]: https://github.com/unbound-force/replicator/releases/tag/v0.2.0
[0.1.0]: https://github.com/unbound-force/replicator/releases/tag/v0.1.0
