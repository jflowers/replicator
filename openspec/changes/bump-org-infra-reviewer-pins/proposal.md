# Proposal: Bump Both org-infra Reviewer Pins to v0.9.0

**Change**: bump-org-infra-reviewer-pins
**Issue**: [#156](https://github.com/unbound-force/replicator/issues/156)
**Supersedes**: [#152](https://github.com/unbound-force/replicator/pull/152) (single-reference Dependabot bump)

## Why

`internal/workflowtest/ci_dependencies_test.go` asserts both org-infra
reusable-workflow references against one shared `orgInfraSHA`
constant. PR #152 bumps only the deps reviewer to v0.9.0 and leaves the
Dependabot reviewer at v0.8.0, so `make check` fails at
`ci_dependencies_test.go:124` and the required `Build and Test` check is
red.

Because the two references share a single constant, any single-reference
bump breaks the other assertion by construction. Bumping both references
in lockstep keeps the shared constant and is smaller than restructuring
the test to support per-reference versions.

## What Changes

1. `.github/workflows/ci_dependencies.yml:29` (`call_deps_reviewer`) —
   bump `uses: .../reusable_deps_reviewer.yml@` from
   `bd3718a218d649b093269fe4a979c4a4632dfad2 # v0.8.0` to
   `8f347ba1aa2c7e9eb3b70e69c1709ddb456cd64a # v0.9.0`.

2. `.github/workflows/ci_dependencies.yml:37` (`call_dependabot_reviewer`) —
   bump `uses: .../reusable_dependabot_reviewer.yml@` to the same SHA and
   `# v0.9.0` comment.

3. `internal/workflowtest/ci_dependencies_test.go:19` — update
   `orgInfraSHA` to `8f347ba1aa2c7e9eb3b70e69c1709ddb456cd64a`.

The SHA is treated as untrusted until re-derived against
`complytime/org-infra` at authoring time (CI convention CI-002): resolve
the `v0.9.0` tag and dereference to the commit SHA before writing the
reference. As of this proposal, `v0.9.0` is a lightweight tag pointing at
`8f347ba1aa2c7e9eb3b70e69c1709ddb456cd64a`.

Full 40-character SHA pins and accurate `# vX.Y.Z` version comments stay
intact. Pinned action SHAs are a protected gate value (AGENTS.md
Gatekeeping Value Protection item 4); they MUST NOT be weakened, loosened,
or regex-relaxed to green the test.

**Out of scope (non-normative)**: other org-infra references in
`ci_checks.yml:34` and `release.yml:47,63` remain on v0.8.0 and can be
bumped separately. They are not asserted by `orgInfraSHA`. This is an
informational boundary, not a requirement that those files must stay on
v0.8.0.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `dependency-review-automation`: refines the existing "Organization
  Dependency Review" pin contract so both org-infra reviewer references are
  pinned to the same v0.9.0 SHA, with explicit CI-001 (pin), CI-002
  (authoring-time re-derivation), and CI-003 (version comment) citations.
  The base spec for this capability currently lives under
  `openspec/changes/automate-dependency-update/` and is not yet synced into
  `openspec/specs/`, so `openspec validate` refuses a MODIFIED delta; this
  change therefore adds the refined requirement under the same capability
  namespace. Once the predecessor is archived, this delta merges into
  `dependency-review-automation` rather than a parallel capability.

### Removed Capabilities
- None.

## Impact

- **CI**: the `Dependencies` workflow runs the v0.9.0 reusable reviewers.
- **Behaviour change**: our workflow auto-approves Dependabot PRs only when
  `review_conclusion == 'success'`. After the bump, a failed dependency
  scan reports `failure`, so auto-approval stops and the PR routes to
  manual review. This fails closed and is the intended upstream fix, not a
  regression.
- **Tests**: `TestCIDependenciesWorkflow_StructureEnforcesGuardedApproval`
  in `internal/workflowtest/ci_dependencies_test.go` is updated to expect
  the new shared SHA. Existing consumer-fixture tests (e.g. the
  unsuccessful-review-conclusion case) already cover the fail-closed
  reporting path, so no new consumer test is required.
- **Docs**: the fail-closed scan-failure behaviour is recorded in the
  established in-repo record `docs/ci-workflow-security.md` (the durable
  home for this workflow's approval gating), and summarised in the PR
  description. No website documentation issue is required (CI/CD pipeline
  change is exempt from the Website Documentation Gate).

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

No MCP tool interfaces, artifact formats, or inter-agent communication
change. This is a CI dependency-pin update.

### II. Composability First

**Assessment**: PASS

The change is limited to pinned workflow references and the test constant
that asserts them. It introduces no runtime dependency and preserves the
standalone, single-binary behaviour of replicator.

### III. Observable Quality

**Assessment**: PASS

The workflow regression test continues to assert both pins against a full
40-character SHA with a version comment, keeping the quality gate
machine-verifiable. The upstream change makes failed dependency scans
observable as `failure` instead of masking them as `success`.

### IV. Testability

**Assessment**: PASS

`internal/workflowtest/ci_dependencies_test.go` asserts the pins in
isolation with no external services, and the updated constant keeps those
assertions meaningful rather than weakening them.
