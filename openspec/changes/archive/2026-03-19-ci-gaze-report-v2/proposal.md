# CI Gaze Report v2

## Why

The `ci-gaze-report` branch added a `gaze report --ai=opencode` step to the unit+integration CI job, but it has two problems:

1. **Fork PR secret restriction**: GitHub Actions does not expose repository secrets to PRs from forks. The `OPENCODE_API_KEY` secret is unavailable during fork PR checks, causing the gaze report step to fail with "Model not found" or "empty output". Since gaze uses a fork workflow (PRs come from `jflowers/gaze` to `unbound-force/gaze`), every PR fails this step.

2. **Push-only viability**: The step works on push to `main` (secrets are available), but it shouldn't block PRs that can't possibly pass it.

The fix is to make the gaze report step run only on push to main (post-merge quality gate), not on PRs. This matches how the step is used — it gates the quality of merged code, not draft PRs. For PRs, the JSON-format report (no AI adapter needed) can run as a non-blocking check to catch threshold violations without requiring secrets.

## What Changes

Split the gaze report into two modes:
1. **PR mode**: `gaze report --format=json` with thresholds — no AI adapter, no secrets needed, runs on every PR for both Go versions. Catches CRAPload/GazeCRAPload regressions before merge.
2. **Push mode**: `gaze report --ai=opencode` with full AI formatting — runs only on push to main on Go 1.24. Produces the human-readable quality report for post-merge visibility.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `.github/workflows/test.yml`: Gaze quality report step split into PR-safe (JSON thresholds) and push-only (AI-formatted) modes

### Removed Capabilities
- None

## Impact

- `.github/workflows/test.yml` — modified

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

CI workflow configuration only.

### II. Composability First

**Assessment**: PASS

The JSON-format report requires no external dependencies (no npm, no opencode CLI). The AI-formatted report is isolated to the push-to-main path where secrets are available.

### III. Observable Quality

**Assessment**: PASS

Both PR and push paths enforce the same thresholds. The PR path catches regressions before merge; the push path adds human-readable reporting after merge.

### IV. Testability

**Assessment**: N/A

No Go code changes.
