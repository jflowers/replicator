## Why

`replicator init` scaffolds five slash-command files (forge, org, inbox, forge-status, handoff) that each contain `<protect>` tags on line 5. These tags are designed to prevent DCP (the Diff Context Protocol plugin) from modifying protected sections during code generation. However, `replicator init` does not create the `.opencode/dcp.jsonc` configuration file that enables DCP's `protectTags` feature. Without this file, the `<protect>` tags are inert — DCP has no configuration telling it to honor them.

This was identified as a gap left by the `protect-tags-slash-commands` change, which added `<protect>` tags but placed the `protectTags` setting in `opencode.json` (wrong location — causes a config validation error). The correct location is `.opencode/dcp.jsonc`, which is a DCP plugin config file, not part of OpenCode's core configuration.

The replicator repo itself already has `.opencode/dcp.jsonc` committed (via `b147bc0`), but projects that run `replicator init` do not receive this file.

Related: unbound-force/unbound-force#502

## What Changes

Add a `ScaffoldDCP()` function to the `agentkit` package that creates `.opencode/dcp.jsonc` with `protectTags: true` during `replicator init`. The function uses idempotent merge logic (separate from the existing `Scaffold()` file walk) so it can handle existing DCP configurations gracefully.

## Capabilities

### New Capabilities
- `ScaffoldDCP`: Creates `.opencode/dcp.jsonc` with DCP schema reference and `protectTags: true` enabled. Idempotent — skips if already configured, merges if file exists but lacks `protectTags`.
- `checkDCPConfig` (doctor): Verifies that the current project has a valid `.opencode/dcp.jsonc` (or `.dcp.json`) with `protectTags: true` when command files with `<protect>` tags are present. Warns (not fails) when missing, since DCP is optional.

### Modified Capabilities
- `replicator init`: Now calls `ScaffoldDCP()` after `Scaffold()`, ensuring projects receive a complete agent kit with working protect-tag support.
- `replicator doctor`: Adds a 5th check (`dcp_config`) that verifies per-project DCP configuration health alongside the existing environment checks.

### Removed Capabilities
- None

## Impact

- **`internal/agentkit/agentkit.go`**: New `ScaffoldDCP()` function with idempotent merge logic
- **`internal/agentkit/agentkit_test.go`**: New tests for DCP scaffolding (fresh, skip, merge scenarios)
- **`cmd/replicator/init.go`**: Call `ScaffoldDCP()` after `Scaffold()` in `runInit()`
- **`cmd/replicator/init_test.go`**: Add `dcp.jsonc` to assertions in existing init tests
- **`internal/doctor/checks.go`**: New `checkDCPConfig()` function as 5th health check
- **`internal/doctor/checks_test.go`**: Tests for DCP config check (present, missing, missing protectTags)
- Existing scaffold file count (15) is unchanged — DCP is a separate function
- All projects initialized with `replicator init` will now have working `<protect>` tag support out of the box
- `replicator doctor` will warn users when DCP config is missing or incomplete

## Constitution Alignment

Assessed against the Replicator constitution (`.specify/memory/constitution.md`), which extends the Unbound Force org constitution v1.1.0.

### I. Autonomous Collaboration

**Assessment**: PASS

`ScaffoldDCP()` is a standalone function with a clear interface (`targetDir string`) returning `(ScaffoldResult, error)`. It operates independently from `Scaffold()` and produces self-describing results. The DCP config file it creates enables artifact-level protection — a form of autonomous collaboration where agents respect protected boundaries without runtime coupling.

### II. Composability First

**Assessment**: PASS

The function is independently callable — it does not require `Scaffold()` to run first. It handles both `.opencode/dcp.json` and `.opencode/dcp.jsonc` file extensions. The DCP config file itself is optional — projects function without it (protect tags are simply inert). No mandatory dependencies are introduced.

### III. Observable Quality

**Assessment**: PASS

`ScaffoldDCP()` returns a `ScaffoldResult` with `Path` and `Action` fields (created/skipped/updated), matching the existing pattern used by `Scaffold()`. The init command renders these results using the same styled output. The DCP config file uses the standard JSON schema reference for validation.

### IV. Testability

**Assessment**: PASS

All scenarios (fresh directory, existing config with protectTags, existing config without protectTags, `.dcp.json` alias) are testable using `t.TempDir()` with no external services. Tests verify file existence, content correctness, and idempotent behavior in isolation.
<!-- scaffolded by uf vdev -->
