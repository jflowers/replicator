# Tasks: Skip CI Build, Test, and Lint for Learnings-Only PRs

**Change**: ci-skip-build-for-learnings

<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency on
  prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Workflow gating

- [x] 1.1 [P] In `.github/workflows/ci.yml`, add a `changes` job:
      checkout with the existing pinned `actions/checkout`
      (v7.0.1) and `fetch-depth: 0`; on `pull_request`, run
      `git diff --name-only --no-renames "$BASE_SHA...$HEAD_SHA"`
      with `BASE_SHA` and `HEAD_SHA` supplied via the step's
      `env:` block from `github.event.pull_request.base.sha` and
      `github.event.pull_request.head.sha` (never interpolated
      into `run:` -- event-context template injection is a zizmor
      finding; cf. `ci_dependencies.yml`), and
      set output `code` to `'true'` when any changed file lacks
      the `.uf/dewey/learnings/` prefix, `'false'` otherwise; on
      `push`, set `code: 'true'` unconditionally; fail the job
      if the diff command fails (CLS-FR-001). Give
      `build-and-test` `needs: [changes]` and
      `if: needs.changes.outputs.code == 'true'`, steps
      unchanged (CLS-FR-002). Add a `ci-ok` job named `CI OK`
      with `if: always()`, `needs: [changes, build-and-test]`,
      and a step that fails when any needed job result is
      `failure` or `cancelled` (CLS-FR-003).
- [x] 1.2 [P] In `.github/workflows/ci_checks.yml`, add the same
      `changes` job (CLS-FR-001); give `call-reusable-ci`
      `needs: [changes]` and the gating `if:`, with the reusable
      call and its permissions unchanged (CLS-FR-002); add a
      `ci-ok` job named `Standardized CI OK` with
      `if: always()`, `needs: [changes, call-reusable-ci]`, and
      the same fail-on-`failure`/`cancelled` step (CLS-FR-003).
      Do not add `paths` or `paths-ignore` to either workflow's
      `on:` block (CLS-FR-005).
- [x] 1.3 [P] Update `.github/settings.yml` so
      `required_status_checks.contexts` under the `main` branch
      protection lists exactly `CI OK` and `Standardized CI OK`
      (keeping `strict: true`), replacing `Build and Test` and
      `Standardized CI / Run linters`. The declarative file must
      agree with the live protection task 5.2 writes, so a
      settings sync cannot revert the required checks (CLS-FR-003).

## 2. Gating checkpoint

- [x] 2.1 Run `make check` and confirm vet plus the full test
      suite pass with the phase 1 workflow edits in place. The
      next phase MUST NOT begin while this checkpoint fails
      (meta-prompt MP-018).

## 3. Workflow gating regression test

- [x] 3.1 Add `internal/workflowtest/ci_gating_test.go` in the
      style of `ci_dependencies_test.go`, reading the repository
      workflow files with the existing `yamlSection` /
      `readRepositoryFile` helpers (no Node.js needed for
      structural assertions). Assert: both workflows define a
      `changes` job with output `code`; `build-and-test` and
      `call-reusable-ci` declare `needs` on `changes` and the
      gating `if:`; the `ci-ok` jobs are named `CI OK` and
      `Standardized CI OK`, use `if: always()`, need their
      workflow's `changes` job plus gated job, and fail on
      `failure`/`cancelled`; neither workflow's `on:` block
      contains `paths` or `paths-ignore`; and
      `.github/settings.yml` lists exactly `CI OK` and
      `Standardized CI OK` in `required_status_checks.contexts`
      (CLS-FR-006).

## 4. Regression test checkpoint

- [x] 4.1 Run `make check` and confirm the new gating test
      passes and no existing test regresses. The next phase MUST
      NOT begin while this checkpoint fails (meta-prompt
      MP-018).

## 5. Required checks, docs, and verification

- [x] 5.1 Update `AGENTS.md` required-check references to the
      new aggregate names (`CI OK`, `Standardized CI OK`) while
      keeping job-name facts (`Build and Test` as the job that
      runs govulncheck) accurate. Sweep `README.md` and `docs/`
      for required-check claims and update any that go stale.
- [ ] 5.2 Update branch protection on `main` so the required
      checks are exactly `CI OK` and `Standardized CI OK`:
      fetch the current protection (`gh api
      repos/unbound-force/replicator/branches/main/protection`),
      rewrite `required_status_checks.contexts` to the two
      aggregate names, and PUT the full payload back otherwise
      unchanged. Requires a maintainer with admin. Perform it
      when this change merges, and push a trivial update to any
      other open PR so it emits the new checks (design D6,
      risk "Required-check transition window").
- [x] 5.3 Run `openspec validate ci-skip-build-for-learnings`
      and resolve every validation error.
- [x] 5.4 Constitution alignment verification per the proposal:
      I (N/A -- no MCP or artifact interfaces), II (PASS -- no
      runtime or new action dependency), III (PASS -- observable
      `skipped` states, fail-closed aggregates, no threshold or
      flag weakened), IV (PASS -- isolated workflow regression
      test, `make check` green).
      Audited independently against
      .specify/memory/constitution.md and the actual diffs, not
      the proposal's claims: I -- the change surface
      (git diff --name-status main) is workflow YAML,
      .github/settings.yml, AGENTS.md, spec artifacts, and one
      test file; no MCP, tools, comms, or cmd/ code. II -- the
      only new action reference reuses the already-pinned
      actions/checkout@3d3c42e5 (v7.0.1); go.mod, go.sum, and
      every non-test Go file are untouched. III -- the gated
      jobs' diffs add only `needs:` and `if:`; every step,
      flag, ratchet, and pin is unchanged; detection fails
      closed on diff errors; both aggregates run `if: always()`
      and fail on failure/cancelled; settings.yml lists exactly
      the two aggregates with `strict: true` kept. IV --
      ci_gating_test.go uses only the pure os.ReadFile/regex
      helpers (no Node, network, or git subprocess); `make
      check` green and `openspec validate` error-free,
      2026-11-24.
- [ ] 5.5 Post-merge acceptance (issue #182 "Done when"): open a
      PR adding only a file under `.uf/dewey/learnings/` and
      confirm every required check passes within a couple of
      minutes without build, tests, or linters running; confirm
      a PR touching any other file, alone or with learnings,
      runs full CI as before; confirm push runs on `main` are
      unchanged.
- [x] 5.6 Phase checkpoint (MP-018 exception, design D9): this
      phase produces no code, so it intentionally closes without
      a test-suite run; its gates are `openspec validate`
      (task 5.3), the constitution check (task 5.4), and the
      post-merge acceptance (task 5.5). Recorded here so the
      exception is auditable in tasks.md, not only in design
      prose.
