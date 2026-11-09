## Context

`internal/workflowtest/ci_dependencies_test.go` defines a single
`orgInfraSHA` constant (currently `bd3718a…`, v0.8.0) and asserts it
against both reviewer references in `.github/workflows/ci_dependencies.yml`:
`reusable_deps_reviewer.yml@` (line 124) and
`reusable_dependabot_reviewer.yml@` (line 125). The workflow currently
pins both references to `bd3718a… # v0.8.0` (`ci_dependencies.yml:29` and
`:37`).

PR #152 bumps only the deps reviewer to v0.9.0. Because the test couples
both references to one constant, the single-reference bump fails the
assertion at `ci_dependencies_test.go:124` and `make check` goes red — the
same failure this change resolves.

## Goals / Non-Goals

### Goals
- Bump both org-infra reviewer references to v0.9.0 in lockstep.
- Keep the single shared `orgInfraSHA` constant and update it to the v0.9.0 SHA.
- Preserve full 40-character SHA pins and accurate `# vX.Y.Z` comments.
- Make `make check` pass without weakening any assertion.
- Document the fail-closed auto-approval behaviour change.

### Non-Goals
- Restructuring the test to support per-reference version constants.
- Bumping the org-infra references in `ci_checks.yml` or `release.yml`.
- Changing consumer-side approval logic in `.github/workflows/ci_dependencies.yml`.
- Fixing the latent shared-constant coupling (scoped to a future change).

## Decisions

### D1: Lockstep bump over test restructure

Bumping both references and the shared constant together is smaller than
splitting the test into per-reference constants. It preserves a single
source of truth and matches the decision recorded in issue #156. The
restructuring alternative is deferred.

### D2: Re-derive the SHA at authoring time

Per CI-002, the v0.9.0 SHA is untrusted until re-resolved against
`complytime/org-infra`. At authoring time, resolve the `v0.9.0` tag and
dereference it to a commit (as of this proposal it is a lightweight tag at
`8f347ba1aa2c7e9eb3b70e69c1709ddb456cd64a`), then confirm
`reusable_dependabot_reviewer.yml` exists at that commit before writing
the reference. Do not copy the SHA from issue prose alone.

### D3: Keep the assertions strict

The change updates the expected value; it does not loosen the assertion.
The test continues to require a full 40-character SHA with a `# vX.Y.Z`
comment (`assertFullSHAPins`). Pinned action SHAs are a protected gate
value (AGENTS.md Gatekeeping Value Protection item 4), so the pins and the
assertion remain intact.

### D4: Treat the behaviour change as intended, document it

The v0.9.0 deps reviewer reports `REVIEW_CONCLUSION` from the step
`outcome` (the step keeps `continue-on-error: true`). A failed scan now
surfaces as `failure` instead of the previously masked `success`, so
Dependabot auto-approval halts and the PR routes to manual review. This is
fail-closed and is recorded in the PR description; no consumer logic
changes.

### D5: Leave out-of-scope references untouched

`ci_checks.yml:34` and `release.yml:47,63` are not asserted by
`orgInfraSHA`. They stay on v0.8.0 and are bumped separately.

## Risks / Trade-offs

### R1: More manual Dependabot reviews

Failed dependency scans now block auto-approval. This is the intended
fail-closed behaviour; the trade-off is additional manual review for
genuinely failing scans, which is preferable to masking failures.

### R2: Latent shared-constant coupling

One constant still serves two independently versioned references, so a
future single-reference bump re-breaks the test. Accepted and scoped out;
a follow-up could split constants per reference.

### R3: Upstream tag drift or retargeting

A moved or retargeted tag would invalidate the SHA. Mitigated by
re-derivation at authoring time and by pinning the full commit SHA, so the
workflow only changes when the pin is explicitly updated.

## Constitution Alignment

Refer to the proposal's Constitution Alignment: I (N/A — no artifact
interfaces), II (PASS — pinned references only, no runtime dependency),
III (PASS — machine-verifiable pins and observable fail-closed reporting),
IV (PASS — isolated workflow regression test).
