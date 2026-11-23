---
tag: mirror-parity
author: jay-flowers
category: pattern
created_at: 2026-10-28T20:21:55Z
identity: mirror-parity-20261028T202155-jay-flowers
tier: draft
---

Mirror-parity verification from a diff alone: when a PR edits both a canonical file and its byte-identical scaffold mirror (e.g., .opencode/commands/uf.X.md and internal/scaffold/assets/opencode/commands/uf.X.md), identical git `index <old>..<new>` blob-hash lines across the pair plus identical hunk bodies prove byte parity without shell access or a fresh checkout — useful when the review workspace sits on a stale branch (working tree on main behind the PR base, so the after-state files are unreadable). Corroborate with CI: the Build and Test check runs `go test ./internal/scaffold/...`, which includes the drift-detection test. Matching pre- and post-image abbreviated blob hashes across two paths means both blobs are content-identical; abbreviation collisions across a PR's file set are negligible.
