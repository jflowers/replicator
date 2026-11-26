## ADDED Requirements

None. This change modifies existing requirements; no new
requirements are added.

## MODIFIED Requirements

### Requirement: Code Ownership

The repository MUST have a `.github/CODEOWNERS` file that assigns code
ownership to the `@unbound-force/overlords` team.

The file MUST include:
- A `*` catch-all pattern assigning `@unbound-force/overlords` as default
  owners for all files
- An explicit `.github/` entry for CI/CD workflows and GitHub configuration
- An explicit `.specify/memory/constitution.md` entry for governance documents
- An explicit `.opencode/` entry for agent configuration

Code owner approval is enforced by the `safe-settings: code repos`
organization ruleset defined in the `unbound-force/.github`
repository, which enables `require_code_owner_review` for pull
requests on this repository.

#### Scenario: PR triggers code owner review

- **GIVEN** a `.github/CODEOWNERS` file exists with a `*` catch-all pattern
- **WHEN** a contributor opens a PR modifying any file
- **THEN** GitHub requests a review from `@unbound-force/overlords`

#### Scenario: Branch protection enforces code owner approval

- **GIVEN** the `safe-settings: code repos` organization ruleset enables
  `require_code_owner_review` for pull requests on this repository
- **AND** a `.github/CODEOWNERS` file defines owners
- **WHEN** a PR is submitted without code owner approval
- **THEN** the PR cannot be merged

Previously the scenario's first GIVEN read: "`.github/settings.yml`
sets `require_code_owner_reviews: true`" and the requirement
carried no enforcement-source statement.

## REMOVED Requirements

None. No existing requirements are removed by this change.
