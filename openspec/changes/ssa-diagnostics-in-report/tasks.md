## 1. Data Model

- [x] 1.1 Add `SSADegradedPackages []string` with JSON tag `ssa_degraded_packages,omitempty` to `PackageSummary` in `internal/taxonomy/types.go`.
- [x] 1.2 Add `ssa_degraded_packages` as an optional string array property to `QualitySchema` in `internal/report/schema.go`.
- [x] 1.3 Add `SSADegraded bool` and `SSADegradedPackages []string` fields to `ReportSummary` in `internal/aireport/payload.go`.

## 2. Quality Pipeline

- [x] 2.1 In `quality.Assess` (`internal/quality/quality.go`), set `summary.SSADegradedPackages = []string{testPkg.PkgPath}` when SSA is degraded.
- [x] 2.2 Change `runQualityForPackage` return signature from `([]taxonomy.QualityReport, bool)` to `([]taxonomy.QualityReport, string)` — return the degraded package path (empty if not degraded).
- [x] 2.3 Update `runQualityStep` to collect degraded package paths into a `[]string` and set both `summary.SSADegraded` and `summary.SSADegradedPackages` on the aggregate `PackageSummary`.
- [x] 2.4 Propagate `SSADegraded` and `SSADegradedPackages` from the quality step result to `ReportSummary` in `runProductionPipeline`.

## 3. Text Report

- [x] 3.1 Add a diagnostics section to the quality text report (`internal/quality/report.go`) that lists degraded packages when `SSADegraded` is true.

## 4. Tests

- [x] 4.1 Update `TestAssess_SSADegraded` to verify `SSADegradedPackages` contains the package path.
- [x] 4.2 Update `TestRunProductionPipeline_AllStepsSucceed` to verify `SSADegradedPackages` is nil on the quality summary.
- [x] 4.3 Add schema validation test for `ssa_degraded_packages` field.
- [x] 4.4 Run tests for affected packages — all pass.

## 5. Documentation

- [x] 5.1 Update `AGENTS.md` Recent Changes.
- [x] 5.2 Run `go build ./...` and `go vet ./...` — clean.
