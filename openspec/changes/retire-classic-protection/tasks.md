## 1. Spec deltas

- [ ] 1.1 Write `specs/ci-learnings-skip/spec.md` modifying
      CLS-FR-003 (enforcement source) and CLS-FR-006 (test
      asserts the retired local state, structurally)
- [ ] 1.2 Write `specs/ci-checks/spec.md` modifying the Branch
      Protection requirement (ruleset as enforcement source,
      `CI OK` + `Standardized CI / Run linters` pair,
      non-strict policy)
- [ ] 1.3 Write `specs/ci-security-hygiene/spec.md` modifying
      the Code Ownership scenario's enforcement source
      (organization ruleset)

## 2. Regression test

- [ ] 2.1 Rewrite `TestCIGatingSettings_ClassicProtectionRetired`
      in `internal/workflowtest/ci_gating_test.go`: assert the
      `branches` entry for `main` sets `protection: null`, that
      no `required_status_checks` key is declared, and that the
      assertions are scoped to the `main` entry and matched by
      key shape rather than substring
- [ ] 2.2 Verify the test fails when `protection: null` is
      replaced with a protection block or a
      `required_status_checks` key is added (temporarily, not
      committed)

## 3. Verification

- [ ] 3.1 Run `openspec validate --change
      retire-classic-protection` and resolve any reported
      issues
- [ ] 3.2 Run `make check` (vet + test) and confirm the suite
      passes
- [ ] 3.3 Confirm constitution alignment: the test reads only
      local files (Constitution IV), the spec deltas name the
      ruleset as enforcement source (Constitution I), and
      `make check` keeps the retirement observable
      (Constitution III)
