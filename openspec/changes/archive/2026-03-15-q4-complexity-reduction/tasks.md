## 1. scaffold.Run (complexity 22 -> 11)

- [x] 1.1 Extract `applyDefaults(opts *Options) error` from `scaffold.Run`. Handles `TargetDir` defaulting via `os.Getwd`, `Version` defaulting, `Stdout` defaulting. Update `Run` to call it.
- [x] 1.2 Extract `handleToolOwnedFile(outPath string, content []byte) (string, error)` from the WalkDir callback. Handles read-existing, compare, write-if-different logic. Returns action string.
- [x] 1.3 Extract `writeNewFile(outPath string, content []byte, exists bool) (string, error)` from the WalkDir callback. Handles `os.MkdirAll` + `os.WriteFile` + action label. Also extracted `processAssetFile` for the full per-file logic.
- [x] 1.4 Run existing scaffold tests: `go test -race -count=1 ./internal/scaffold/`
- [x] 1.5 Verify complexity: `gocyclo internal/scaffold/scaffold.go | grep Run` — 22 -> 11

## 2. crap.WriteText (complexity 18 -> 2)

- [x] 2.1 Extract `writeScoreTable(w io.Writer, sorted []Score, threshold float64) error` from `WriteText`. Handles table construction, row building, marker logic, style function.
- [x] 2.2 Extract `writeSummarySection(w io.Writer, summary Summary) error` from `WriteText`. Handles basic stats + CRAPload + GazeCRAP optional rendering.
- [x] 2.3 Extract `writeQuadrantSection(w io.Writer, counts map[Quadrant]int) error` from `WriteText`. Handles quadrant breakdown.
- [x] 2.4 Extract `writeWorstSection(w io.Writer, worst []Score, threshold float64) error` from `WriteText`. Handles worst offenders table.
- [x] 2.5 Run existing crap tests: `go test -race -count=1 ./internal/crap/`
- [x] 2.6 Verify complexity: `gocyclo internal/crap/report.go | grep WriteText` — 18 -> 2

## 3. classify.ComputeScore (complexity 16 -> 9)

- [x] 3.1 Extract `accumulateSignals(signals []taxonomy.Signal) (score int, hasPositive, hasNegative bool)` from `ComputeScore`. Handles weight summation loop with zero/empty filtering.
- [x] 3.2 Extract `classifyLabel(score, contractualThreshold, incidentalThreshold int) (taxonomy.ClassificationLabel, string)` from `ComputeScore`. Handles threshold comparison -> label + reasoning.
- [x] 3.3 Run existing classify tests: `go test -race -count=1 ./internal/classify/`
- [x] 3.4 Verify complexity: `gocyclo internal/classify/score.go | grep ComputeScore` — 16 -> 9

## 4. aireport.Run (complexity 16 -> 5)

- [x] 4.1 Extract `validateRunnerOpts(opts *RunnerOptions) error` from `Run`. Handles format defaulting, nil adapter check, binary validation.
- [x] 4.2 Extract `runJSONPath(payload *ReportPayload, opts RunnerOptions) error` from `Run`. Handles JSON encode to stdout + threshold evaluation.
- [x] 4.3 Extract `runTextPath(payload *ReportPayload, opts RunnerOptions) error` from `Run`. Handles marshal, timeout, adapter invocation, empty check, stdout write, step summary, thresholds.
- [x] 4.4 Run existing aireport tests: `go test -race -count=1 -short ./internal/aireport/`
- [x] 4.5 Verify complexity: `gocyclo internal/aireport/runner.go | grep 'aireport.Run '` — 16 -> 5

## 5. OllamaAdapter.Format (complexity 16 -> 12)

- [x] 5.1 Extract `resolveOllamaHost(configHost string) (string, error)` as a package-level unexported function. Handles config -> env -> default fallback, URL validation (scheme + host), returns full `/api/generate` URL.
- [x] 5.2 Run existing ollama tests: `go test -race -count=1 -short ./internal/aireport/ -run Ollama`
- [x] 5.3 Verify complexity: `gocyclo internal/aireport/adapter_ollama.go | grep Format` — 16 -> 12

## 6. Documentation & Verification

- [x] 6.1 Update `AGENTS.md` Recent Changes with a summary of this change.
- [x] 6.2 Run full test suite: `go test -race -count=1 -short ./...` — all 12 packages pass.
- [x] 6.3 Run `go build ./...` and `go vet ./...` — clean.
- [x] 6.4 Verify all 5 functions complexity reduced: Run 22→11, WriteText 18→2, ComputeScore 16→9, Run 16→5, Format 16→12.
