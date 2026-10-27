<!-- scaffolded by uf v0.17.0 -->

## Why

`replicator init` currently skips every existing agent kit file unless
`--force` is passed. This means re-running `init` after the embedded kit
content evolves (a new binary release) never updates stale scaffolded
files — drift accumulates silently between `.opencode/` and
`internal/agentkit/content/`. The replicator repo itself currently has
1 of 15 mirrored pairs drifted: `.opencode/skills/always-on-guidance/SKILL.md`
(65-line pre-hardening on-disk copy vs the 71-line hardened embedded copy).
That pair is covered by `TestSkillFiles_DriftDetection`, so CI already
catches it (the test fails on the current tree); the wider gap is that
12 of 15 pairs lack drift coverage entirely (issue #70).

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
  `replicator init` to sync the 1 drifted `.opencode/` file plus the
  updated `replicator-cli` skill mirror, proving the feature
  end-to-end.

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
- `.opencode/` (1 drifted file, always-on-guidance, plus the
  `replicator-cli` mirror after task 3.2 updates its embedded copy) —
  synced to embedded content via the new smart-sync run.
- `CHANGELOG.md`, `README.md` (if it documents skip behavior),
  embedded `replicator-cli` skill (documents `--force`), `AGENTS.md`
  (Drift Detection section: smart-sync repair path, overwrite-by-default
  warning, updated review guidance).
- Website: user-facing `init` behavior change requires two
  `unbound-force/website` issues before merge — a documentation
  issue per the Website Documentation Gate and a blog-labeled
  issue covering the migration story for existing users
  (default flip, `--skip-existing` opt-out, preserving
  intentional customizations; precedent issue #51).

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

Coverage strategy: the new smart-sync branches (updated /
skip-existing / identical-skip) are exercised by the new unit tests
in `internal/agentkit` and `cmd/replicator`; the existing per-package
coverage ratchets (including `internal/agentkit` at 80%) apply
unchanged, and tasks 4.1/4.3 enforce the full CI-equivalent check
set (`go vet`, `govulncheck`, `go test ./... -count=1 -race`,
coverage ratchets, build) derived via the pre-flight skill.
