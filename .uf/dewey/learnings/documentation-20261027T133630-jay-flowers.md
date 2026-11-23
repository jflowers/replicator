---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-10-27T13:36:30Z
identity: documentation-20261027T133630-jay-flowers
tier: draft
---

pflag's MarkDeprecated (cobra's cmd.Flags().MarkDeprecated) sets the flag's Hidden=true, so a deprecated flag and its usage string never render in --help; users see only the runtime deprecation notice printed when the flag is passed. Consequence for review triage: a stale usage string on a deprecated no-op flag is maintainer-facing source text — a documentation-accuracy defect with no user-visible rendering, classifying LOW rather than MEDIUM. This corrects the --help-rendering claim in learning documentation-20261026T230309 (which asserted deprecated flags render in help with their usage text plus a "(DEPRECATED: ...)" suffix); the cli-testing-20261026T230306 account is the accurate one. Verified during the replicator init-smart-sync code review (PR #149): cmd/replicator/init.go retains "Overwrite existing agent kit files" on the now-deprecated --force registration — hidden from help, adjacent to an accurate MarkDeprecated notice.
