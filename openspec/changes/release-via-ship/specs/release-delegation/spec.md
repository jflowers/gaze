## ADDED Requirements

### Requirement: Local gates precede delegation

The release workflow MUST run `validate-module-major`,
`validate-workflow-floor`, and `validate-ai-report-default` to completion
before the release job, and the release job MUST declare them in `needs:`.

#### Scenario: Gates run first

- **GIVEN** `.github/workflows/release.yml` with its three validation jobs
- **WHEN** the release job is defined
- **THEN** it MUST contain `needs: [validate-module-major, validate-workflow-floor, validate-ai-report-default]`

### Requirement: Delegation is pinned and named

The release job MUST call
`jflowers/ship/.github/workflows/go-release.yml` pinned to a full
40-character commit SHA carrying its version comment, and the workflow
MUST keep the name `Release` (post-release-install-check.yml triggers on
its completion).

#### Scenario: Pinned call

- **GIVEN** the release job
- **WHEN** it delegates the pipeline
- **THEN** the `uses:` line MUST match `jflowers/ship/\.github/workflows/go-release\.yml@[0-9a-f]{40} # v\d+\.\d+\.\d+`

### Requirement: Caller contract is asserted

`TestReleaseWorkflow_CallsShip` MUST assert the caller side of the
delegation: the workflow name, the SHA-pinned call, the `needs:`
ordering, every forwarded `with:` input, the six `MACOS_*` secret
mappings, the `HOMEBREW_TAP_TOKEN` rename, and the absence of
`secrets: inherit`, `vars.MACOS_SIGN_IDENTITY`, and
`Developer ID Application:`.

#### Scenario: A dropped with: input fails the guard

- **GIVEN** the release job's `with:` block
- **WHEN** any forwarded input (`tag`, `ci_checks`, `skip_semver_check`, `skip_ci_checks`, `skip_unreleased_check`, `tap`) is removed or mistyped
- **THEN** `TestReleaseWorkflow_CallsShip` MUST fail

### Requirement: Prerelease decision is recorded

The release workflow MUST document the `allow_prerelease` decision at the
`with:` block: the omission is deliberate and behavior-preserving against
ship's declared default.

#### Scenario: Omission documented

- **GIVEN** ship v0.2.1 declares `allow_prerelease` with `default: true`
- **WHEN** the caller omits the input
- **THEN** the `with:` block MUST carry a comment recording the verified default and the deliberate omission
