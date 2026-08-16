## ADDED Requirements

### Requirement: ScaffoldDCP function

The `agentkit` package MUST export a `ScaffoldDCP(targetDir string) (ScaffoldResult, error)` function that creates or updates `.opencode/dcp.jsonc` with DCP protect-tag configuration.

#### Scenario: Fresh directory with no existing DCP config

- **GIVEN** a target directory with no `.opencode/dcp.jsonc` or `.opencode/dcp.json` file
- **WHEN** `ScaffoldDCP(targetDir)` is called
- **THEN** `.opencode/dcp.jsonc` MUST be created with `protectTags: true` under a `compress` key, and the result action MUST be "created"

#### Scenario: Existing DCP config with protectTags already set

- **GIVEN** a target directory with `.opencode/dcp.jsonc` containing `"protectTags": true`
- **WHEN** `ScaffoldDCP(targetDir)` is called
- **THEN** the file MUST NOT be modified, and the result action MUST be "skipped"

#### Scenario: Existing DCP config without protectTags

- **GIVEN** a target directory with `.opencode/dcp.jsonc` that does not contain `"protectTags"`
- **WHEN** `ScaffoldDCP(targetDir)` is called
- **THEN** the file MUST be replaced with the canonical DCP config content (including `protectTags: true`), and the result action MUST be "updated"

#### Scenario: Both `.dcp.jsonc` and `.dcp.json` exist

- **GIVEN** a target directory with both `.opencode/dcp.jsonc` and `.opencode/dcp.json`
- **WHEN** `ScaffoldDCP(targetDir)` is called
- **THEN** the function MUST operate on `.opencode/dcp.jsonc` (preferred extension) and ignore `.opencode/dcp.json`

#### Scenario: `.dcp.json` alias (without `c` extension)

- **GIVEN** a target directory with `.opencode/dcp.json` (not `.jsonc`) containing DCP configuration
- **WHEN** `ScaffoldDCP(targetDir)` is called
- **THEN** the function MUST operate on `.opencode/dcp.json` (respecting the user's extension choice), not create a new `.opencode/dcp.jsonc`

### Requirement: DCP config content

The scaffolded `.opencode/dcp.jsonc` MUST contain:
1. A `$schema` reference to the DCP JSON schema
2. Comments explaining the purpose of `protectTags`
3. A `compress` object with `protectTags: true`

### Requirement: Init command integration

The `replicator init` command MUST call `ScaffoldDCP()` after `Scaffold()` and render the DCP result using the same styled output (green for created, dim for skipped, yellow for updated).

#### Scenario: `replicator init` on a fresh directory

- **GIVEN** a directory that has not been initialized
- **WHEN** `replicator init` is run
- **THEN** `.opencode/dcp.jsonc` MUST exist alongside the 15 agent kit files, and the init output MUST include the DCP file status

#### Scenario: `replicator init` on an already-initialized directory

- **GIVEN** a directory that was previously initialized (`.opencode/dcp.jsonc` already exists with `protectTags: true`)
- **WHEN** `replicator init` is run again
- **THEN** `.opencode/dcp.jsonc` MUST NOT be modified, and the output MUST show "skipped" for the DCP file

### Requirement: ScaffoldDCP result shape

`ScaffoldDCP()` MUST return `(ScaffoldResult, error)` where `ScaffoldResult` has `Path` (string) and `Action` (string) fields. The `Action` field MUST be one of: "created", "skipped", "updated".

### Requirement: `.opencode/` directory creation

If the `.opencode/` directory does not exist, `ScaffoldDCP()` MUST create it before writing the config file.

### Requirement: Doctor DCP config check

The `doctor` package MUST include a `checkDCPConfig(projectDir string) CheckResult` function that verifies per-project DCP configuration health. The function MUST accept an explicit directory parameter for testability (no implicit `os.Getwd()`). The check MUST be registered as the 5th check in `Run()`.

#### Scenario: DCP config present with protectTags

- **GIVEN** a working directory with `.opencode/dcp.jsonc` containing `"protectTags": true` and `.opencode/commands/` containing files with `<protect>` tags
- **WHEN** `checkDCPConfig()` is called
- **THEN** the result status MUST be "pass" and the message MUST contain "protectTags enabled"

#### Scenario: No protect-tagged commands exist

- **GIVEN** a working directory with no `.opencode/commands/` directory or no files containing `<protect>` tags
- **WHEN** `checkDCPConfig()` is called
- **THEN** the result status MUST be "pass" and the message MUST contain "no protect-tagged commands"

#### Scenario: Protect-tagged commands exist but no DCP config

- **GIVEN** a working directory with `.opencode/commands/` containing files with `<protect>` tags but no `.opencode/dcp.jsonc` or `.opencode/dcp.json`
- **WHEN** `checkDCPConfig()` is called
- **THEN** the result status MUST be "warn" and the message MUST contain "replicator init"

#### Scenario: DCP config exists but missing protectTags

- **GIVEN** a working directory with `.opencode/dcp.jsonc` that does not contain `"protectTags"` and `.opencode/commands/` containing files with `<protect>` tags
- **WHEN** `checkDCPConfig()` is called
- **THEN** the result status MUST be "warn" and the message MUST contain "protectTags"

### Requirement: Doctor check result shape

The `checkDCPConfig()` function MUST return a `CheckResult` with `Name` set to `"dcp_config"`. The `Status` field MUST be "pass" or "warn" (never "fail" — DCP is optional).

### Requirement: Doctor check count update

`Run()` MUST return 5 results (up from 4) when all checks complete. The `dcp_config` check MUST be the 5th check.

## MODIFIED Requirements

### Requirement: `Run()` signature change

`Run()` MUST accept a `projectDir string` parameter: `Run(store *db.Store, cfg *config.Config, projectDir string) ([]CheckResult, error)`. The `projectDir` is passed to `checkDCPConfig()` for per-project checks. Callers (e.g., `cmd/replicator/doctor.go`) MUST pass `os.Getwd()` or equivalent.

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
