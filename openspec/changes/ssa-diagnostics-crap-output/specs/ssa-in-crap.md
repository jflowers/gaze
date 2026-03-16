## ADDED Requirements

### Requirement: SSADegradedPackages on crap.Summary

`crap.Summary` MUST include `SSADegradedPackages []string` listing packages where SSA failed.

## MODIFIED Requirements

### Requirement: buildContractCoverageFunc tracks degradation

`buildContractCoverageFunc` MUST track which packages had SSA failures and pass them through to `crap.Options` so they appear in the CRAP summary.

### Requirement: Quality text report shows SSA diagnostics

When SSA degraded packages exist, the quality text report MUST display an SSA diagnostics section.
