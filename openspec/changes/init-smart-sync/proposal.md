<!-- scaffolded by uf v0.17.0 -->

## Why

`replicator init` currently skips every existing agent kit file unless
`--force` is passed. This means re-running `init` after the embedded kit
content evolves (a new binary release) never updates stale scaffolded
files — drift accumulates silently between `.opencode/` and
`internal/agentkit/content/`. The replicator repo itself currently has
6 of 15 mirrored pairs drifted (missing `<protect>` tags in four
commands, missing delegation-contract updates in two agents), and CI
cannot catch it because drift-detection tests only cover 2 skills plus
1 structural check (issue #70).

The skip-by-default design was chosen to protect user-customized kit
files, but it makes the common case (stale kit, unmodified locally)
impossible without remembering an undocumented-in-practice `--force`
flag. Additionally, the locally installed binary predates embedded
content changes, so even `--force` writes stale bytes until the binary
is rebuilt.

## What Changes

- `agentkit.Scaffold` gains a smart-sync default: create missing
  files, skip byte-identical files, and update drifted files
  (embedded content is canonical).
- `ScaffoldResult.Action` gains an `"updated"` value for drifted
  files that were rewritten.
- New `--skip-existing` flag on `replicator init` restores the old
  skip-all-existing behavior for repos that intentionally customize
  kit files.
- `--force` becomes a deprecated no-op (smart-sync is now the
  default behavior); registered via `cobra` flag deprecation so
  existing scripts keep working.
- `ScaffoldDCP` and `cells.json` handling are unchanged (DCP config
  is already smart; `cells.json` is user state).
- Replicate repo dogfood: rebuild binary and run the new
  `replicator init` to sync the 6 drifted `.opencode/` files, proving
  the feature end-to-end.

## Capabilities

### New Capabilities

- `agentkit-scaffolding`: Smart-sync scaffolding of the embedded
  agent kit — create missing, skip identical, update drifted, with an
  opt-out skip-existing mode and a deprecated `--force` alias.

### Modified Capabilities

- None (no prior OpenSpec capability specs exist; `openspec/specs/`
  is empty).

### Removed Capabilities

- None.

## Impact

- `internal/agentkit/agentkit.go` — `Scaffold` signature change
  (bool force parameter replaced by a mode), byte-comparison logic.
- `internal/agentkit/agentkit_test.go` — new smart-sync unit tests.
- `cmd/replicator/init.go` — flag wiring, `--force` deprecation,
  `--skip-existing` flag, `"updated"` output styling, help text.
- `cmd/replicator/init_test.go` — rewritten skip/force tests, new
  smart-sync and flag tests.
- `.opencode/` (6 drifted files) — synced to embedded content via
  the new smart-sync run.
- `CHANGELOG.md`, `README.md` (if it documents skip behavior),
  embedded `replicator-cli` skill (documents `--force`).
- Website: user-facing `init` behavior change requires a
  `unbound-force/website` issue before merge.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Kit files are artifacts communicated between binary releases and
project directories; smart-sync keeps those artifacts faithful to the
embedded source of truth without runtime coupling. `ScaffoldResult`
remains self-describing JSON with an explicit per-file action.

### II. Composability First

**Assessment**: PASS

The binary remains fully standalone. No new dependencies are
introduced — smart-sync uses stdlib `bytes.Equal` and the existing
`embed.FS`. Dewey/website integrations remain optional.

### III. Observable Quality

**Assessment**: PASS

Every scaffolded file gets a machine-parseable action
(`created`/`skipped`/`updated`) in `ScaffoldResult` and styled CLI
output. The new `"updated"` action makes drift repair observable
where it was previously invisible.

### IV. Testability

**Assessment**: PASS

All new behavior is testable in isolation with `t.TempDir()` and
in-memory embedded content — no external services, no network, no
database. Existing test conventions (stdlib `testing`, no testify)
are followed.
