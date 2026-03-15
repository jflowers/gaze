## Why

Five functions remain in Q4 Dangerous of the GazeCRAP quadrant — high cyclomatic complexity (16-22) with contract coverage that can't compensate. All five already have strong test coverage; the path forward is decomposition, not more tests.

| Function | Complexity | Package |
|----------|-----------|---------|
| `scaffold.Run` | 22 | `internal/scaffold` |
| `crap.WriteText` | 18 | `internal/crap` |
| `classify.ComputeScore` | 16 | `internal/classify` |
| `aireport.Run` | 16 | `internal/aireport` |
| `(*OllamaAdapter).Format` | 16 | `internal/aireport` |

Reducing these to complexity 10 or below moves them out of Q4 and reduces the GazeCRAPload from 8 toward 3 (the Q3 functions that need contract assertions, not decomposition).

## What Changes

Extract unexported helper functions from each of the five functions. Each extraction moves a cohesive group of branches into a standalone function, reducing the parent's cyclomatic complexity. All helpers are package-private. No API surface changes, no output changes, no behavioral changes.

## Capabilities

### New Capabilities
- `scaffold.applyDefaults`: Defaults `TargetDir`, `Version`, `Stdout` on Options.
- `scaffold.handleToolOwnedFile`: Tool-owned overwrite-on-diff logic (read existing, compare, write if different).
- `scaffold.writeNewFile`: Directory creation + file write + action label.
- `crap.writeScoreTable`: Table construction with threshold markers and styles.
- `crap.writeSummarySection`: Summary stats rendering with CRAPload/GazeCRAP.
- `crap.writeQuadrantSection`: Quadrant breakdown rendering.
- `crap.writeWorstSection`: Worst offenders rendering.
- `classify.accumulateSignals`: Sum weights, track positive/negative presence.
- `classify.classifyLabel`: Threshold-based label assignment with reasoning.
- `aireport.validateRunnerOpts`: Format defaulting, nil adapter check, binary validation.
- `aireport.runJSONPath`: JSON encode + threshold evaluation.
- `aireport.runTextPath`: AI adapter invocation + output writing + step summary.
- `aireport.resolveOllamaHost`: Host resolution with config/env/default fallback + URL validation.

### Modified Capabilities
- `scaffold.Run`: Orchestrator calling extracted helpers. Complexity 22 -> ~8.
- `crap.WriteText`: Orchestrator calling section writers. Complexity 18 -> ~4.
- `classify.ComputeScore`: Orchestrator calling signal/label helpers. Complexity 16 -> ~6.
- `aireport.Run`: Orchestrator calling validation/path helpers. Complexity 16 -> ~6.
- `(*OllamaAdapter).Format`: Orchestrator calling host/request helpers. Complexity 16 -> ~8.

### Removed Capabilities
- None.

## Impact

| Package | Files Modified | Helpers Added |
|---------|---------------|--------------|
| `internal/scaffold` | `scaffold.go` | 3 |
| `internal/crap` | `report.go` | 4 |
| `internal/classify` | `score.go` | 2 |
| `internal/aireport` | `runner.go`, `adapter_ollama.go` | 5 |

No test file changes expected — existing tests exercise the full functions and catch regressions.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No changes to artifact formats or cross-hero interfaces. All extracted helpers are package-private.

### II. Composability First

**Assessment**: PASS

No API surface changes. All functions retain their current signatures and behavior. Callers are unaffected.

### III. Observable Quality

**Assessment**: PASS

No changes to output formats, JSON schemas, or report structure. The refactoring improves internal code quality metrics without altering any observable outputs.

### IV. Testability

**Assessment**: PASS

All five functions already have strong test coverage. The decomposition makes individual phases independently testable for future test additions. Existing tests catch any regressions from the extraction.
