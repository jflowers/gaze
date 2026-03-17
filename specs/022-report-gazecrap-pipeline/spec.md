# Feature Specification: GazeCRAP Data in Report Pipeline

**Feature Branch**: `022-report-gazecrap-pipeline`  
**Created**: 2026-03-17  
**Status**: Ready  
**Input**: User description: "Wire ContractCoverageFunc into gaze report pipeline so GazeCRAPload, quadrant distribution, and GazeCRAP scores appear in CI reports"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - CI Report Shows GazeCRAP Quadrant Distribution (Priority: P1)

A developer runs `gaze report ./... --ai=opencode --coverprofile=coverage.out` in CI. The resulting AI-formatted report includes a GazeCRAP Quadrant Distribution table showing how many functions fall into Q1 (Safe), Q2 (Complex But Tested), Q3 (Needs Tests), and Q4 (Dangerous). Today this section shows "N/A" because the report pipeline never computes GazeCRAP data.

**Why this priority**: GazeCRAP is the distinguishing metric of gaze — it measures contract coverage, not just line coverage. Without it in the CI report, the report is missing its most important signal. The `--max-gaze-crapload` threshold flag already exists but has no data to evaluate against.

**Independent Test**: Run `gaze report --format=json --coverprofile=<valid-profile> ./...` and verify the JSON output contains `gaze_crap` and `quadrant` fields per function, and `quadrant_counts` and `gaze_crapload` fields in the CRAP summary.

**Acceptance Scenarios**:

1. **Given** a project with a valid coverprofile and test files, **When** the user runs `gaze report --format=json --coverprofile=coverage.out ./...`, **Then** the JSON output contains `gaze_crap` and `quadrant` fields for each scored function.
2. **Given** a project with a valid coverprofile, **When** the user runs `gaze report --format=json --coverprofile=coverage.out ./...`, **Then** the CRAP summary contains `quadrant_counts` with counts for Q1, Q2, Q3, and Q4.
3. **Given** a project with a valid coverprofile, **When** the user runs `gaze report --format=json --coverprofile=coverage.out ./...`, **Then** the CRAP summary contains a numeric `gaze_crapload` field (not null or absent).

---

### User Story 2 - GazeCRAPload Threshold Enforcement Works (Priority: P2)

A developer sets `--max-gaze-crapload=5` on the `gaze report` command. When the analysis runs, the GazeCRAPload count (number of Q4 Dangerous functions) is computed and compared against the threshold. If GazeCRAPload exceeds the threshold, the command exits with a non-zero status and a clear failure message. Today, GazeCRAPload always evaluates as 0 because no GazeCRAP data is computed, so the threshold silently passes regardless of code quality.

**Why this priority**: CI quality gates are the primary consumer of `gaze report`. A threshold flag that silently passes defeats the purpose of automated governance. This story depends on Story 1 (GazeCRAP data must exist before thresholds can evaluate it).

**Independent Test**: Run `gaze report --coverprofile=coverage.out --max-gaze-crapload=0 ./...` against a project with at least one Q4 function and verify the command exits non-zero with a "GazeCRAPload: N/0 (FAIL)" message.

**Acceptance Scenarios**:

1. **Given** a project with Q4 functions and `--max-gaze-crapload=0`, **When** the user runs `gaze report`, **Then** the command exits non-zero and reports the GazeCRAPload count exceeds the threshold.
2. **Given** a project with zero Q4 functions and `--max-gaze-crapload=5`, **When** the user runs `gaze report`, **Then** the threshold passes and the GazeCRAPload line shows "0/5 (PASS)".

---

### User Story 3 - AI-Formatted Report Includes GazeCRAP Section (Priority: P3)

When `gaze report --ai=opencode` produces the AI-formatted text report, the AI receives GazeCRAP data in the JSON payload and renders the quadrant distribution table, GazeCRAPload summary line, and uses quadrant information to inform recommendations. Today the AI correctly reports "N/A" because the data is absent from the payload.

**Why this priority**: This is a presentation concern that follows automatically from Story 1. Once the JSON payload contains GazeCRAP data, the AI agent prompt already has instructions to render it. No prompt changes are needed — only the data pipeline fix.

**Independent Test**: Run `gaze report --ai=opencode --coverprofile=coverage.out ./...` and verify the output contains a "GazeCRAP Quadrant Distribution" table with actual counts (not "N/A").

