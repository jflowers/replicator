<!-- scaffolded by uf v0.17.0 -->

## Context

`agentkit.Scaffold(targetDir string, force bool)` walks the embedded
`content/` FS and writes each file to `.opencode/`. Today, any
existing destination file is skipped unless `force` is true. The
embedded kit evolves with the binary, but re-running `init` never
repairs stale copies, so drift accumulates silently (currently 1 of
15 mirrored pairs is drifted in this repo — the always-on-guidance
skill, which `TestSkillFiles_DriftDetection` already flags at HEAD).
Constitution
alignment is documented in `proposal.md` (all four principles PASS);
the decisions below preserve that alignment.

## Goals / Non-Goals

### Goals

- Re-running `replicator init` converges `.opencode/` kit files to
  the embedded content (create missing / skip identical / update
  drifted).
- Preserve an escape hatch (`--skip-existing`) for repos that
  intentionally customize kit files.
- Keep `--force` accepted (deprecated no-op) so existing scripts do
  not break.
- Make drift repair observable via a new `"updated"` action.
- Dogfood: use the new behavior to fix this repo's own 1-file drift
  (always-on-guidance) plus the `replicator-cli` mirror updated by
  task 3.2.

### Non-Goals

- Changing `ScaffoldDCP` logic (already smart: create / skip-if-
  `protectTags` / overwrite).
- File-set convergence — smart-sync never deletes `.opencode/` kit
  files that are absent from the embedded content (orphaned
  scaffolded files are left in place); convergence is content-only
  for the paths the embedded kit contains.
- Changing `cells.json` handling (user state, always skipped).
- Provenance tracking of who modified a kit file (no hashes-of-old-
  versions registry).
- Backup files (`.bak`) for overwritten kit files.
- Expanding drift-detection test coverage to all 15 mirrored pairs
  (tracked separately as issue #70).
- Automated coordination with UF CLI scaffolding over the shared
  `.opencode/` paths — the ownership boundary is documented as a
  design decision (D8), not enforced.
- MCP tool surface changes — `init` is CLI-only, so no TypeScript
  parity fixtures are affected.

## Decisions

- **D1 — Smart-sync default, embedded canonical**: `Scaffold`
  compares on-disk bytes to embedded bytes with `bytes.Equal`.
  Identical → skip; different → overwrite with embedded content.
  Rationale: the mirror rule in AGENTS.md already declares the two
  copies MUST be byte-identical; treating embedded as canonical
  matches that model. (Observable Quality: the outcome is reported
  per file.)
- **D2 — Mode parameter replaces force bool**: `Scaffold(targetDir
  string, mode ScaffoldMode)` with `ScaffoldModeSync` (default) and
  `ScaffoldModeSkipExisting`. A two-value mode is clearer than a
  bool and leaves room for future modes without another breaking
  signature change. (Composability First: stdlib-only, no new
  dependencies.)
- **D3 — New action `"updated"`**: `ScaffoldResult.Action` reports
  `created` / `skipped` / `updated`. The old `"overwritten"` value
  disappears from `Scaffold` because force-mode is gone; the CLI
  renders `updated` with the existing warn (yellow) style that
  `overwritten` used.
- **D4 — `--force` deprecated, not removed**: registered via
  `cmd.Flags().BoolVar` + `MarkDeprecated("force", ...)` so cobra
  prints a deprecation notice when used. Behavior equals the new
  default. Avoids breaking CI scripts and muscle memory while the
  ecosystem migrates.
- **D5 — `--skip-existing` opt-out**: restores pre-smart-sync
  behavior for customized repos; maps to `ScaffoldModeSkipExisting`.
- **D6 — Unchanged ancillary paths**: `ScaffoldDCP` and `cells.json`
  logic untouched (see Non-Goals).
- **D7 — Dogfood remediation as a task**: after implementation, a
  task rebuilds the binary (`make install`) and runs the new
  `replicator init` at the repo root to sync the 1 pre-existing
  drifted file (always-on-guidance) plus the `replicator-cli` mirror
  (drifted by task 3.2's embedded-content update), then re-runs
  `make check`. This validates the feature on the exact drift that
  motivated it.
- **D8 — Path-ownership boundary with UF CLI scaffolding**:
  `replicator init` writes only the 15 mirrored kit paths backed by
  `internal/agentkit/content/`; the embedded side is
  replicator-only, but the `.opencode/` side of those pairs is
  shared — the UF CLI also scaffolds it and may regenerate copies
  on uf version bumps (the current always-on-guidance drift is
  exactly such a newer UF-scaffold generation). Embedded replicator
  content stays canonical for these paths; newer UF-scaffold content
  on a shared path is preserved by reconciling it into the embedded
  copy (task 3.4's merge pattern) before syncing, never by silent
  overwrite. `--skip-existing` is the opt-out for repos that want
  uf-managed content to win. Automated coordination between the two
  tools is a Non-Goal.

## Risks / Trade-offs

- **Clobbering local customizations by default** (the reason
  skip-by-default existed): accepted. Kit files are declared managed
  mirrors of embedded content; customizing them was already a
  drift. Mitigation: `--skip-existing` opt-out, `updated` output
  makes every overwrite visible, and the risk is documented in help
  text. Spec 003's skip-without-force scenario is intentionally
  superseded by this change.
- **Silent repair could surprise in CI contexts**: `init` output
  lists every `updated` file, so the repair is observable, never
  silent.
- **Signature change ripples**: `Scaffold` is called from
  `cmd/replicator/init.go` and tests only; the ripple is contained
  and covered by TDD.
- **Stale-binary confusion**: smart-sync repairs drift only relative
  to the embedded content in the *installed* binary. Documented in
  help text; dogfood task demonstrates the rebuild-then-init flow.
