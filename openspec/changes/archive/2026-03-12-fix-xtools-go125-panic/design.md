## Context

`gaze report` uses `golang.org/x/tools/go/ssa` to build an SSA representation
of the target codebase for quality analysis. Under Go 1.25, the SSA builder in
`x/tools@v0.42.0` panics when it encounters a generic variadic parameter with a
named type (e.g. `jsontext.Value` from `go-json-experiment/json`). This affects
any project whose dependency graph includes that package — increasingly common as
`golang.org/x/net@v0.51.0+` pulls it in transitively.

The fix is purely a dependency version bump: `x/tools@v0.43.0` (released
2026-03-12) resolves the SSA builder panic. No gaze production source changes are
required.

A secondary gap is that gaze's CI only tests against Go 1.24, leaving Go 1.25
regressions undetected until user reports.

## Goals / Non-Goals

### Goals
- Resolve the `gaze report` SSA panic under Go 1.25 by upgrading `x/tools`.
- Prevent future toolchain-version regressions by adding Go 1.25 to the CI matrix.
- Keep the fix minimal: dependency bump + CI config only.

### Non-Goals
- Patching the SSA builder directly — that belongs upstream in `x/tools`.
- Backporting a fix for Go 1.24 users (the panic only manifests on 1.25).
- Adding new gaze features or changing any analysis behaviour.
- Supporting Go versions older than 1.24 (unchanged minimum).

## Decisions

### D1 — Upgrade to `x/tools@v0.43.0` (not a patch to `v0.42.0`)

`v0.43.0` is the first release that contains the upstream fix for the
`subst.go:559` panic. Vendoring a patch to `v0.42.0` would be fragile and
create a private fork. Upgrading to the published release is the correct path
and aligns with the project's no-private-fork convention.

**Trade-off**: `x/tools@v0.43.0` requires `go 1.25.0`, raising gaze's declared
minimum Go version from 1.24.2 to 1.25.0. Users building gaze from source on
Go 1.24 will need to upgrade. Pre-built binaries distributed via GoReleaser are
unaffected (self-contained). This is an acceptable trade-off: Go 1.25 has been
out since the panic was reported, and requiring the latest stable toolchain is
standard practice.

### D2 — CI matrix uses `"1.24"` and `"1.25"` (floating patch, not pinned)

Using floating minor-version strings (e.g. `"1.25"`) in `actions/setup-go`
automatically picks up security patch releases within the minor version. This is
the approach used by most Go projects and avoids manual patch-version bumps in
CI config. Both jobs (unit+integration and e2e) receive the same matrix so
coverage is symmetric.

### D3 — No code-level recovery / panic guard in gaze

An alternative would be to wrap the SSA build in a `recover()` and degrade
gracefully when it panics. This is rejected: it would mask the root cause, return
incomplete quality results silently, and violate the Accuracy principle. The
correct fix is to eliminate the panic at its source (upstream).

### D4 — `go.mod` minimum bumped to `1.25.0` rather than keeping `1.24.2`

`go mod tidy` automatically updates the `go` directive when a dependency requires
a higher version. This is the correct and expected behaviour. We document the
change explicitly rather than fighting the toolchain.

## Risks / Trade-offs

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| `x/tools@v0.43.0` introduces a different regression | Low | Existing test suite (`-race -count=1 -short ./...`) validates the upgrade before merge. CI matrix will catch it on both Go versions. |
| Users on Go 1.24 cannot build from source | Medium | Pre-built binaries unaffected. Documented in proposal. Go 1.25 is available on all major platforms. |
| CI cost doubles for two test jobs | Low | Both jobs are fast (unit+integration ~2min; e2e ~15min). Doubling is acceptable given the regression-prevention value. |
| Future `x/tools` upgrades require further `go` version bumps | Low | Tracked by routine dependency maintenance; no special risk here. |
<!-- scaffolded by unbound vdev -->
