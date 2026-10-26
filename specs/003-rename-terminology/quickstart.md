# Quickstart: Terminology Rename + Agent Kit

**Branch**: `003-rename-terminology` | **Date**: 2026-04-06

## At a Glance

| What | Details |
|------|---------|
| **Goal** | Rename hive→org, swarmmail→comms, swarm→forge. Ship agent kit via `replicator init`. |
| **Phases** | 8 (dir renames → imports → tool names → fixtures → agent kit → init → docs → verify) |
| **Files changed** | ~35 Go files + 4 JSON fixtures + 15 new embedded files |
| **Tools renamed** | 45 of 53 (8 `hivemind_*` unchanged) |
| **New package** | `internal/agentkit/` (embed.FS + scaffold logic) |
| **New flag** | `replicator init --force` |
| **CI commands** | `go vet ./...` → `go test ./... -count=1 -race` → `go build` |

> **`replicator init --force` superseded 2026-10-26** by `openspec/changes/init-smart-sync/` -- smart-sync overwrites drifted kit files by default; `--skip-existing` restores the old skip behavior.

## Rename Cheat Sheet

```
hive_*      → org_*       (11 tools)
swarmmail_* → comms_*     (10 tools)
swarm_*     → forge_*     (24 tools)
hivemind_*  → hivemind_*  (8 tools, UNCHANGED)
```

## Package Rename Commands

```bash
# Phase 1: Directory renames
git mv internal/hive internal/org
git mv internal/swarmmail internal/comms
git mv internal/swarm internal/forge
git mv internal/tools/hive internal/tools/org
git mv internal/tools/swarmmail internal/tools/comms
git mv internal/tools/swarm internal/tools/forge
```

## Import Path Find-Replace

```
OLD: github.com/unbound-force/replicator/internal/hive
NEW: github.com/unbound-force/replicator/internal/org

OLD: github.com/unbound-force/replicator/internal/swarmmail
NEW: github.com/unbound-force/replicator/internal/comms

OLD: github.com/unbound-force/replicator/internal/swarm
NEW: github.com/unbound-force/replicator/internal/forge

OLD: github.com/unbound-force/replicator/internal/tools/hive
NEW: github.com/unbound-force/replicator/internal/tools/org

OLD: github.com/unbound-force/replicator/internal/tools/swarmmail
NEW: github.com/unbound-force/replicator/internal/tools/comms

OLD: github.com/unbound-force/replicator/internal/tools/swarm
NEW: github.com/unbound-force/replicator/internal/tools/forge
```

## Package Declaration Updates

```
internal/org/*.go:     package hive     → package org
internal/comms/*.go:   package swarmmail → package comms
internal/forge/*.go:   package swarm    → package forge
internal/tools/org/:   package hive     → package org
internal/tools/comms/: package swarmmail → package comms  (note: was "swarmmail" not "swarmmailtools")
internal/tools/forge/: package swarm    → package forge
```

## Import Alias Updates

In files that use import aliases (serve.go, docs.go, parity_test.go):

```go
// OLD
import (
    "github.com/unbound-force/replicator/internal/tools/hive"
    swarmmailtools "github.com/unbound-force/replicator/internal/tools/swarmmail"
    swarmtools "github.com/unbound-force/replicator/internal/tools/swarm"
)

// NEW
import (
    "github.com/unbound-force/replicator/internal/tools/org"
    commstools "github.com/unbound-force/replicator/internal/tools/comms"
    forgetools "github.com/unbound-force/replicator/internal/tools/forge"
)
```

**Note**: The `hive` tool package was imported without an alias (it didn't conflict). After rename to `org`, it still won't conflict, so no alias needed. But `comms` and `forge` tool packages need aliases to avoid conflicting with the domain packages `internal/comms` and `internal/forge`.

## Docs Category Map

```go
// OLD
{"hive_", "Hive"},
{"swarmmail_", "Swarm Mail"},
{"swarm_", "Swarm"},
{"hivemind_", "Memory"},

// NEW
{"org_", "Org"},
{"comms_", "Comms"},
{"forge_", "Forge"},
{"hivemind_", "Memory"},
```

## Agent Kit File Tree

```
.opencode/
├── command/
│   ├── forge.md           # /forge orchestrator
│   ├── org.md             # /org cell management
│   ├── inbox.md           # /inbox check messages
│   ├── forge-status.md    # /forge:status
│   └── handoff.md         # /handoff session end
├── skills/
│   ├── always-on-guidance/SKILL.md
│   ├── forge-coordination/SKILL.md
│   ├── replicator-cli/SKILL.md
│   ├── testing-patterns/SKILL.md
│   ├── system-design/SKILL.md
│   ├── learning-systems/SKILL.md
│   └── forge-global/SKILL.md
└── agents/
    ├── coordinator.md
    ├── worker.md
    └── background-worker.md
```

## Verification Commands

```bash
# Build + test (CI parity)
go vet ./...
go test ./... -count=1 -race
go build -o bin/replicator ./cmd/replicator

# Parity tests
go test -tags parity ./test/parity/ -count=1 -v

# Grep for stale names (expect zero matches)
grep -rn '"hive_\|"swarmmail_\|"swarm_' --include='*.go' | grep -v hivemind_

# Init smoke test
tmpdir=$(mktemp -d)
./bin/replicator init --path "$tmpdir"
find "$tmpdir" -type f | wc -l  # expect 16

# Tool count
# Start MCP server, call tools/list, verify:
#   11 org_* + 10 comms_* + 24 forge_* + 8 hivemind_* = 53
```

## Key Gotchas

1. **Build breaks between Phase 1 and Phase 2**: After `git mv`, all imports are broken. Phase 2 must complete before `go build` works again. Commit Phases 1+2 together.

2. **Package name ≠ directory name after git mv**: `git mv internal/hive internal/org` moves the directory but does NOT update `package hive` declarations inside the files. Must manually update all `package` declarations.

3. **Import alias conflicts**: After rename, `internal/org` (domain) and `internal/tools/org` (handlers) both have package name `org`. The tool handler package in `serve.go` and `docs.go` needs an alias (e.g., `orgtools`). Same for `comms`/`forge`.

4. **Parity fixture keys**: The JSON fixture files use tool names as top-level keys. The `request.name` field inside each fixture also contains the tool name. Both must be updated.

5. **`hivemind_*` is intentionally unchanged**: Don't rename these 8 tools. They're deprecated Dewey proxy stubs that will be removed in a future version.

6. **Description strings**: Some tool descriptions reference old names (e.g., "alias for hive_cells"). These must be updated too.

7. **Prompt-generating tools**: `swarm/spawn.go`, `swarm/review.go`, etc. generate text prompts that may contain tool name references. These must be updated to use `forge_*`, `org_*`, `comms_*` names.
