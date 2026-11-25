# CI Workflow Security

## Dependency review

`.github/workflows/ci_dependencies.yml` reviews the dependency changes of
every pull request targeting `main`, with the `pull_request` event (and of
pushes to `main`).

- **Permissions:** `contents: read`, `issues: none`, and
  `pull-requests: none`, for the workflow and for its one job, which
  repeats them.
- **What it runs:** one job, `General`, calling complytime/org-infra's
  `reusable_deps_reviewer.yml` (`actions/dependency-review-action`) pinned
  to a full commit SHA. It writes nothing: no comment, approval or merge.
- **Secrets and checkout:** no `secrets:` mapping or inheritance, and no
  checkout action in this workflow.

## Dependabot auto-merge

`.github/workflows/dependabot-automerge.yml` calls jflowers/ship's
`deps-automerge.yml`, pinned to a full commit SHA, with the
`pull_request` event.

- **Permissions:** `permissions: {}` for the workflow; the one job asks for
  `contents: write` and `pull-requests: write`, which turning on auto-merge
  needs.
- **Who it acts for:** ship's job runs only when the actor and the PR author
  are `dependabot[bot]` and the head branch is in this repository, so fork
  PRs and people's pushes to a Dependabot branch never get past its `if`.
- **What it does:** for a patch or minor update of a package owned by
  jflowers, unbound-force or zero-dot-force, it runs
  `dependency-review-action`, checks that the repository allows auto-merge
  and that `main` requires at least one status check, then turns on
  auto-merge by rebase. GitHub merges only after the required checks pass
  and the code-owner review the org ruleset requires. Other updates are left
  for a maintainer.
- **Release age:** Dependabot's `cooldown` (`.github/dependabot.yml`) holds
  third-party releases for a day before opening a PR; our own orgs'
  releases are excluded.
- **Secrets:** none are passed.

## Fullsend shim

`.github/workflows/fullsend.yaml` is a Fullsend-managed generated shim. It
receives `issues` (`opened`, `edited`, `labeled`), `issue_comment` (`created`),
`pull_request_target` (`opened`, `synchronize`, `ready_for_review`, `closed`,
`labeled`, `unlabeled`), and `pull_request_review` (`submitted`) events. The
workflow default is `permissions: {}`. Its `dispatch` job grants `actions`,
`id-token`, `contents`, `issues`, and `pull-requests` write permissions plus
`packages: read`; the separate `stop-fix` job grants `contents: read`,
`issues: write`, and `pull-requests: write`.

- **Secrets and delegation:** `dispatch` forwards
  `FULLSEND_GCP_WIF_PROVIDER`, `FULLSEND_GCP_PROJECT_ID`,
  `OTEL_EXPORTER_OTLP_TRACES_HEADERS`, and `OTEL_EXPORTER_OTLP_HEADERS` to
  `fullsend-ai/fullsend/.github/workflows/reusable-dispatch.yml` pinned at
  `d5f36921ac754705619f38c637ef692873809fbc` (`v0.43.0`). That reusable
  workflow is the remote trust boundary for the privileged dispatch and any
  agent execution; it is not implemented or verified by this repository.
- **Checkout and local execution:** The shim has no checkout action. It does
  not execute untrusted pull-request repository code locally, including when
  `pull_request_target` supplies the base-branch workflow. The `stop-fix` job
  runs only its checked-in shell step against event metadata and `github.token`.
- **Ownership:** Fullsend generates this file and identifies its upstream
  scaffold source in the header. Local maintainers must use an
  upstream-supported Fullsend configuration or generator update rather than
  editing the generated shim directly.
- **zizmor exception:** No compatible upstream hardening path currently exists.
  Issue #127 explicitly authorizes a finding-specific `dangerous-triggers`
  suppression for `.github/workflows/fullsend.yaml` in `.github/zizmor.yml`.
  The suppression does not disable zizmor or apply to another rule or workflow;
  remove it when Fullsend supplies a compatible hardening update.
