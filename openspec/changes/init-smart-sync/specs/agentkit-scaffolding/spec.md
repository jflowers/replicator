<!-- scaffolded by uf v0.17.0 -->

## ADDED Requirements

### Requirement: Smart-Sync Default Behavior

The `replicator init` command MUST scaffold each embedded agent kit
file according to its on-disk state: files that do not exist MUST be
created; files whose bytes are identical to the embedded content
MUST be skipped; files whose bytes differ from the embedded content
MUST be overwritten with the embedded content (embedded content is
canonical).

#### Scenario: Missing file is created

- **GIVEN** a target directory without `.opencode/commands/forge.md`
- **WHEN** `replicator init` runs
- **THEN** the file is created with embedded content and the result
  action is `created`

#### Scenario: Identical file is skipped

- **GIVEN** a target directory where `.opencode/commands/forge.md`
  exists and is byte-identical to the embedded content
- **WHEN** `replicator init` runs
- **THEN** the file is not rewritten and the result action is
  `skipped`

#### Scenario: Drifted file is updated

- **GIVEN** a target directory where `.opencode/commands/forge.md`
  exists with content that differs from the embedded content
- **WHEN** `replicator init` runs
- **THEN** the file is overwritten with the embedded content and the
  result action is `updated`

### Requirement: Skip-Existing Opt-Out

The `replicator init` command MUST provide a `--skip-existing` flag
that preserves all existing agent kit files regardless of content
drift, restoring the pre-smart-sync skip behavior for repositories
that intentionally customize kit files.

#### Scenario: Customized file preserved with flag

- **GIVEN** a target directory where `.opencode/commands/forge.md`
  exists with locally customized content
- **WHEN** `replicator init --skip-existing` runs
- **THEN** the file is not modified and the result action is
  `skipped`

#### Scenario: Missing files still created with flag

- **GIVEN** a target directory where `.opencode/commands/forge.md`
  exists with customized content and `.opencode/commands/org.md` does
  not exist
- **WHEN** `replicator init --skip-existing` runs
- **THEN** `forge.md` is untouched and `org.md` is created

### Requirement: Deprecated Force Flag

The `--force` flag MUST remain registered for backward compatibility
but MUST NOT change scaffold behavior (smart-sync is the default);
the flag MUST be marked deprecated via the CLI framework so users are
informed. In any flag combination, `--skip-existing` takes precedence
over the default smart-sync, and `--force` never overrides it.

#### Scenario: Force is a no-op

- **GIVEN** a target directory where an agent kit file exists with
  customized content
- **WHEN** `replicator init --force` runs
- **THEN** the file is updated to embedded content (same as the
  default smart-sync behavior) and a deprecation notice is emitted

#### Scenario: Combined flags — skip-existing wins

- **GIVEN** a target directory where an agent kit file exists with
  customized content
- **WHEN** `replicator init --force --skip-existing` runs
- **THEN** the file is preserved (the `--skip-existing` opt-out takes
  precedence) and the `--force` deprecation notice is still emitted

### Requirement: Updated Action Reporting

`ScaffoldResult.Action` MUST report `updated` for drifted files
rewritten by smart-sync, and CLI output MUST distinguish `updated`
from `created`, `skipped`, and `overwritten`.

#### Scenario: Machine-parseable result

- **GIVEN** a drifted agent kit file
- **WHEN** `replicator init` runs
- **THEN** the per-file result reports `"action": "updated"` and the
  CLI renders it with a distinct style

### Requirement: Unchanged Ancillary Behavior

The `cells.json` state file MUST NOT be overwritten when it exists,
and the DCP configuration scaffolding logic (create when missing,
skip when `protectTags` present, overwrite when absent) MUST remain
unchanged.

#### Scenario: cells.json preserved

- **GIVEN** a target directory with existing `.uf/replicator/cells.json`
  containing user state
- **WHEN** `replicator init` runs
- **THEN** `cells.json` content is unchanged
