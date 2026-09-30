## Why

Cyclomatic complexity (used in CRAP scores) measures the number of linearly independent paths through a function. Cognitive complexity (Sonar's algorithm) measures how hard code is to *understand* — penalizing nesting, recursion, and breaks in linear flow. Both are valuable; cognitive complexity better captures the human/agent comprehension burden. Adding cognitive complexity to Gaze provides a complementary lens on code risk, especially for functions that are simple in path count but deeply nested or structurally confusing.

## What Changes

1. **Cognitive Complexity Analyzer**: Port the SonarSource cognitive complexity algorithm to Go AST analysis in a new `internal/cognitive/` package. Computes per-function cognitive complexity with nesting penalties, increment rules for control flow structures, and recursion detection.

2. **GazeCRAP-CC Variant**: Introduce a `GazeCRAP-CC` score that combines cognitive complexity with test coverage using the formula `GazeCRAP-CC = CC^2 * (1 - coverage)^3 + CC` (mirrors CRAP but substitutes cognitive complexity for cyclomatic complexity). Reported alongside existing CRAP and GazeCRAP scores, not replacing them.

3. **CI Gate Flag**: Add `--max-cognitive-complexity=<N>` flag to `gaze analyze` (and `gaze crap`) for CI gating on per-function cognitive complexity.

4. **JSON Output**: Extend JSON output to include `cognitive_complexity` and `gaze_crap_cc` per function.

## Capabilities

### New Capabilities
- `cognitive-complexity`: Per-function cognitive complexity computation via Go AST, following the SonarSource specification (nesting penalties, increment rules for if/else/switch/for/logical operators/goto/recursion).

### Modified Capabilities
- `crap-scoring`: Extended with `GazeCRAP-CC` variant alongside existing CRAP and GazeCRAP scores. New `--max-cognitive-complexity` CI gate flag.

## Impact

- **New package**: `internal/cognitive/` — cognitive complexity analyzer (Go AST walker).
- **Modified packages**: `internal/crap/` (Score struct, Summary struct, formula, text/JSON reporters), `internal/provider/goprovider/` (new `CognitiveComplexityProvider`), `internal/crap/provider.go` (new interface), `cmd/gaze/` (new flag wiring), `internal/report/` (JSON schema update), `internal/protocol/` (external analyzer protocol extension).
- **External analyzer protocol**: New `cognitive_complexity` method for language analyzers.
- **JSON Schema**: New fields `cognitive_complexity` and `gaze_crap_cc` on function scores.
- **No breaking changes**: All additions are additive; existing CRAP/GazeCRAP scores unchanged.

## Constitution Alignment

Assessed against the Gaze project constitution.

### I. Accuracy

**Assessment**: PASS

Cognitive complexity is a deterministic AST metric — the Sonar specification defines exact increment and nesting rules. The implementation will be verified against functions of known cognitive complexity (hand-computed test fixtures). False positives (incorrect cognitive complexity values) will be treated as bugs with regression tests.

### II. Minimal Assumptions

**Assessment**: PASS

The cognitive complexity analyzer operates on Go AST only — no source annotations, restructuring, or configuration required. It follows the same provider pattern as cyclomatic complexity, enabling external analyzers for other languages. No changes to existing analysis behavior.

### III. Actionable Output

**Assessment**: PASS

Cognitive complexity scores are reported per function in both text and JSON output, guiding users toward functions that are hard to understand. The `--max-cognitive-complexity` gate provides a clear CI pass/fail signal. GazeCRAP-CC complements existing CRAP/GazeCRAP scores, giving users a richer picture of risk.

### IV. Testability

**Assessment**: PASS

The cognitive complexity analyzer is a pure function of Go AST — trivially testable in isolation with synthetic AST fixtures. The GazeCRAP-CC formula is a pure arithmetic function. Both follow existing patterns (unit tests with table-driven cases, synthetic test fixtures in `testdata/`). Coverage strategy specified in design.
