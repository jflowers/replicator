## Context

PR #205 set `protection: null` for `main` in
`.github/settings.yml` and rewrote
`TestCIGatingSettings_ClassicProtectionRetired` to match,
deleting the assertion that `.github/settings.yml` lists the
required checks in `required_status_checks.contexts`. Required
checks are now enforced by the `safe-settings: replicator
required checks` ruleset in unbound-force/.github
(`repos/replicator.yml`, merged as unbound-force/.github#82),
which also drops strict required-status-checks policy.
Four live spec surfaces (CLS-FR-003, CLS-FR-006, the
ci-checks Branch Protection requirement, and the
ci-security-hygiene code-owner scenario) still name
`.github/settings.yml` as the enforcement source, and no
change re-pointed them.

## Goals / Non-Goals

### Goals

- Re-point the four live spec surfaces at the safe-settings
  rulesets as the branch-protection enforcement source.
- Rebuild the workflowtest settings assertion around what
  this repository still owns: `.github/settings.yml` must not
  manage branch protection (`main` sets `protection: null`,
  no `required_status_checks` key).
- Make that assertion structural — scoped to the `main`
  branch entry and matched by key shape — so comment text and
  formatting changes neither weaken it nor fail it
  spuriously (review feedback items 5 and 6 on PR #205).
- Keep `make check` failing when classic protection or
  `required_status_checks` re-enters `.github/settings.yml`.

### Non-Goals

- Verifying the live ruleset from CI. A Go test that calls
  the GitHub API would require network access and
  credentials, violating Constitution IV; the cross-repo
  observability gap is tracked as follow-up work.
- Changing any runtime behavior. `.github/settings.yml`
  already carries `protection: null`, and the ruleset is
  already active.
- Modifying the workflows themselves (`ci.yml`,
  `ci_checks.yml`) — CLS-FR-001 through CLS-FR-005 wiring is
  untouched.

## Decisions

1. **Enforcement source of record is unbound-force/.github.**
   The spec deltas name the ruleset (`safe-settings:
   replicator required checks`) and its defining file
   (`repos/replicator.yml`) so agents and humans can locate
   the policy without tribal knowledge (Constitution I).
2. **The local regression test pins the negative space.**
   This repository cannot test the ruleset, but it can test
   that its own settings file never re-introduces classic
   protection. `main`'s `protection` MUST be `null` and the
   file MUST NOT declare `required_status_checks` — that is
   the invariant `make check` enforces (Constitution III).
3. **Structural assertions over text matching.** The test
   scopes assertions to the `main` entry under `branches:`
   and matches `required_status_checks` by key shape
   (`(?m)^\s*required_status_checks:`), not by substring
   anywhere in the file. Comments may freely name the keys
   they explain, and non-semantic formatting changes do not
   fail the suite (review items 5 and 6).
4. **Scenario GIVEN re-pointed, not the intent.** The
   ci-security-hygiene code-owner scenario keeps its
   observable outcome (PR without code-owner approval cannot
   merge); only the enforcement mechanism named in the GIVEN
   changes, matching the deployed org ruleset.
5. **Supersession noted inline.** Each MODIFIED requirement
   carries a `Previously:` annotation quoting the retired
   sentence, so the delta reads as history, not silent
   rewriting.

## Risks / Trade-offs

- **Cross-repo enforcement is unverifiable from this
  repository's test suite.** Accepted: a live-API test would
  trade Constitution IV compliance for coverage; follow-up
  work tracks a release-checklist or integration check.
- **Spec deltas depend on unarchived changes.** The MODIFIED
  requirements modify text introduced by
  ci-skip-build-for-learnings, adopt-org-infra-ci, and
  ci-security-hygiene, which are live but not archived. This
  is the corpus's existing convention (all requirements live
  in unarchived changes); archiving order must apply this
  change after those.
- **Ruleset drift is invisible locally.** If
  unbound-force/.github reverts the ruleset, no test here
  fails. Documented in the proposal's Non-Goals and tracked
  as the filed follow-up; the alternative (network test) is
  constitutionally barred.
