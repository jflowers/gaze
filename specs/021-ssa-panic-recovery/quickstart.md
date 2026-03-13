# Quickstart: SSA Panic Recovery

**Feature**: 021-ssa-panic-recovery
**Date**: 2026-03-13

## What This Changes

Two functions in gaze's analysis pipeline (`BuildSSA` and `BuildTestSSA`) call
`prog.Build()` from `golang.org/x/tools/go/ssa`. Under Go 1.25, this call can
panic for packages that use certain generic variadic type patterns (e.g.,
`go-json-experiment/json`). This feature adds `recover()` guards so the panic
is caught and converted to a graceful skip.

## Files to Modify

1. **`internal/analysis/mutation.go`** — Add `recover()` guard to `BuildSSA()`
2. **`internal/quality/pairing.go`** — Add `recover()` guard to `BuildTestSSA()`

## Implementation Pattern

```text
func BuildSSA(pkg *packages.Package) (ssaPkg *ssa.Package) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("warning: SSA build skipped for %s: internal panic recovered", pkg.PkgPath)
            log.Printf("debug: SSA panic value: %v", r)
            ssaPkg = nil
        }
    }()
    // ... existing code unchanged ...
}
```

Same pattern for `BuildTestSSA`, setting `err` in the recovery path instead.

## Testing

Add tests to `mutation_test.go` and `pairing_test.go` that:
1. Verify the `safeSSABuild` helper returns nil for non-panicking functions
2. Verify the helper catches panics and returns the panic value
3. Verify `BuildSSA` and `BuildTestSSA` return nil/error (not crash) when
   `prog.Build()` would panic

## Verification

```bash
# Must all pass identically to pre-change
go build ./...
go test -race -count=1 -short ./...
```

## No-Op Verification

The `recover()` + `defer` pattern has zero cost in the non-panic path (Go
runtime only allocates recovery state when a panic actually occurs). Existing
benchmarks in `internal/analysis/bench_test.go` validate this.
