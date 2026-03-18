## Why

The CRAP JSON output reports "262 functions" and "35.4% avg contract coverage" without indicating that contract coverage is based on only 82 of those 262 functions (the ones where SSA succeeded). Users see the summary and think 35.4% represents the whole project, when it actually represents only 31% of the functions. The SSA degradation data exists in the quality pipeline (PR #54) but isn't surfaced in the CRAP output.

## What Changes

1. Add `SSADegradedPackages []string` to `crap.Summary` so the CRAP JSON output indicates which packages had SSA failures.
2. Change `runQualityForPackage` to return the degraded package path so `buildContractCoverageFunc` can track it.
3. Display SSA diagnostics in the quality text report when packages are degraded.

## Impact

| File | Change |
|------|--------|
| `internal/taxonomy/types.go` | Add `SSADegradedPackages` to `PackageSummary` (already exists from PR #54) |
| `internal/aireport/payload.go` | `SSADegraded`/`SSADegradedPackages` on `ReportSummary` (already exists from PR #54) |
| `internal/aireport/runner_steps.go` | Track degraded package paths (already done in PR #54) |
| `cmd/gaze/main.go` | Track SSA-degraded packages in `buildContractCoverageFunc` and surface in quality text report |

## Constitution Alignment

All PASS. Improves Accuracy (Principle I) by preventing users from misinterpreting partial metrics as complete.
