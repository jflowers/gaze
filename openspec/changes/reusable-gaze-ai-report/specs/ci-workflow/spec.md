# CI Workflow Requirements

## ADDED Requirements

### Requirement: Reusable AI report MUST be provider-agnostic

`.github/workflows/reusable_gaze_ai_report.yml` MUST expose a
`workflow_call` contract where the coverage artifact and thresholds are
caller inputs, the optional `model` input is passed through when set, and
the consumer's entire opencode configuration — provider, model,
permissions, literal API key — arrives as the `opencode-config` secret,
exported as `OPENCODE_CONFIG_CONTENT`. The workflow MUST NOT hard-code any
provider, model, or credentials, and MUST NOT run the caller's tests.

#### Scenario: Consumer runs the report with its own provider

- **GIVEN** a repository calling the workflow with the
  `opencode-config` secret available
- **WHEN** the reusable workflow job runs
- **THEN** `gaze report ./... --ai=opencode` runs with the model and
  provider from the secret's config, and the report is appended to the
  job summary

#### Scenario: Absent secret skips the report without failing

- **GIVEN** a fork PR where the `opencode-config` secret is unavailable
- **WHEN** the reusable workflow job runs
- **THEN** the opencode install and report steps are skipped with a
  notice, the extracted coverprofile is still validated, and the job
  succeeds

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

The AI report MUST execute only where the `opencode-config` secret
resolves (push to main and same-repo PRs; fork PRs never receive org
secrets); the `gaze-ai-report` job MAY run on fork PRs with the
opencode install and report steps skipped to validate the wiring. The
PR-safe "Gaze threshold check" step MUST remain unchanged in
`unit-and-integration`.

#### Scenario: Push to main runs the EnMaaS report

- **GIVEN** a push to `main` with the org secret
  `OPENCODE_CONFIG_CONTENT` configured (EnMaaS, GLM 5.3)
- **WHEN** the `gaze-ai-report` job runs
- **THEN** the report runs with the config's
  `enmaas/rits/zai-org/glm-5-3` model and thresholds 38/6/8

#### Scenario: GCP machinery is gone

- **GIVEN** the `unit-and-integration` job
- **WHEN** the workflow runs
- **THEN** no GCP WIF auth, setup-gcloud, or google-vertex model
  references remain

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
