# CI Workflow Requirements

## ADDED Requirements

### Requirement: Reusable AI report MUST be provider-agnostic

`.github/workflows/reusable_gaze_ai_report.yml` MUST expose a
`workflow_call` interface where the model, the opencode provider
configuration, the coverage artifact name, and the thresholds are
caller-supplied inputs, and the provider API key is passed as the
`provider-key` secret. The workflow MUST NOT hard-code any provider,
model, or credentials, and MUST NOT run the caller's test suite.

#### Scenario: Consumer runs the report with its own provider

- **GIVEN** a repository calling the workflow with an opencode config
  and the `provider-key` secret available
- **WHEN** the reusable workflow job runs
- **THEN** `gaze report ./... --ai=opencode` runs with the caller's
  model and provider, and the report is appended to the job summary

#### Scenario: Absent secret skips the report without failing

- **GIVEN** a fork PR where the `provider-key` secret is unavailable
- **WHEN** the reusable workflow job runs
- **THEN** the install and report steps are skipped with a notice and
  the job succeeds

### Requirement: Coverage MUST arrive via artifact handoff

The reusable workflow MUST obtain the coverprofile by downloading the
caller-named artifact from the calling workflow run and MUST accept the
path via the `coverprofile` input (default `coverage.out`).

#### Scenario: Caller uploads coverage.out

- **GIVEN** the calling workflow uploaded an artifact containing
  `coverage.out`
- **WHEN** the reusable workflow job runs
- **THEN** the artifact is downloaded and the report runs with
  `--coverprofile` pointing at the extracted file

### Requirement: Thresholds MUST be enforced on the AI report path

When threshold inputs are set, the report command MUST include the
corresponding `--max-crapload`, `--max-gaze-crapload`, and
`--min-contract-coverage` flags, and a threshold breach MUST fail the
job.

#### Scenario: Threshold breach fails the job

- **GIVEN** threshold inputs set and a report run where a metric
  breaches a threshold
- **WHEN** `gaze report` exits non-zero
- **THEN** the reusable workflow job fails

### Requirement: AI report MUST run only where secrets are available

The AI report MUST execute only where the `provider-key` secret
resolves (push to main); the `gaze-ai-report` job MAY run on PRs with
the install and report steps skipped to validate the wiring. The
PR-safe "Gaze threshold check" step MUST remain unchanged in
`unit-and-integration`.

#### Scenario: Push to main runs the EnMaaS report

- **GIVEN** a push to `main` with `ENMAAS_KEY` configured
- **WHEN** the `gaze-ai-report` job runs
- **THEN** the report runs with `enmaas/rits/zai-org/glm-5-3` and
  thresholds 38/6/8

#### Scenario: GCP machinery is gone

- **GIVEN** the `unit-and-integration` job
- **WHEN** the workflow runs
- **THEN** no GCP WIF auth, setup-gcloud, or google-vertex model
  references remain

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
