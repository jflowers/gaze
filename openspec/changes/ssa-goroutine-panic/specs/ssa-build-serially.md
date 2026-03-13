## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: SSA builder mode includes BuildSerially

Previously: `BuildSSA` and `BuildTestSSA` used `ssa.InstantiateGenerics` as the sole builder mode flag.

Both `BuildSSA` and `BuildTestSSA` MUST use `ssa.InstantiateGenerics | ssa.BuildSerially` as the builder mode when calling `ssautil.AllPackages`. This ensures `prog.Build()` runs all SSA construction on the calling goroutine, making the existing `safeSSABuild` `recover()` guard effective against panics from any package in the dependency graph.

#### Scenario: Panic in transitive dependency during SSA build (analysis)
- **GIVEN** a loaded package whose transitive dependencies include a type that triggers a panic in `go/types.NewSignatureType` (e.g., `go-json-experiment/json` generic variadic parameters under Go 1.25)
- **WHEN** `BuildSSA` is called with that package
- **THEN** the panic MUST be caught by `safeSSABuild` and `BuildSSA` MUST return nil (not crash the process)

#### Scenario: Panic in transitive dependency during SSA build (quality)
- **GIVEN** a loaded test package whose transitive dependencies include a type that triggers a panic in `go/types.NewSignatureType`
- **WHEN** `BuildTestSSA` is called with that package
- **THEN** the panic MUST be caught by `safeSSABuild` and `BuildTestSSA` MUST return an error (not crash the process)

#### Scenario: SSA build succeeds with BuildSerially
- **GIVEN** a loaded package with no problematic types
- **WHEN** `BuildSSA` is called with that package
- **THEN** `BuildSSA` MUST return a valid `*ssa.Package` (BuildSerially does not break normal operation)

#### Scenario: SSA build succeeds with BuildSerially (quality)
- **GIVEN** a loaded test package with no problematic types
- **WHEN** `BuildTestSSA` is called with that package
- **THEN** `BuildTestSSA` MUST return a valid `*ssa.Program` and `*ssa.Package` with nil error

### Requirement: Spec 021 research assumption corrected

The research document `specs/021-ssa-panic-recovery/research.md` R2 SHOULD include a correction noting that `prog.Build()` spawns goroutines by default and `ssa.BuildSerially` is required for `recover()` to work correctly.

#### Scenario: Research document reflects actual behavior
- **GIVEN** the research document at `specs/021-ssa-panic-recovery/research.md`
- **WHEN** a developer reads section R2
- **THEN** the document SHOULD note that the original synchronous assumption was incorrect and `BuildSerially` was applied as a fix

## REMOVED Requirements

None.
