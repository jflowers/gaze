## Why

When SSA construction fails for some packages, gaze produces metrics based on partial data without telling users which packages failed or how much data is missing. The `SSADegraded` boolean on `PackageSummary` indicates that results are partial, but:

1. The per-package failure identity is lost — only an aggregate boolean is stored
2. Users can't tell "3/8 packages had SSA failures" — there's no count
3. SSA degradation is buried inside the quality JSON blob — not surfaced at the report payload level where the AI adapter and threshold evaluation can see it

This causes users to misinterpret partial metrics as complete ones, leading to incorrect conclusions about codebase quality (issue #46).

## What Changes

1. Add `SSADegradedPackages []string` to `PackageSummary` to track which packages had SSA failures.
2. Add `SSADegraded bool` and `SSADegradedPackages []string` to `ReportSummary` (the aireport-level summary) so degradation is visible at the payload level.
3. Update the quality text report to include an SSA diagnostics section when packages are degraded.

## Capabilities

### New Capabilities
- `SSADegradedPackages` field on `PackageSummary`: List of package paths where SSA construction failed. Machine-readable for CI and AI consumers.
- `SSADegraded` and `SSADegradedPackages` on `ReportSummary`: Propagated to the aireport payload level so the AI adapter sees degradation in the top-level summary, not buried in quality JSON.
- SSA diagnostics section in quality text report: "SSA construction failed for 3/8 packages: [list]"

### Modified Capabilities
- `runQualityStep`: Tracks degraded package names (not just boolean).
- `runQualityForPackage`: Returns package path when degraded.
- Quality text report: Adds diagnostics section when `SSADegraded` is true.

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/taxonomy/types.go` | Add `SSADegradedPackages []string` to `PackageSummary` |
| `internal/report/schema.go` | Add `ssa_degraded_packages` to `QualitySchema` |
| `internal/aireport/payload.go` | Add `SSADegraded` and `SSADegradedPackages` to `ReportSummary` |
| `internal/aireport/runner_steps.go` | Track degraded package names in `runQualityStep` and `runQualityForPackage` |
| `internal/quality/quality.go` | Set `SSADegradedPackages` on `PackageSummary` in `Assess` |
| `internal/quality/report.go` | Add diagnostics section to `WriteText` |
| Tests | New tests for degraded package tracking and text report display |
| `AGENTS.md` | Update Recent Changes |

## Constitution Alignment

### I. Accuracy — PASS
The change improves accuracy by surfacing which data is missing, preventing users from treating partial metrics as complete.

### II. Minimal Assumptions — PASS
No new user-facing requirements. Additive fields with omitempty.

### III. Actionable Output — PASS
The diagnostics section tells users exactly which packages failed and how many, enabling targeted investigation.

### IV. Testability — PASS
Per-package tracking is testable via the existing `BuildSSAFunc` injection point. The diagnostics section is testable via text output assertions.
