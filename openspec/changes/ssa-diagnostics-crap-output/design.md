## Context

PR #54 added `SSADegradedPackages` to the quality pipeline (`PackageSummary`, `ReportSummary`). But the CRAP pipeline (`crap.Summary`) has no SSA awareness. When `gaze crap --format=json` is run, the output doesn't indicate which packages had SSA failures or how many functions have quality data.

## Approach

Add `SSADegradedPackages []string` to `crap.Summary`. Populate it from the quality pipeline during `buildContractCoverageFunc` by tracking which packages return degraded results from `analyzePackageCoverage`. Display in the text report when non-empty.
