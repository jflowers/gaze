## Tasks

- [x] 1. Add `SSADegradedPackages []string` with JSON tag `ssa_degraded_packages,omitempty` to `crap.Summary`.
- [x] 2. Update `analyzePackageCoverage` in `cmd/gaze/main.go` to return degraded package path alongside reports.
- [x] 3. Update `buildContractCoverageFunc` to collect degraded packages and return alongside the coverage function. Add `SSADegradedPackages` field to `crap.Options`.
- [x] 4. Update `buildSummary` to propagate `SSADegradedPackages` from `Options` to `Summary`.
- [x] 5. Add `writeSSADiagnostics` to display SSA diagnostics in CRAP text report when packages are degraded.
- [x] 6. Update AGENTS.md Recent Changes.
- [x] 7. Run `go build ./...`, `go vet ./...`, and tests. Fix hardcoded coverage profile line numbers.
