## Context

The release pipeline lived inline in `.github/workflows/release.yml`:
three local validation gates, then preflight (complytime/org-infra
reusable_release_preflight.yml), GoReleaser (reusable_release_goreleaser.yml),
native macOS signing and notarization, and Homebrew cask publishing —
roughly 380 lines. jflowers/ship now owns that pipeline once for every
unbound-force Go CLI (unbound-force/unbound-force#428), so this repo's
workflow shrank to what the caller genuinely decides.

## Goals / Non-Goals

### Goals

- Keep the three offline, unskippable local gates owned by this repo
- Delegate the shared pipeline to ship, pinned to a full commit SHA
- Keep the workflow name `Release` (post-release-install-check.yml
  listens for it)
- Guard the caller contract in `TestReleaseWorkflow_CallsShip`
- Record the prerelease decision explicitly

### Non-Goals

- Changing what the local gates validate
- Changing the release artifacts (archives, checksums, signatures, cask)
- Owning org-infra's pin — that moves to ship so one bump moves every
  repo
- Asserting ship's internal pipeline from this repo

## Decisions

### D1: Delegate, don't fork

The caller keeps the three gates and calls ship's `go-release.yml`
pinned to `86af2e7b2d324a73f688431910f957ef7accebd4 # v0.2.1`. Everything
after the gates — settings, preflight, GoReleaser, sign-macos, cask
publishing — is ship's. Alternatives considered: forking ship's workflow
per repo (recreates the duplication the change removes), or calling
org-infra directly (every repo then re-pins org-infra independently).

### D2: Caller/callee test-coverage boundary

`TestReleaseWorkflow_CallsShip` asserts only what the caller decides:
the workflow name, the SHA pin, `needs:` gate ordering, the six
`MACOS_*` secret mappings, the `HOMEBREW_TAP_TOKEN` rename, the six
forwarded `with:` inputs, and the forbidden patterns. Ship's pipeline —
including the partial-signing-secrets fail-before-tagging behavior — is
tested in jflowers/ship, whose `scripts/test-steps.sh` and
`scripts/test-audit-cask.sh` exercise the cask steps via yq. This repo
documents that boundary rather than duplicating the callee's coverage
(Constitution IV: test the contract, not the implementation).

### D3: allow_prerelease omitted deliberately

The old inline pipeline passed `allow_prerelease: true` to org-infra's
preflight. Ship v0.2.1 declares `allow_prerelease` (boolean,
`default: true`) and forwards it to that same preflight, so omitting the
input is behavior-preserving — verified against ship's workflow file at
the pinned SHA. The decision is recorded in a comment at the `with:`
block rather than by forwarding `allow_prerelease: true` explicitly:
restating ship's default at every caller would drift silently if ship's
default changed, while the comment names the verified fact. Prerelease
tags stay valid at the local gates (`validate-workflow-floor.sh` accepts
canonical prerelease suffixes; only prerelease floors are rejected).

### D4: Secrets mapped by name, never inherited

`secrets: inherit` cannot cross into another owner's workflow —
inherited secrets would silently arrive empty and the release would ship
unsigned. The call maps each of the six `MACOS_*` secrets by name plus
`HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}` (this
repo's token name differs from ship's). Ship's settings job fails before
anything is tagged when the signing set is partial.
