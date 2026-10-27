---
tag: go-testing
author: jay-flowers
category: pattern
created_at: 2026-10-26T23:03:14Z
identity: go-testing-20261026T230314-jay-flowers
tier: draft
---

Go tests can assert "file was NOT rewritten" beyond content equality by pinning mtime: call os.Chtimes(path, pinned, pinned) with a far-past timestamp (e.g., time.Unix(1000000000, 0)), capture os.Stat().ModTime() before and after the operation, and assert equality. Any rewrite refreshes mtime to now, so a far-past pin makes the check deterministic without sleeps and immune to filesystem timestamp-granularity issues. Pattern used in replicator's TestScaffold_SmartSyncSkipsIdenticalFile and TestRunInit_ForceIsNoOp (init smart-sync change).
