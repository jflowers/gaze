## Context

After the adapter-format-decomposition, crapload-analyze-decomposition, and pipeline-step-testing changes, 5 functions remain in Q4 Dangerous. All have cyclomatic complexity 16-22 and strong existing test coverage. The fix is mechanical decomposition — extract cohesive groups of branches into helper functions.

## Goals / Non-Goals

### Goals
- Reduce all 5 functions below complexity 10
- Move all 5 out of Q4 Dangerous
- Preserve all existing tests (zero regressions)
- Keep helpers unexported and co-located with the parent function

### Non-Goals
- Adding new tests (existing coverage is sufficient)
- Changing any public API signatures
- Modifying output formats or behavior
- Addressing Q3 functions (need contract assertions, separate concern)

## Decisions

### D1: scaffold.Run decomposition

Extract 3 helpers from the 77-line WalkDir callback:

1. **`applyDefaults(opts *Options) error`** — Defaults `TargetDir` (via `os.Getwd`), `Version`, and `Stdout`. Returns error only if `Getwd` fails. Moves 4 branches out.

2. **`handleToolOwnedFile(outPath string, content []byte) (string, error)`** — The tool-owned overwrite-on-diff block: reads existing file, compares with `bytes.Equal`, writes if different, returns action string ("updated" or "skipped (identical)"). Moves 5 branches out.

3. **`writeNewFile(outPath string, content []byte, exists bool) (string, error)`** — Creates parent directory via `os.MkdirAll`, writes file, returns action string ("overwritten" or "created"). Moves 4 branches out.

**Expected result**: `Run` complexity 22 → ~9.

### D2: crap.WriteText decomposition

Extract 4 section-writing helpers. Each section is independent — no data flows between them except through the `Summary` struct:

1. **`writeScoreTable(w io.Writer, sorted []Score, threshold float64) error`** — Table construction, row building, marker/style logic. Moves 6 branches out.
2. **`writeSummarySection(w io.Writer, summary Summary) error`** — Basic stats + CRAPload + GazeCRAP optional fields. Moves 4 branches out.
3. **`writeQuadrantSection(w io.Writer, counts map[Quadrant]int) error`** — Quadrant breakdown rendering. Moves 2 branches out.
4. **`writeWorstSection(w io.Writer, worst []Score, threshold float64) error`** — Worst offenders table. Moves 3 branches out.

**Expected result**: `WriteText` complexity 18 → ~5.

### D3: classify.ComputeScore decomposition

Extract 2 pure-function helpers:

1. **`accumulateSignals(signals []taxonomy.Signal) (score int, hasPositive, hasNegative bool)`** — Signal weight summation loop with zero/empty filtering. Pure function. Moves 5 branches out.
2. **`classifyLabel(score, contractualThreshold, incidentalThreshold int) (taxonomy.ClassificationLabel, string)`** — Threshold comparison → label + reasoning string. Pure function. Moves 3 branches out.

Clamping (2 branches) is trivial enough to stay inline. Signal filtering (2 branches) is already a simple loop.

**Expected result**: `ComputeScore` complexity 16 → ~7.

### D4: aireport.Run decomposition

Extract 3 helpers matching the function's natural three-path structure:

1. **`validateRunnerOpts(opts *RunnerOptions) error`** — Format defaulting, nil adapter check, binary validation. Moves 4 branches out.
2. **`runJSONPath(payload *ReportPayload, opts RunnerOptions) error`** — JSON encode to stdout + threshold evaluation. Moves 2 branches out.
3. **`runTextPath(payload *ReportPayload, opts RunnerOptions) error`** — Marshal payload, set timeout, invoke adapter, check empty, write stdout, write step summary, evaluate thresholds. Moves 7 branches out.

**Expected result**: `Run` complexity 16 → ~5.

### D5: OllamaAdapter.Format decomposition

Extract 1 helper (the host resolution has the most concentrated branching):

1. **`resolveOllamaHost(configHost string) (string, error)`** — Config → env → default fallback, URL validation (scheme + host check). Returns the full `/api/generate` URL. Moves 5 branches out.

The request construction and execution are already relatively clean and have clear error returns. Extracting the host resolution alone reduces complexity sufficiently.

**Expected result**: `Format` complexity 16 → ~10.

## Risks / Trade-offs

### R1: Minimal risk — pure refactoring with strong existing tests

All 5 functions have integration-level tests that exercise every branch. The mechanical extraction preserves exact behavior. Any regression will be caught by existing tests.

### R2: Helper proliferation

14 new helpers across 4 packages. Each is small (5-15 lines), co-located with its parent, and has clear GoDoc. This is acceptable — the alternative is monolithic functions with 16-22 branches.
