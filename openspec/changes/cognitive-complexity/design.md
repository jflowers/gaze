## Context

Gaze currently uses cyclomatic complexity (via gocyclo) as the complexity input to CRAP and GazeCRAP scores. Cyclomatic complexity measures linearly independent paths but does not capture how hard code is for humans or agents to *read*. Sonar's cognitive complexity algorithm fills this gap by penalizing nesting, recursion, and breaks in linear flow.

The existing provider architecture (`internal/crap/provider.go`) already decouples complexity acquisition from scoring. Adding cognitive complexity follows the same pattern: a new provider interface, a Go implementation using `go/ast`, and integration into the scoring pipeline.

## Goals / Non-Goals

### Goals
- Compute per-function cognitive complexity via Go AST following the SonarSource specification
- Add `GazeCRAP-CC` variant using cognitive complexity in the CRAP formula
- Add `--max-cognitive-complexity` CI gate flag to `gaze analyze` and `gaze crap`
- Include `cognitive_complexity` and `gaze_crap_cc` in JSON output per function
- Extend external analyzer protocol with `cognitive_complexity` method
- Maintain backward compatibility — all existing scores and outputs unchanged

### Non-Goals
- Replacing cyclomatic complexity or existing CRAP/GazeCRAP scores
- Supporting languages other than Go in the built-in analyzer (external analyzers handle other languages via protocol)
- Cognitive complexity delta/baseline comparison (future work, mirrors Group 2c CRAP delta)
- Integrating cognitive complexity into quadrant classification (GazeCRAP-CC is reported alongside, not used for quadrants)

## Decisions

### D1: New `internal/cognitive/` package

The cognitive complexity analyzer lives in its own package, separate from `internal/analysis/` (side effect detection) and `internal/provider/goprovider/` (gocyclo wrapper). Rationale: cognitive complexity is a distinct algorithm with its own AST walking logic, test fixtures, and evolution path. Keeping it separate avoids bloating existing packages and follows the composability principle.

The package exports:
- `AnalyzeFunc(fset *token.FileSet, f *ast.FuncDecl) int` — single-function cognitive complexity
- `AnalyzeFile(fset *token.FileSet, f *ast.File) []FuncCognitiveComplexity` — file-level batch
- `FuncCognitiveComplexity` struct — mirrors `FunctionComplexity` shape

### D2: Sonar algorithm mapping to Go AST

The Sonar specification maps to Go constructs as follows:

| Sonar Increment | Go AST Node |
|---|---|
| `if` | `*ast.IfStmt` |
| `else if` | `*ast.IfStmt` in `Else` position |
| `else` | `*ast.IfStmt.Else` (when non-nil and not another `*ast.IfStmt`) |
| `switch` | `*ast.SwitchStmt`, `*ast.TypeSwitchStmt` |
| `for` | `*ast.ForStmt`, `*ast.RangeStmt` |
| `&&`, `\|\|` | `*ast.BinaryExpr` with `token.LAND`, `token.LOR` |
| `goto` | `*ast.BranchStmt` with `token.GOTO` |
| recursion | `*ast.CallExpr` where function name matches enclosing function |
| ternary | N/A in Go (no ternary operator) |
| `case` labels | No increment (per Sonar spec) |

Nesting penalty: +1 per level of nesting for `if`, `switch`, `for`. Nesting level tracked via a stack depth counter incremented on entry to nesting structures and decremented on exit. Logical operator sequences (`a && b && c`) count as a single increment per sequence (not per operator).

### D3: `CognitiveComplexityProvider` interface

A new interface in `internal/crap/provider.go`:

```go
type CognitiveComplexityProvider interface {
    Analyze(patterns []string, rootDir string) ([]FunctionCognitiveComplexity, error)
}
```

`FunctionCognitiveComplexity` mirrors `FunctionComplexity` with a `CognitiveComplexity int` field. This follows the existing pattern — same shape, different metric.

### D4: GazeCRAP-CC formula

`GazeCRAP-CC = CC^2 * (1 - coverage/100)^3 + CC`

where CC = cognitive complexity and coverage = line coverage (0-100). This mirrors the CRAP formula exactly, substituting cognitive complexity for cyclomatic complexity. The same formula is used for `GazeCRAP-CC` as for `CRAP` — only the complexity input differs.

The score is computed in `internal/crap/crap.go` alongside `Formula` and `GazeCRAP`. A new function `CognitiveFormula(complexity int, coveragePct float64) float64` implements this.

### D5: Score struct extension

`crap.Score` gains two new optional fields:
- `CognitiveComplexity *int` — nil when not computed (preserves backward compatibility for external analyzers that don't emit it)
- `GazeCRAPCC *float64` — nil when cognitive complexity or coverage unavailable

Using pointer types maintains JSON `omitempty` behavior and backward compatibility.

### D6: CI gate flag

`--max-cognitive-complexity=<N>` on `gaze analyze` and `gaze crap`. When any function exceeds N, exit code 1. Follows the same pattern as `--max-crapload` and `--max-gaze-crapload` — `int` + `cmd.Flags().Changed()` for zero-as-live-threshold semantics.

### D7: External analyzer protocol

New method `cognitive_complexity` in the JSON-RPC protocol. Request: same shape as `complexity` (package patterns). Response: `[]FunctionCognitiveComplexityData`. Optional — analyzers that don't implement it get a graceful degradation (cognitive complexity fields remain nil).

### D8: Coverage strategy

- **Unit tests**: `internal/cognitive/` tested with synthetic Go AST fixtures (parse source strings via `go/parser`). Table-driven tests covering all increment rules, nesting penalties, recursion detection, and logical operator sequences. Target: 100% branch coverage.
- **Integration tests**: `testdata/src/` fixtures with hand-computed cognitive complexity values. Verify `GoCognitiveComplexityProvider.Analyze` returns correct values.
- **Formula tests**: `CognitiveFormula` tested with known inputs/outputs, mirroring existing `TestFormula_*` tests.
- **Gate tests**: `--max-cognitive-complexity` flag tested via CLI unit tests (pass/fail/unchanged scenarios).

## Risks / Trade-offs

- **Algorithm fidelity**: The Sonar specification is designed for Java/C-family languages. Some constructs (e.g., Go's `select`, `defer`) are not covered by the spec. Decision: `select` gets a switch-like increment; `defer` does not increment (it is a resource management statement, not a control flow branch). Document these Go-specific decisions.
- **Performance**: Cognitive complexity requires a full AST walk per function. Mitigation: the AST is already loaded for side effect analysis; the cognitive walk is lightweight compared to SSA construction.
- **User confusion**: Adding a third complexity metric (cyclomatic, cognitive, GazeCRAP-CC) may overwhelm users. Mitigation: the text report omits cognitive complexity (the `crap` text report does not gain a cognitive/GazeCRAP-CC column); JSON always includes it for machine consumption.
