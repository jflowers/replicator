# CI Workflow Security

## Dependency review

`.github/workflows/ci_dependencies.yml` handles pull requests targeting
`main` with the `pull_request` event. This record covers that pull-request
path only.

- **Default permissions:** `contents: read`, `issues: none`, and
  `pull-requests: none`. The two reusable-reviewer jobs repeat those
  read-only permissions.
- **Write boundary:** Only the comment and approval jobs request
  `pull-requests: write` (the comment job also requests `issues: read`). Both
  jobs run only when the event is `pull_request` and the author is
  `dependabot[bot]`; the approval job also requires successful reviewer
  results and validated low- or medium-risk release data.
- **Secrets:** The workflow declares no `secrets:` mapping or inheritance, so
  it does not pass repository or Dependabot secrets to jobs or reusable
  workflows. GitHub provides each job's `GITHUB_TOKEN`; the effective token
  permissions can be lower than the permissions requested by a job.
- **Checkout:** The workflow contains no checkout action. It does not clone or
  execute pull-request repository code.
- **Action integrity:** Every `uses:` reference is pinned to a full commit
  SHA, including both reusable workflows and the two step actions.
- **Failed-scan gate:** the general dependency reviewer reports its review
  conclusion from the scan step's `outcome`. With the v0.9.0 reviewer, a
  failed dependency scan surfaces as `failure` instead of being masked as
  `success`, so the approval job's `review_conclusion == 'success'` gate
  halts Dependabot auto-approval and the pull request requires manual
  review (fail-closed).

Untrusted pull-request code cannot execute with privileged credentials in this
workflow: it is never checked out, default permissions are read-only, and the
only write-capable jobs are limited to Dependabot pull requests.

### Effective token behavior and safe degradation

GitHub evaluates the requested job permissions against the event context. For
same-repository pull requests, the two reviewer jobs receive only their
declared read permissions; the comment and approval jobs do not run unless the
author is `dependabot[bot]`. For fork pull requests, GitHub supplies a
read-only `GITHUB_TOKEN` and does not make repository secrets available. The
Dependabot `pull_request` context also has a restricted, read-only token, so
its comment and approval jobs can receive an authorization response such as
`403 Resource not accessible by integration` even though they request
`pull-requests: write`.

The workflow does not compensate by restoring `pull_request_target`, broadening
permissions, or using a privileged token. It still runs both dependency
reviewers and preserves their analysis in the workflow logs. If report
publication receives the expected restricted-token authorization failure, the
report script emits a notice and skips the comment. If approval receives that
failure, it emits a notice, records a failed job, and creates no approval; a
maintainer must review and approve the Dependabot update manually. Other API
failures remain failures rather than being silently skipped.

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
