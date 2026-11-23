<!--
  Retroactive record for PR #362 (unbound-force/unbound-force#428).
  Implementation predates this artifact; tasks are complete as written.
  The with:-input guard, the allow_prerelease decision, and the AGENTS.md
  CI/CD refresh landed during review-feedback triage — the caller-guard
  rewrite and explicit allow_prerelease pass in 01504be superseded the
  triage draft's omission-and-comment approach.
-->

## 1. Local Gates Retained

- [x] 1.1 Keep `validate-module-major` (tag major vs. go.mod module path) unchanged.
- [x] 1.2 Keep `validate-workflow-floor` (tag vs. minWorkflowTag floor; above-floor releases require the floor tag to exist and contain the reusable AI-report workflow) unchanged.
- [x] 1.3 Keep `validate-ai-report-default` (reusable workflow's gaze-version default must equal the tag) unchanged.

## 2. Delegation

- [x] 2.1 Replace the inline preflight/GoReleaser/sign-macos/cask jobs with a call to `jflowers/ship/.github/workflows/go-release.yml` pinned to a full commit SHA with its version comment.
- [x] 2.2 Forward `tag`, `ci_checks` (the three required check-run names), the three `skip_*` flags, and `tap: unbound-force/homebrew-tap`.
- [x] 2.3 Map the six `MACOS_*` secrets by name and `HOMEBREW_TAP_TOKEN` from `HOMEBREW_TAP_GITHUB_TOKEN`; keep the workflow name `Release`.
- [x] 2.4 Grant the call `contents: write`, `checks: read`, `id-token: write`.

## 3. Review-Feedback Fixes (PR #362 triage)

- [x] 3.1 Assert every forwarded `with:` input in `TestReleaseWorkflow_CallsShip` (exact-map assertion over the parsed workflow, both directions, so a dropped, mistyped, or added input fails).
- [x] 3.2 Record the `allow_prerelease` decision at the `with:` block: ship v0.2.1's default is verified `true`, and the call passes `allow_prerelease: true` explicitly with a comment naming that default, so rc tags keep working even if ship's default changes.

## 4. Verification

- [x] 4.1 `go build ./cmd/gaze` — zero errors.
- [x] 4.2 `go test -race -count=1 -short ./...` — all tests pass, including `TestReleaseWorkflow_CallsShip`.
- [x] 4.3 `go vet ./...` — clean.
- [x] 4.4 `golangci-lint run` — zero findings.

## 5. Documentation

- [x] 5.1 Refresh `AGENTS.md` CI/CD section item 3 for the ship delegation (landed in 01504be; also tracked as a filed follow-up issue from PR #362 review).
- [x] 5.2 No website issue — release pipeline internals; no user-facing CLI, flag, or workflow change.
