---
tag: add-dcp-config
author: jay-flowers
category: pattern
created_at: 2026-08-16T18:40:06Z
identity: add-dcp-config-20260816T184006-jay-flowers
tier: draft
---

The doctor package's Run() function signature was extended from Run(store *db.Store, cfg *config.Config) to Run(store *db.Store, cfg *config.Config, projectDir string) to support per-project checks like checkDCPConfig. The key design decision (D9) was to accept an explicit directory parameter rather than calling os.Getwd() inside the check function. This enables test isolation with t.TempDir() — tests pass a temp directory directly instead of needing unsafe os.Chdir() calls. The os.Getwd() call lives in the CLI layer (cmd/replicator/doctor.go) where it's the caller's responsibility. This pattern should be followed for any future per-project doctor checks.
