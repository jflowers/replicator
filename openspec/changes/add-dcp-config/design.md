## Context

`replicator init` scaffolds 15 agent kit files into `.opencode/` (5 commands, 7 skills, 3 agents). All five command files contain `<protect>` tags that mark execution-critical sections (guardrails, checklists, mandatory gates) for DCP preservation during context pruning. However, the DCP configuration file (`.opencode/dcp.jsonc`) that enables `protectTags: true` is not scaffolded, leaving the protect tags inert.

The replicator repo itself has `.opencode/dcp.jsonc` committed, but user projects initialized via `replicator init` do not receive it.

## Goals / Non-Goals

### Goals
- Scaffold `.opencode/dcp.jsonc` with `protectTags: true` during `replicator init`
- Handle idempotent behavior: skip if already configured, merge if file exists but lacks `protectTags`
- Handle `.opencode/dcp.json` (without `c`) as an alias
- Maintain separation from the existing `Scaffold()` file walk
- Add a `dcp_config` health check to `replicator doctor` that warns when DCP config is missing or incomplete

### Non-Goals
- Full JSONC parsing (comment stripping, AST manipulation) — too complex for the value
- Modifying `opencode.json` — `protectTags` belongs in DCP plugin config, not core config
- Changing the existing 15-file scaffold count — DCP is a separate function
- Supporting custom DCP config beyond `protectTags: true`

## Decisions

### D1: Separate `ScaffoldDCP()` function (not embedded in file walk)

The existing `Scaffold()` function walks an embedded filesystem and writes files with simple create/skip/overwrite logic. DCP config requires merge semantics (detect existing `protectTags`, preserve user customizations). A separate function keeps these concerns cleanly separated.

This aligns with **Composability First** — `ScaffoldDCP()` is independently callable without requiring `Scaffold()` to run first.

### D2: String scan for `protectTags` detection

Use `strings.Contains(content, `"protectTags"`)` to detect whether an existing DCP config already has protectTags configured, rather than parsing JSONC. JSONC parsing would require stripping comments before JSON unmarshaling, adding complexity for a simple boolean check.

Trade-off: If a user has `protectTags` in a comment but not as an actual setting, we'd incorrectly skip. This is an acceptable false-positive — the user clearly knows about the setting.

### D3: Prefer `.jsonc` extension, check both

Check for `.opencode/dcp.jsonc` first, then `.opencode/dcp.json`. If neither exists, create `.opencode/dcp.jsonc`. If `.dcp.json` exists (without `c`), operate on that file to respect the user's choice.

### D4: Return single `ScaffoldResult`

`ScaffoldDCP()` returns `(ScaffoldResult, error)` with the same `Path`/`Action` shape used by `Scaffold()`. Actions: "created" (fresh), "skipped" (already has protectTags), "updated" (merged protectTags into existing file). This aligns with **Observable Quality** — consistent, machine-parseable output.

### D5: DCP config content matches replicator's own `.opencode/dcp.jsonc`

The scaffolded file will contain:
```jsonc
{
  "$schema": "https://raw.githubusercontent.com/Opencode-DCP/opencode-dynamic-context-pruning/master/dcp.schema.json",
  // Enable <protect> tag preservation during DCP compression.
  // Slash command files in .opencode/commands/ use <protect> tags
  // to mark execution-critical sections (guardrails, checklists,
  // mandatory gates) that must survive context pruning.
  "compress": {
    "protectTags": true
  }
}
```

### D6: Doctor check uses `warn` status (not `fail`)

DCP is optional — projects function without it (protect tags are simply inert). Like the Dewey check, the DCP config check returns `warn` when config is missing or incomplete, not `fail`. This follows the established pattern: only environment essentials (git, database, config dir) cause failures.

### D7: Doctor check scans for `<protect>` tags in `.opencode/commands/`

Rather than unconditionally warning about missing DCP config, the check first looks for `.opencode/commands/*.md` files containing `<protect>`. If no protect-tagged commands exist, the check returns `pass` with "no protect-tagged commands found" — the DCP config is not needed. This avoids false warnings for projects that don't use protect tags.

### D8: Doctor check reuses string scan pattern from D2

The doctor check uses the same `strings.Contains` approach as `ScaffoldDCP()` for detecting `protectTags` in the DCP config file. Consistency across the codebase.

### D9: Doctor `checkDCPConfig()` accepts explicit directory parameter

`checkDCPConfig(projectDir string) CheckResult` takes an explicit directory path rather than using `os.Getwd()`. This enables isolated testing with `t.TempDir()` without requiring `os.Chdir()` (which is not thread-safe). The `Run()` function obtains the working directory via `os.Getwd()` once and passes it to `checkDCPConfig()`. This follows the pattern of `checkDatabase(store)` and `checkDewey(deweyURL)` — checks receive their dependencies as parameters. The `Run()` signature changes to `Run(store *db.Store, cfg *config.Config, projectDir string)`.

### D10: Update strategy replaces file content entirely

When `ScaffoldDCP()` encounters an existing config file without `protectTags`, the "updated" action replaces the entire file with the canonical DCP config content (from D5). This is simpler and safer than attempting to merge into an arbitrary JSONC structure, which could produce invalid JSON. The trade-off is that user customizations beyond `protectTags` are lost — but DCP configs are typically simple, and the non-goal of "supporting custom DCP config beyond `protectTags: true`" makes this acceptable.

### D11: Both files exist — prefer `.jsonc`

When both `.opencode/dcp.jsonc` and `.opencode/dcp.json` exist simultaneously, `ScaffoldDCP()` and `checkDCPConfig()` operate on `.opencode/dcp.jsonc` (the preferred extension) and ignore `.opencode/dcp.json`. This is consistent with D3's "check `.jsonc` first" rule.

## Risks / Trade-offs

- **String-based detection is imprecise**: `strings.Contains` could match `protectTags` in comments or non-standard locations. Accepted — the false-positive rate is negligible and avoids JSONC parsing complexity.
- **Full replacement on update**: When updating an existing config without `protectTags`, the entire file is replaced with the canonical content (D10). User customizations beyond `protectTags` are lost. This is safer than attempting partial JSONC merges that could produce invalid JSON.
- **File extension ambiguity**: Supporting both `.json` and `.jsonc` adds a small amount of logic. Worth it for user flexibility — DCP supports both.
- **Doctor check is per-project**: Unlike the other 4 checks (environment-level), the DCP check inspects the current working directory. This is a slight conceptual shift but justified — `replicator init` is also per-project, and doctor should verify what init scaffolded.
<!-- scaffolded by uf vdev -->
