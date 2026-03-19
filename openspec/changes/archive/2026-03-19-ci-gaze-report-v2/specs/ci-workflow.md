# CI Workflow Requirements

## ADDED Requirements

### Requirement: Threshold check MUST run on PRs without secrets

A "Gaze threshold check" step MUST run on both PRs and pushes, using `--format=json` (no AI adapter), enforcing `--max-crapload`, `--max-gaze-crapload`, and `--min-contract-coverage` thresholds. This step MUST NOT require any repository secrets.

#### Scenario: Fork PR runs threshold check
- **GIVEN** a PR from a fork with no access to `OPENCODE_API_KEY`
- **WHEN** the Go 1.24 unit+integration job runs
- **THEN** the "Gaze threshold check" step executes and passes/fails based on thresholds alone

#### Scenario: Threshold violation fails the PR
- **GIVEN** a PR that increases CRAPload above `--max-crapload`
- **WHEN** the threshold check runs
- **THEN** the step exits non-zero and the PR check fails

### Requirement: AI-formatted report MUST run only on push to main

A "Gaze quality report" step MUST run only on push to main (not on PRs), using `--ai=opencode` with `OPENCODE_API_KEY`. The `npm install opencode-ai` step MUST also be conditional on push to main.

#### Scenario: Push to main runs AI report
- **GIVEN** a push to the `main` branch
- **WHEN** the Go 1.24 unit+integration job runs
- **THEN** the "Gaze quality report" step executes with `--ai=opencode`

#### Scenario: PR does NOT run AI report
- **GIVEN** a PR (from any source)
- **WHEN** the Go 1.24 unit+integration job runs
- **THEN** no "Gaze quality report" step executes

## MODIFIED Requirements

### Requirement: Both paths MUST use the same threshold values

The threshold check and the AI report MUST use the same `--max-crapload`, `--max-gaze-crapload`, and `--min-contract-coverage` values.

## REMOVED Requirements

None.
