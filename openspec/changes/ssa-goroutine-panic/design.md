## Context

Spec 021 added `safeSSABuild` wrappers with `recover()` around `prog.Build()` in both `BuildSSA` (analysis) and `BuildTestSSA` (quality). The assumption (research.md R2) was that `prog.Build()` is synchronous — all work runs on the calling goroutine.

This assumption is wrong. The `ssa.Program.Build()` implementation in `golang.org/x/tools@v0.43.0` (`builder.go:3144-3159`) spawns one goroutine per package by default:

```go
func (prog *Program) Build() {
    var wg sync.WaitGroup
    for _, p := range prog.packages {
        if prog.mode&BuildSerially != 0 {
            p.Build()
        } else {
            wg.Add(1)
            go func(p *Package) { p.Build(); wg.Done() }(p)
        }
    }
    wg.Wait()
}
```

The `go/types.NewSignatureType` panic with `go-json-experiment/json` generic variadic parameters happens inside a child goroutine spawned by `Build()`. Go's `recover()` is goroutine-scoped, so the `safeSSABuild` wrapper in the calling goroutine never catches it.

## Goals / Non-Goals

### Goals
- Make the existing `recover()` guards actually catch panics from `prog.Build()`
- Minimal change — fix the root cause without architectural rework
- Preserve all existing tests, degradation paths, and API surfaces

### Non-Goals
- Rewriting `safeSSABuild` to use goroutine+channel isolation (unnecessary with `BuildSerially`)
- Per-package `pkg.Build()` iteration with individual recovery (unnecessary complexity)
- Fixing the upstream `x/tools` or `go/types` bug
- Performance optimization of SSA build parallelism

## Decisions

### D1: Add `ssa.BuildSerially` to the builder mode flags

The `ssa.Program` supports a `BuildSerially` mode flag that forces `Build()` to call `p.Build()` directly instead of spawning goroutines. When this flag is set, all SSA construction runs on the calling goroutine, making `recover()` effective.

Change both call sites from:

```go
ssautil.AllPackages([]*packages.Package{pkg}, ssa.InstantiateGenerics)
```

to:

```go
ssautil.AllPackages([]*packages.Package{pkg}, ssa.InstantiateGenerics|ssa.BuildSerially)
```

**Rationale**: This is a one-line fix at each call site that solves the root cause. The `BuildSerially` flag is a documented, supported feature of `x/tools/go/ssa` designed for exactly this kind of scenario (controlling goroutine behavior during build). No architectural changes, no new dependencies, no API surface changes.

### D2: Accept the performance trade-off

Serial build loses parallelism across transitive dependencies. For gaze's use case this is acceptable:
- `ssautil.AllPackages` receives a single input package
- The parallelism was across that package's transitive dependency graph
- Correctness (not crashing) is a hard requirement; build speed is a soft preference
- SSA build time is a small fraction of total analysis time (dominated by `go/packages.Load`)
- The existing `safeSSABuild` tests remain valid — they test the `recover()` pattern which still works correctly with serial build

### D3: Update spec 021 research assumption

The research document at `specs/021-ssa-panic-recovery/research.md` R2 states `prog.Build()` is synchronous. This is incorrect and should be corrected with a note documenting the actual behavior and the `BuildSerially` fix.

## Risks / Trade-offs

### R1: SSA build is slower with BuildSerially

Serial build means packages are built one at a time instead of in parallel. For large dependency graphs, this could increase SSA construction time. Mitigation: gaze analyzes one package at a time, so the dependency graph per `Build()` call is bounded. Profiling can be done post-merge if users report performance regression.

### R2: Future x/tools versions may change Build() internals

If a future `x/tools` version removes goroutine spawning from `Build()` or changes the `BuildSerially` semantics, the flag would become a no-op (harmless). If they add new goroutine-spawning paths that bypass the `BuildSerially` check, the panic could resurface. Mitigation: the `BuildSerially` flag is a public API contract — changing its semantics would be a breaking change in `x/tools`.
