# Tasks: Skip CI Build and Test for Learnings-Only PRs

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
      (v7.0.1) and `fetch-depth: 2`; on `pull_request`, check
      that the merge commit's second parent (`HEAD^2`) is
      `HEAD_SHA`, supplied via the step's `env:` block from
      `github.event.pull_request.head.sha` (never interpolated
      into `run:` -- event-context template injection is a zizmor
      finding; cf. `ci_dependencies.yml`), run
      `git diff --raw --no-renames 'HEAD^1' HEAD` (never a
      diff against `pull_request.base.sha`, which can predate
      the `main` merged into the branch), and
      set output `code` to `'true'` when any changed file lacks
      the `.uf/dewey/learnings/` prefix, `'false'` otherwise; on
      `push`, set `code: 'true'` unconditionally; fail the job
      if the parent check or the diff command fails
      (CLS-FR-001). Give
      `build-and-test` `needs: [changes]` and
      `if: needs.changes.outputs.code == 'true'`, steps
      unchanged (CLS-FR-002). Add a `ci-ok` job named `CI OK`
      with `if: always() && !(github.event_name == 'push' && cancelled())`,
      `needs: [changes, build-and-test]`, and a step that fails
      when any needed job result is `failure` or `cancelled`
      (CLS-FR-003). Check out only on `pull_request`
      (CLS-FR-004).
- [x] 1.2 [P] Leave `.github/workflows/ci_checks.yml` as on
      `main`: no `changes` job, no gate on `call-reusable-ci`,
      no aggregate, so secret scanning and PR title validation
      run on every PR (CLS-FR-002). Do not add `paths` or
      `paths-ignore` to either workflow's `on:` block
      (CLS-FR-005).
- [x] 1.3 [P] Update `.github/settings.yml` so
      `required_status_checks.contexts` under the `main` branch
      protection lists exactly `CI OK` and
      `Standardized CI / Run linters` (keeping `strict: true`),
      replacing `Build and Test` with `CI OK`. The declarative file must
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
      structural assertions). Assert: `ci.yml` defines a
      `changes` job with output `code`; `build-and-test`
      declares `needs` on `changes` and the gating `if:`; the
      `ci-ok` job is named `CI OK`, uses `always()` exempting
      only cancelled push runs, needs
      `changes` and `build-and-test`, and fails on
      `failure`/`cancelled`; `ci_checks.yml`'s only job is
      `call-reusable-ci`, with neither `if:` nor `needs`;
      neither workflow's `on:` block contains `paths` or
      `paths-ignore`; and `.github/settings.yml` lists exactly
      `CI OK` and `Standardized CI / Run linters` in
      `required_status_checks.contexts` (CLS-FR-006).

## 4. Regression test checkpoint

- [x] 4.1 Run `make check` and confirm the new gating test
      passes and no existing test regresses. The next phase MUST
      NOT begin while this checkpoint fails (meta-prompt
      MP-018).

## 5. Required checks, docs, and verification

- [x] 5.1 Update `AGENTS.md` required-check references
      (`CI OK`, `Standardized CI / Run linters`) while
      keeping job-name facts (`Build and Test` as the job that
      runs govulncheck) accurate. Sweep `README.md` and `docs/`
      for required-check claims and update any that go stale.
- [ ] 5.2 Update branch protection on `main` so the required
      checks are exactly `CI OK` and
      `Standardized CI / Run linters`: fetch the current
      protection (`gh api
      repos/unbound-force/replicator/branches/main/protection`),
      replace `Build and Test` with `CI OK` in
      `required_status_checks.contexts`, and PUT the full payload
      back otherwise
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
      job's diff adds only `needs:` and `if:`; every step,
      flag, ratchet, and pin is unchanged; ci_checks.yml is
      unchanged from main so secret scanning and PR title
      validation run on every PR; detection fails closed on
      diff errors; the aggregate runs `always()` (skipped only
      for cancelled push runs, never PRs) and fails
      on failure/cancelled; settings.yml lists exactly `CI OK`
      and `Standardized CI / Run linters` with `strict: true`
      kept. IV --
      ci_gating_test.go uses only the pure os.ReadFile/regex
      helpers (no Node, network, or git subprocess); `make
      check` green and `openspec validate` error-free,
      2026-11-24.
- [ ] 5.5 Post-merge acceptance (issue #182 "Done when"): open a
      PR adding only a file under `.uf/dewey/learnings/` and
      confirm every required check passes within a few minutes
      without the Go build or tests running, and that the
      secret scan and PR title check still ran; confirm
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

## 6. Align with sibling repos

- [x] 6.1 Switch the `changes` diff to `--raw` and count only
      regular `.md` files (or deletions) under
      `.uf/dewey/learnings/` as learnings; symlinks, submodules,
      empty diffs and unparsable lines set `code: 'true'`
      (CLS-FR-001)
- [x] 6.2 Add `TestCIGatingWorkflows_DetectionScriptBehavior`,
      which runs the extracted `changes` script against real
      merge commits (15 scenarios), and update the text pins
- [x] 6.3 Address the council's MEDIUM test-depth advisories: move
      the `CI OK` verdict from a step-level `contains() ||
      contains()` condition into a shell step over
      `join(needs.*.result, ' ')`, add
      `TestCIGatingWorkflows_AggregateVerdictBehavior` (9 result
      combinations), and pin the `changes` output mapping,
      `id: detect`, and the `HEAD_SHA` env indirection. Each
      pinned property was checked by a mutation that the tests
      catch.
