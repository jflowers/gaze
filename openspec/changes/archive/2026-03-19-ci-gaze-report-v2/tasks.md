# Tasks: CI Gaze Report v2

## 1. Replace single step with two conditional steps

- [x] 1.1 In `.github/workflows/test.yml`, replace the current "Gaze quality report" step (lines 35-47) with a "Gaze threshold check" step that uses `--format=json`, has `if: matrix.go-version == '1.24'` (runs on PRs AND pushes), and redirects JSON output to `/dev/null`. Use the same thresholds: `--max-crapload=16 --max-gaze-crapload=5 --min-contract-coverage=8`.
- [x] 1.2 After the threshold check step, add a "Gaze quality report" step with `if: matrix.go-version == '1.24' && github.event_name == 'push'` that runs the full `--ai=opencode` report with `OPENCODE_API_KEY` env var. Use the same thresholds.
- [x] 1.3 Move the "Install OpenCode" step's `if` condition to `matrix.go-version == '1.24' && github.event_name == 'push'` so npm install only runs on push to main.

## 2. Clean up the old ci-gaze-report branch

- [x] 2.1 Close PR #67 (ci-gaze-report) with a comment explaining it's superseded by this change.
- [x] 2.2 Delete the `ci-gaze-report` branch from origin (jflowers/gaze).

## 3. Verification

- [x] 3.1 Verify the workflow YAML is valid (indentation, `if` syntax, step ordering).
- [ ] 3.2 Commit, push, and verify: the PR check runs the threshold step (JSON-only, no secrets needed) and skips the AI report step.
