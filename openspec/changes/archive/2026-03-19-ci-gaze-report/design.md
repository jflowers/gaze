# Design: CI Gaze Report

## Context

The unit+integration job in `.github/workflows/test.yml` runs `go test -race -count=1 -short ./...` on both Go 1.24 and 1.25. The `gaze report` step should only run on Go 1.24 to avoid doubling the analysis time. The test step needs to emit a coverprofile that `gaze report --coverprofile` can reuse.

## Goals / Non-Goals

### Goals
- Run `gaze report --format=text --coverprofile=coverage.out` on Go 1.24 in the unit+integration job
- Enforce `--max-crapload` and `--max-gaze-crapload` thresholds
- Reuse the coverage profile from the test step (no second `go test` run)
- Keep the report output visible in CI logs (text format, no AI adapter)

### Non-Goals
- Running an AI-formatted report (no `--ai` flag — that requires API keys and is for downstream projects)
- Adding `gaze report` to the E2E suite
- Adding `gaze report` to Go 1.25

## Decisions

### D1: Conditional step using `if: matrix.go-version == '1.24'`

GitHub Actions matrix jobs share the same `steps` list. To run `gaze report` only on Go 1.24, use an `if` condition on the step. This is simpler than splitting the matrix into separate jobs.

### D2: Emit coverprofile from the test step

Change the test step to `go test -race -count=1 -short -coverprofile=coverage.out ./...` on Go 1.24 (via `if` condition on a separate test step, or by conditionally adding the flag). The simplest approach: add a second "Test with coverage" step conditional on Go 1.24, and keep the existing "Test" step conditional on Go 1.25.

Actually, simpler: change the existing test step to always emit `-coverprofile=coverage.out`. The coverprofile is a small file and `go test` supports it. The Go 1.25 run will also emit it but it won't be consumed.

### D3: Threshold values

Use the same thresholds as the `get-out` CI for consistency: `--max-crapload=16 --max-gaze-crapload=5`. These can be adjusted later as the codebase improves.

### D4: Text format, no AI adapter

Use `--format=text` (the default). The text report goes to stdout and is visible in CI logs. No `--ai` flag — this avoids requiring `OPENCODE_API_KEY` or any AI CLI in gaze's own CI.

## Risks / Trade-offs

- **CI time increase**: `gaze report` with `--coverprofile` skips the internal `go test` run but still runs the 4-step analysis pipeline (CRAP + quality + classify + docscan). Expected: 2-5 minutes additional on Go 1.24.
- **Coverprofile compatibility**: `go test -coverprofile=coverage.out ./...` writes a single merged profile. This is the standard format `gaze report` expects.
