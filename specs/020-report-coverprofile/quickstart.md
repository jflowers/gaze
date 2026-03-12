# Quickstart: --coverprofile for gaze report

**Feature**: 020-report-coverprofile
**Audience**: Developers integrating `gaze report` into an existing CI pipeline

---

## The Problem

Running `gaze report` in CI today triggers two full test suite executions:

1. Your existing `go test` step (with `-race`, `-count=1`, verbose output)
2. An internal `go test -short` run triggered by `gaze report` to generate a coverage profile

The second run is invisible, weaker (no race detector, skips short-guarded tests), and adds latency.

---

## The Solution

Pass your existing coverage profile to `gaze report` via `--coverprofile`:

```bash
# Step 1: Run tests once, generating the coverage profile you already need
go test -race -count=1 -coverprofile=coverage.out ./...

# Step 2: Use that profile — gaze skips the internal go test run
gaze report ./... --ai=claude --coverprofile=coverage.out
```

---

## GitHub Actions Example

```yaml
- name: Test
  run: go test -race -count=1 -coverprofile=coverage.out ./...

- name: Gaze Report
  run: gaze report ./... --ai=claude --coverprofile=coverage.out
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
```

With this setup, tests run **once** per CI job. The coverage data used for CRAP scores is the same high-quality profile produced by the race-detecting test run.

---

## Combining with Threshold Gates

```bash
go test -race -count=1 -coverprofile=coverage.out ./...

gaze report ./... \
  --ai=claude \
  --coverprofile=coverage.out \
  --max-crapload=20 \
  --min-contract-coverage=60
```

Threshold evaluation uses the CRAP scores derived from your supplied profile. If thresholds are breached, `gaze report` exits non-zero and CI fails.

---

## JSON-Only Mode (No AI)

The flag works for `--format=json` too — no AI adapter needed:

```bash
go test -race -count=1 -coverprofile=coverage.out ./...
gaze report ./... --format=json --coverprofile=coverage.out > report.json
```

---

## Error Cases

| Situation | Error message |
|-----------|---------------|
| Path does not exist | `--coverprofile "/path/coverage.out": stat ...: no such file or directory` |
| Path is a directory | `--coverprofile "/path/dir" is a directory, not a file` |
| File has invalid content | Error stored in JSON `errors.crap` field: `parsing coverage profile: <parser error>` (suffix from Go coverage parser; use `--format=json` to inspect) |

---

## What Stays the Same

- Omitting `--coverprofile` preserves the existing behavior exactly — `gaze report` generates coverage internally via `go test -short`.
- All other flags (`--ai`, `--model`, `--format`, threshold flags) are unaffected.
- The quality, classification, and docscan analysis steps are unaffected — only the CRAP step uses the coverage profile.
