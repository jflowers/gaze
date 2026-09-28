<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
-->

# Tasks: Fix macOS Signing Identity Secret

## 1. Prerequisites

- [x] 1.1 Confirm that an authorized repository administrator has configured
  the `MACOS_SIGN_IDENTITY` Actions secret with the current Apple Developer ID
  Application certificate label; do not add, reveal, or commit its value.

## 2. Release Signing Configuration

- [x] 2.1 Update `.github/workflows/release.yml` so
  `check-signing-secrets` requires all six secrets consumed by `sign-macos`,
  and map the signing identity directly from `secrets.MACOS_SIGN_IDENTITY`
  with no variable or hardcoded fallback.
- [x] 2.2 Extend the existing `TestReleaseWorkflow_MacOSSigningIdentity` in
  `cmd/gaze/release_workflow_test.go`. Read the workflow as configuration and
  preserve its existing gate-condition, routing, identity-mapping, and
  forbidden-fallback assertions. Isolate the `check-signing-secrets` section
  and add the missing per-secret assertions that all six environment mappings
  use `secrets.<NAME>`.
- [x] 2.3 [P] Update the macOS-signing prerequisite documentation in
  `README.md` to list the sixth secret and its certificate-label value.
  Preserve completed feature specs as point-in-time design artifacts.
- [x] 2.4 Update `check-signing-secrets` to emit a GitHub warning naming each
  missing secret before reporting `has_secrets=false`. Print constant secret
  names only; never print or interpolate values.
- [x] 2.5 After 2.2 and 2.4, extend
  `TestReleaseWorkflow_MacOSSigningIdentity` to assert all six
  `::warning::Signing configuration missing: <NAME>` lines individually.
  Assert that warning lines contain no `${{ ... }}` or shell-variable
  interpolation.

## 3. Verification

- [x] 3.1 Run `go build ./...`, `go test -race -count=1 -short -timeout 15m
  -coverprofile=coverage.out ./...`, and `go test -race -count=1 -run
  'TestRunSelfCheck' -timeout 30m ./cmd/gaze/...` to match the Test workflow's
  relevant build and test gates.
- [ ] 3.2 Run `golangci-lint run` and validate the changed workflow through
  the repository's MegaLinter-equivalent local tooling; fix all reported YAML
  or workflow syntax errors without changing CI gates.
- [ ] 3.3 Have a repository administrator confirm on the next permitted
  release that a complete secret set runs `sign-macos` and that removing the
  identity secret selects the unsigned-cask path. Do not create a test release
  solely for this change.
- [x] 3.4 Re-check the proposal's Constitution Alignment: secret-based
  configuration preserves autonomous collaboration and composability, the
  workflow state remains observable, and the added static test remains
  isolated from external services. Also verify the Gaze Constitution assessment
  for accuracy, minimal assumptions, actionable output, and testability.

<!-- spec-review: passed -->
