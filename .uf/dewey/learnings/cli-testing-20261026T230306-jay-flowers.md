---
tag: cli-testing
author: jay-flowers
category: gotcha
created_at: 2026-10-26T23:03:06Z
identity: cli-testing-20261026T230306-jay-flowers
tier: draft
---

cobra/pflag MarkDeprecated side effects: pflag sets the flag Hidden=true, so the deprecated flag (and its usage string) disappears from --help entirely — the usage string remains maintainer-facing source text and should still be kept accurate, but it never renders to users. The deprecation notice itself is printed through the command's output writer during flag parsing, so tests capture it with root.SetOut(&buf) before root.Execute() (pattern proven by TestInitCmd_ForceDeprecationNotice in cmd/replicator/init_test.go).
