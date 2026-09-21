# Delta Spec: CLI and Report Pipeline Integration

## ADDED Requirements

### Requirement: `gaze docscan` Analyzer Flags

The `gaze docscan` command MUST accept `--analyzer` and `--language` flags consistent with the existing `gaze crap`, `gaze quality`, and `gaze report` commands.

#### Scenario: Explicit analyzer flag
- **GIVEN** a user runs `gaze docscan --analyzer ./snake-eyes ./...`
- **WHEN** the command executes
- **THEN** gaze MUST spawn `./snake-eyes` (with `defer session.Close()` for cleanup), perform the initialize handshake, call `analyze` (and `doc_coverage` if supported via a new `Session.DocCoverage` method), compute API documentation coverage, and include the results in the output. If `doc_coverage` call fails at runtime despite the capability being announced, gaze MUST fall back to heuristic coverage and log a warning. The `doc_coverage` call MUST use `AnalysisTimeout` (consistent with `analyze`).

#### Scenario: Auto-discovered analyzer
- **GIVEN** a `gaze-analyzer-python` binary is on the user's PATH
- **AND** the user runs `gaze docscan --language python ./...`
- **WHEN** the command executes
- **THEN** gaze MUST discover and use `gaze-analyzer-python` via the three-tier discovery mechanism

#### Scenario: doc_coverage runtime failure fallback
- **GIVEN** an analyzer that announces `DocCoverage: true` in capabilities
- **AND** the analyzer returns a JSON-RPC error when `doc_coverage` is called
- **WHEN** `gaze docscan --analyzer ./analyzer ./...` is run
- **THEN** gaze MUST fall back to heuristic coverage using `analyze` output
- **AND** the JSON output MUST have `"source": "heuristic"` in `api_coverage`
- **AND** a warning MUST be logged to stderr

#### Scenario: No analyzer available
- **GIVEN** no `--analyzer` flag, no `.gaze.yaml` analyzer config, and no `gaze-analyzer-*` binary on PATH
- **WHEN** the user runs `gaze docscan ./...`
- **THEN** the command MUST produce the existing `[]DocumentFile` output with `api_coverage: null` (no error, graceful fallback)

### Requirement: Docscan JSON Output Structure Change

The `gaze docscan` JSON output MUST change from a bare `[]DocumentFile` array to a structured object:

```json
{
  "documents": [],
  "api_coverage": null
}
```

- `documents` MUST contain the existing `[]DocumentFile` array
- `api_coverage` MUST contain the `APICoverageReport` when an analyzer is available, or `null` otherwise

#### Scenario: JSON output with analyzer
- **GIVEN** an analyzer that reports 10 public functions, 8 documented
- **WHEN** `gaze docscan --analyzer ./analyzer ./...` is run
- **THEN** the JSON output MUST have a `documents` array AND an `api_coverage` object with `total_symbols: 10`, `documented_symbols: 8`, `coverage_percent: 80.0`

#### Scenario: JSON output without analyzer
- **GIVEN** no analyzer is available
- **WHEN** `gaze docscan ./...` is run
- **THEN** the JSON output MUST have a `documents` array AND `api_coverage: null`

### Requirement: Report Pipeline Docscan Step Extension

The `RunDocscanStep` function in `internal/aireport/runner_steps.go` MUST accept an optional analyzer session parameter and a `context.Context` for computing API documentation coverage alongside the existing Markdown scan. The Go-native pipeline invokes it with a nil session (heuristic-only); the external-analyzer report path invokes it directly with the session it already holds.

#### Scenario: Report pipeline with analyzer session
- **GIVEN** a `gaze report --analyzer ./snake-eyes ./...` invocation
- **AND** the analyzer session is already initialized for the CRAP step
- **WHEN** `RunDocscanStep` executes (via `runExternalReportCRAP`)
- **THEN** the step MUST reuse the existing session (not spawn a second analyzer), call `analyze` (cached) and optionally `doc_coverage`, and include `api_coverage` in the docscan JSON output

#### Scenario: Report pipeline without analyzer
- **GIVEN** a `gaze report ./...` invocation with no analyzer
- **WHEN** `RunDocscanStep` executes
- **THEN** the step MUST produce the existing `[]DocumentFile` output wrapped in `{"documents": [...], "api_coverage": null}`

### Requirement: Docscan Output Type

A single shared `DocscanEnvelope` struct MUST be defined in `internal/docscan/apidoc/types.go` and used directly by both consumers (`cmd/gaze` and `internal/aireport`). This avoids duplicating the envelope in each layer while keeping the type out of `internal/docscan` (which would create a circular import, since `apidoc.Analyze` accepts `[]docscan.DocumentFile`):

```go
type DocscanEnvelope struct {
    Documents   []docscan.DocumentFile `json:"documents"`
    APICoverage *APICoverageReport     `json:"api_coverage"`
}
```

#### Scenario: Marshaling with coverage
- **GIVEN** a `DocscanEnvelope` with 3 documents and a non-nil `APICoverageReport`
- **WHEN** marshaled to JSON
- **THEN** the output MUST contain both `documents` and `api_coverage` keys

#### Scenario: Marshaling without coverage
- **GIVEN** a `DocscanEnvelope` with 3 documents and nil `APICoverage`
- **WHEN** marshaled to JSON
- **THEN** the output MUST contain `documents` array and `"api_coverage": null`

## MODIFIED Requirements

### Requirement: `docscan.ScanOptions` — No Modification

`ScanOptions` is NOT modified. It retains its existing fields (`Config *config.GazeConfig`, `PackageDir string`). The caller is responsible for obtaining analyzer data and calling `apidoc.Analyze` separately after calling `Scan`.
### Requirement: `RunDocscanStep` Signature

The docscan step is exported as:

```go
func RunDocscanStep(ctx context.Context, moduleDir string, sess *adapter.Session, stderr io.Writer) (json.RawMessage, error)
```

When `sess` is nil, behavior is the Markdown scan only (no API coverage), wrapped in the shared `DocscanEnvelope` structure. When non-nil, the function MUST additionally compute API coverage. The `context.Context` enables caller cancellation/timeout control for the analyzer invocation.

### Requirement: `pipelineStepFuncs.docscanStep` Type Change

The Go-native pipeline's injected step type MUST be:

```go
docscanStep func(context.Context, string, io.Writer) (json.RawMessage, error)
```

The production default wraps `RunDocscanStep` with a nil session (the Go-native pipeline has no external analyzer). The external-analyzer report path calls `RunDocscanStep` directly with its session rather than flowing through `runProductionPipeline`.
## REMOVED Requirements

_(None.)_
