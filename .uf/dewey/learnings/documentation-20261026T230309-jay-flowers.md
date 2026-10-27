---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-10-26T23:03:09Z
identity: documentation-20261026T230309-jay-flowers
tier: draft
---

When a change supersedes a CLI flag's semantics (deprecation, no-op conversion, default flip), the documentation supersession sweep must include the flag's own usage string at its registration site in source code, not just markdown artifacts. pflag (cobra) renders deprecated flags in --help with their original usage text plus a "(DEPRECATED: ...)" suffix, so a stale usage string surfaces as self-contradictory help text precisely when users consult help during a breaking migration. Observed in replicator opsx/init-smart-sync (2026-10-26): the .md sweep was complete (CHANGELOG, AGENTS.md, mirrored skills, spec supersession notes), but cmd/replicator/init.go's --force registration still read "Overwrite existing agent kit files" after --force became a deprecated no-op — because the sweep pattern grepped only *.md files. Rule of thumb: when a flag's semantics change, grep for the flag name across flag-registration sites (cmd.Flags()...) in addition to *.md, and treat each flag's usage string as a restatement of its behavior contract.
