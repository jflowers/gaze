# Tasks: Redact Env Values from Adapter Stderr

## 1. Redaction helper (internal/aireport)

- [ ] 1.1 Add `internal/aireport/redact.go` with the redaction
  primitives from design D2/D3: an extensible case-insensitive
  sensitive-name marker slice (`KEY`, `TOKEN`, `SECRET`,
  `CONFIG_CONTENT`), a length-threshold constant (32), a minimum
  replacement-length constant (4), and the pure helper
  `redactEnvValues(s string, environ []string) string` that replaces
  sensitive inherited values with `[REDACTED:<VAR>]` markers —
  skipping empty and sub-minimum-length values, deduplicating by value
  (first name in environ order wins), and replacing longest values
  first. Include the GoDoc note on the post-`Wait()` env-snapshot
  assumption (design D4).
- [ ] 1.2 [P] Add table-driven unit tests for `redactEnvValues` in
  `internal/aireport/redact_test.go` (package `aireport`) with a
  synthetic environ covering: name-marker redaction (including
  `OPENCODE_CONFIG_CONTENT` and lowercase var names), threshold
  redaction of oddly-named long values, no redaction of
  short/unmatched values (garble guard), empty-value guard,
  longest-first replacement of overlapping values, dedupe with
  first-name-wins, the `[REDACTED:<VAR>]` marker format, and
  non-redaction of text that merely resembles a marker.
- [ ] 1.3 Wire the scrub into `runSubprocess` (`adapter.go`) as the
  single choke point (design D1): scrub the stderr buffer once after
  `cmd.Wait()` returns and use the scrubbed string for both the
  non-zero-exit error embed and the success-path returned bytes, so
  redaction precedes the `maxAdapterStderrBytes` truncation on every
  path. Update the `runSubprocess` and `maxAdapterStderrBytes`
  GoDoc comments to describe the scrub.
- [ ] 1.4 Add subprocess-level integration tests (existing
  `subprocess_test.go` `sh`/`echo` patterns plus `t.Setenv`): (a)
  non-zero exit with a sensitive value echoed to stderr yields an
  error containing the marker and not the value; (b) exit-0
  empty-stdout run surfaces redacted stderr through
  `formatStderrSuffix`; (c) a sensitive value starting before and
  extending past byte 512 leaves no prefix of the value in the
  surfaced snippet (the marker itself may be cut by the byte cap —
  assert only that the secret is gone).

## 2. Documentation

- [ ] 2.1 [P] Add a defense-in-depth subsection to the Reusable
  Workflow section of `docs/guides/ci-integration.md` (design D5):
  gaze's automatic adapter-stderr scrubbing (what triggers it, the
  `[REDACTED:<VAR>]` marker format, the known limits — sub-threshold
  and transformed values), registering the bare API key as a second
  secret referenced in an unused env var of the caller's job (so
  GitHub masks verbatim occurrences anywhere in logs), and preferring
  a single-line strict-JSON secret value because GitHub's multiline
  secret masking is unreliable. Keep the pretty-printed config example
  for readability with an explicit note that the actual secret value
  should be single-line.

## 3. Verification

- [ ] 3.1 [P] `openspec validate redact-env-values-from-adapter-stderr`
  passes with no issues other than the expected advisory for the
  `originating_issue` metadata key (required by /uf.finale for PR-body
  issue threading; unrecognized — and ignored — by openspec 1.14.1).
  Any requirement, scenario, or artifact issue is a failure.
- [ ] 3.2 [P] CI parity per the CI Parity Gate (derive exact commands
  from `.github/workflows/test.yml` and
  `.github/workflows/mega-linter.yml`): at minimum `go build ./...`,
  `go test -race -count=1 -short ./internal/aireport/...`, the full
  `go test -race -count=1 -short ./...`, and `golangci-lint run` all
  pass.
- [ ] 3.3 [P] Documentation impact assessed per the Documentation
  Validation Gate: all new symbols are unexported with GoDoc comments;
  README.md and AGENTS.md reviewed with updates made only if
  implementation reveals a user-visible behavior worth documenting
  beyond task 2.1.
- [ ] 3.4 [P] Constitution alignment re-verified against the
  implemented code: the proposal's assessments hold — the redaction set
  derives solely from gaze's own inherited environment (Minimal
  Assumptions), markers keep errors diagnosable (Actionable Output),
  and the pure helper with injected environ is testable in isolation
  (Testability); no analysis behavior changed (Accuracy N/A).
- [ ] 3.5 Spec artifacts committed and pushed before implementation
  begins (Spec Commit Gate).
