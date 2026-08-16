---
tag: add-dcp-config
author: jay-flowers
category: gotcha
created_at: 2026-08-16T18:40:02Z
identity: add-dcp-config-20260816T184002-jay-flowers
tier: draft
---

When fixing unchecked errors in test setup code (os.MkdirAll, os.WriteFile, os.ReadFile calls without error checking), apply the fix consistently across ALL test files touched by the change, not just the files where the new tests were added. In the add-dcp-config change, iteration 1 review found unchecked errors in agentkit_test.go and checks_test.go (new test code), but those fixes were not applied to pre-existing unchecked errors in init_test.go which was also modified by the change. The iteration 2 Testing reviewer caught this inconsistency and flagged it as HIGH severity. The pattern to follow: if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("setup MkdirAll: %v", err) } — wrap every setup call with t.Fatalf error checking.
