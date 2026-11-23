## Why

The release pipeline — preflight validation, GoReleaser with cosign
signatures and SBOMs, native macOS signing and notarization, and the
Homebrew cask push — was duplicated between this repo and the other
unbound-force Go tools. jflowers/ship now owns that pipeline once
(unbound-force/unbound-force#428); each consumer keeps only what it
genuinely decides locally.

This change is a retroactive record for PR #362 (implemented and reviewed
before the spec was written; recorded during review-feedback triage).
Fixes unbound-force/unbound-force#428.

## What Changes

### Release delegates to jflowers/ship

`release.yml` keeps three offline, unskippable local gates —
`validate-module-major`, `validate-workflow-floor`,
`validate-ai-report-default` — then calls
`jflowers/ship/.github/workflows/go-release.yml` pinned to a full commit
SHA. The workflow keeps the name `Release`; post-release-install-check.yml
runs when a `Release` run completes.

### Caller contract asserted in tests

`cmd/gaze/release_workflow_test.go` guards what the caller still decides:
the workflow name, the exact job set, the SHA-pinned call, gate ordering
(`needs:`), the six `MACOS_*` secret mappings, the `HOMEBREW_TAP_TOKEN`
rename, the seven forwarded `with:` inputs, and the forbidden patterns
(`vars.MACOS_SIGN_IDENTITY`, `Developer ID Application:`; the secrets
block must be a name mapping, which structurally rules out
`secrets: inherit`).

### Prerelease decision recorded

The previous inline pipeline passed `allow_prerelease: true` to
org-infra's preflight. Ship v0.2.1 declares the same input with
`default: true` (verified at the pinned SHA), and the call forwards
`allow_prerelease: true` explicitly, with a comment naming the verified
default — so rc tags (vX.Y.Z-rc.N) keep working even if ship's default
ever changes.

## Capabilities

### New Capabilities

- `release-delegation` — the release workflow gates locally, then
  delegates to jflowers/ship pinned to a commit SHA

### Modified Capabilities

- None

### Removed Capabilities

- None

## Impact

- **Modified files**: `.github/workflows/release.yml` (the inline
  preflight/GoReleaser/signing/cask pipeline replaced by three gates plus
  one pinned call), `cmd/gaze/release_workflow_test.go` (guard rewritten
  around the caller contract)
- **Risk**: Medium. The released artifacts are unchanged — the same
  org-infra reusable workflows now run through ship, which owns the
  org-infra pin. Signing remains all-or-nothing: a partial `MACOS_*`
  secret set fails in ship's settings job before anything is tagged.
- **Test impact**: `TestReleaseWorkflow_CallsShip` covers the caller
  contract; ship's pipeline is tested in jflowers/ship.

## Constitution Alignment

Assessed against the Gaze project constitution
(`.specify/memory/constitution.md`).

### I. Accuracy

**Assessment**: PASS

The caller guard asserts the exact delegation contract (pin, `with:`
inputs, secret mappings, gate ordering), and the release behavior is
verified unchanged against ship's declared `allow_prerelease` default.

### II. Minimal Assumptions

**Assessment**: PASS

The pipeline assumes only what it already assumed — org-infra's reusable
workflows — now reached through one shared indirection. The caller pins
ship by full commit SHA.

### III. Actionable Output

**Assessment**: N/A

No report output changes.

### IV. Testability

**Assessment**: PASS

The caller contract is tested as data (workflow-file assertions). The
callee pipeline is tested in jflowers/ship; the boundary is documented in
`design.md` D2 rather than duplicated here.
