## ADDED Requirements

### Requirement: SSADegradedPackages on PackageSummary

`PackageSummary` MUST include `SSADegradedPackages []string` (JSON tag `ssa_degraded_packages,omitempty`) listing the package paths where SSA construction failed.

#### Scenario: Single package SSA failure
- **GIVEN** `Assess` is called for a package where `BuildTestSSA` fails
- **WHEN** the function returns
- **THEN** `PackageSummary.SSADegradedPackages` MUST contain the package path

#### Scenario: No SSA failure
- **GIVEN** `Assess` is called for a package where `BuildTestSSA` succeeds
- **WHEN** the function returns
- **THEN** `PackageSummary.SSADegradedPackages` MUST be nil

### Requirement: SSA degradation on ReportSummary

`ReportSummary` MUST include `SSADegraded bool` and `SSADegradedPackages []string` fields propagated from the quality step.

#### Scenario: Degradation propagated to ReportSummary
- **GIVEN** `runQualityStep` processes 3 packages and 1 has SSA failure
- **WHEN** the step completes
- **THEN** `ReportSummary.SSADegraded` MUST be true and `SSADegradedPackages` MUST contain the failed package path

### Requirement: Quality text report diagnostics

When `SSADegraded` is true, the quality text report MUST include a diagnostics section listing the degraded packages and noting that metrics are partial.

#### Scenario: Diagnostics shown in text output
- **GIVEN** a quality report with `SSADegraded: true` and 2 degraded packages
- **WHEN** `WriteText` is called
- **THEN** the output MUST contain the degraded package paths and a note about partial metrics

## MODIFIED Requirements

### Requirement: runQualityForPackage returns degraded package path

Previously: `runQualityForPackage` returned `([]taxonomy.QualityReport, bool)`.

`runQualityForPackage` MUST return the degraded package path as a string (empty if not degraded) instead of a boolean.

## REMOVED Requirements

None.
