# Design: CI Gaze Report v2

## Context

The current workflow has a single gaze report step that requires `OPENCODE_API_KEY` and is conditional on `matrix.go-version == '1.24'`. This step fails on fork PRs because GitHub doesn't expose secrets to fork PR workflows.

## Goals / Non-Goals

### Goals
- PR checks catch threshold violations (CRAPload, GazeCRAPload, contract coverage) without requiring secrets
- Push to main produces the full AI-formatted report with opencode
- Both paths use the same thresholds for consistency
- The PR check runs on Go 1.24 only (same as today)

### Non-Goals
- Running the AI-formatted report on PRs (requires secrets, not possible from forks)
- Changing threshold values
- Adding the report to Go 1.25 or E2E jobs

## Decisions

### D1: Two conditional steps replacing one

Replace the current single "Gaze quality report" step with two steps:

1. **"Gaze threshold check"** — runs on PRs AND pushes, Go 1.24 only:
   ```yaml
   if: matrix.go-version == '1.24'
   run: |
     go build -o gaze ./cmd/gaze
     ./gaze report ./... \
       --format=json \
       --coverprofile=coverage.out \
       --max-crapload=16 \
       --max-gaze-crapload=5 \
       --min-contract-coverage=8 \
       > /dev/null
   ```
   This evaluates thresholds against JSON output (no AI adapter needed, no secrets). The `> /dev/null` discards the JSON payload since we only care about the exit code (thresholds pass/fail).

2. **"Gaze quality report"** — runs ONLY on push to main, Go 1.24 only:
   ```yaml
   if: matrix.go-version == '1.24' && github.event_name == 'push'
   ```
   This is the full `--ai=opencode` report that requires `OPENCODE_API_KEY`. It runs after merge, producing the human-readable report.

**Rationale**: The threshold check is the quality gate (blocks bad code). The AI report is the quality narrative (explains what to fix). Separating them means PRs get fast pass/fail feedback without waiting for AI formatting or failing on missing secrets.

### D2: Remove the Install OpenCode step from PR path

The `npm install -g opencode-ai@1.2.26` step is only needed for the AI-formatted report. Move its `if` condition to match the push-only step so it doesn't run on PRs (saves ~5s and avoids npm dependency for threshold-only checks).

### D3: coverprofile stays on all Go versions

The `-coverprofile=coverage.out` flag on the test step stays for both Go versions. It's harmless and the Go 1.25 job might use it in the future.

## Risks / Trade-offs

- **No AI report on PRs**: PR authors won't see the formatted quality narrative until after merge. The threshold check still catches regressions — the AI report is additive visibility, not a gate.
- **Double gaze run on push**: Both the threshold check and the AI report run on push to main. The threshold check is fast (~30s for JSON-only); the AI report takes ~2min. Total overhead is acceptable.