**Acceptance Scenarios**:

1. **Given** a project with a valid coverprofile and `--ai=opencode`, **When** the user runs `gaze report`, **Then** the AI-formatted output includes a GazeCRAP Quadrant Distribution table with Q1/Q2/Q3/Q4 counts.
2. **Given** a project where all functions are Q1 Safe, **When** the user runs `gaze report --ai=opencode`, **Then** the GazeCRAPload summary line reports "0" with a conversational interpretation.

---

### Edge Cases

- What happens when the coverprofile is missing or invalid? The report should still produce CRAP scores (line-coverage-based) but GazeCRAP fields should be absent (contract coverage cannot be computed without quality analysis). The existing graceful degradation behavior should be preserved.
- What happens when quality analysis fails for some packages (SSA degradation)? GazeCRAP should still be computed for packages that succeed, with degraded packages excluded from the contract coverage callback. The `ssa_degraded_packages` field should reflect which packages were skipped.
- What happens when a package has no test files? Functions in that package should receive 0% contract coverage, placing them in Q3 or Q4 depending on complexity.
- What happens when `--coverprofile` is not provided? The report pipeline runs `go test -coverprofile` internally. The contract coverage callback should still be wired using the internally-generated coverage data.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `gaze report` pipeline MUST compute GazeCRAP scores (combining cyclomatic complexity with contract coverage) for every scored function when quality analysis succeeds.
- **FR-002**: The `gaze report` JSON output MUST include `gaze_crap` and `quadrant` fields for each function in the CRAP scores array when GazeCRAP data is available.
- **FR-003**: The `gaze report` JSON output MUST include `quadrant_counts` (Q1/Q2/Q3/Q4 counts) and `gaze_crapload` (count of Q4 functions) in the CRAP summary when GazeCRAP data is available.
- **FR-004**: The `--max-gaze-crapload` threshold MUST evaluate against the actual computed GazeCRAPload count, not a default of 0.
- **FR-005**: When quality analysis fails entirely (all packages degraded), the report MUST fall back to line-coverage-only CRAP scores with GazeCRAP fields absent, preserving existing graceful degradation behavior.
- **FR-006**: The contract coverage computation in the report pipeline MUST use the same logic as `gaze crap` (the `buildContractCoverageFunc` pattern in `cmd/gaze/main.go`), ensuring consistent scores between standalone `gaze crap` and `gaze report`.
- **FR-007**: The report pipeline MUST NOT require a `--coverprofile` flag to compute GazeCRAP — when no coverprofile is provided, the internally-generated coverage data MUST be used for both line coverage and contract coverage analysis.

### Key Entities

- **ContractCoverageFunc**: A callback function that takes a function name and package path, returns contract coverage percentage. Currently wired in `gaze crap` but missing from `gaze report`.
- **GazeCRAP Score**: A composite metric combining cyclomatic complexity with contract coverage (vs standard CRAP which uses line coverage). Produces quadrant assignments (Q1-Q4).
- **Quadrant Distribution**: Summary counts of functions in each GazeCRAP quadrant — the primary output consumers care about.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `gaze report --format=json --coverprofile=<valid> ./...` produces JSON containing non-null `gaze_crapload` and `quadrant_counts` fields in the CRAP summary for 100% of runs where quality analysis succeeds.
- **SC-002**: The `gaze_crapload` value from `gaze report` matches the `gaze_crapload` value from running `gaze crap` standalone with the same coverprofile and package pattern, within a tolerance of 0 (exact match).
- **SC-003**: The `--max-gaze-crapload` threshold correctly fails the CI pipeline when GazeCRAPload exceeds the limit, with a non-zero exit code and human-readable failure message.
- **SC-004**: The AI-formatted text report (via `--ai=opencode`) includes a GazeCRAP Quadrant Distribution table with actual counts instead of "N/A" for 100% of runs where quality analysis succeeds.

### Assumptions

- The quality analysis step in the report pipeline already runs successfully and produces per-package quality data. This feature wires that data into the CRAP scoring step — it does not add new analysis capabilities.
- The `buildContractCoverageFunc` pattern from `cmd/gaze/main.go` is the canonical implementation. The report pipeline should reuse or replicate this logic, not invent a new approach.
- The existing `--coverprofile` flag behavior (skip internal `go test` when provided) is preserved. Contract coverage analysis uses the same coverage data regardless of source.
