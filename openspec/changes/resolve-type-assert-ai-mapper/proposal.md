## Why

Gaze's assertion mapping pipeline cannot trace through Go type assertions (`result.Content[0].(mcp.TextContent).Text`), and its AI-assisted mapping infrastructure — fully built and tested — has never been wired up to any CLI command. Together, these gaps leave ~15% of assertions unmapped for projects using container/wrapper types, causing functions to score Q3 (Simple But Underspecified) despite strong test coverage. Additionally, the gaze-test-generator agent has no feedback loop to verify that generated tests actually improve contract coverage.

This change addresses three complementary gaps:

1. **Bug fix**: `resolveExprRoot` in the mapping pipeline does not handle `*ast.TypeAssertExpr`, breaking the container unwrap chain for real-world patterns that use type assertions on interface slices.

2. **Feature activation**: The `AIMapperFunc` callback, prompt builder (`BuildAIMapperPrompt`), and response parser (`ParseAIMapperResponse`) exist with 7 passing tests but are never called from any CLI command. Wiring them up as an opt-in `--ai-mapper` flag on `gaze quality` and `gaze crap` handles the long-tail assertions no mechanical pass can reach.

3. **Feedback loop**: The gaze-test-generator agent generates tests from quality data but has no way to verify the generated tests actually improved coverage. Adding a `--verify` action that reruns quality analysis after generation closes the loop.

## What Changes

### 1. TypeAssertExpr in resolveExprRoot (bug fix)

Add `case *ast.TypeAssertExpr: return resolveExprRoot(e.X, info)` to the `resolveExprRoot` function in `internal/quality/mapping.go`. This is a single-line fix that completes the container unwrap pass for expressions containing type assertions.

### 2. Wire AIMapperFunc to CLI (feature activation)

Create an AI mapper adapter that translates the `AIMapperFunc` callback into calls to an AI backend (Claude, Gemini, Ollama, or OpenCode — reusing the existing `internal/aireport` adapter infrastructure). Add `--ai-mapper=<backend>` flag to `gaze quality` and propagate through to `gaze crap` via the `ContractCoverageFunc` pipeline. The AI mapper fires only when all 4 mechanical passes fail, at confidence 50 (lowest tier).

### 3. Test generator verify action (feedback loop)

Add a `verify` action to the gaze-test-generator agent that reruns `gaze quality --format=json` on the target package after test generation, compares before/after contract coverage, and reports the improvement delta.

## Capabilities

### New Capabilities

- `--ai-mapper=claude|gemini|ollama|opencode`: Opt-in AI-assisted assertion mapping on `gaze quality` and `gaze crap` commands. When enabled, unmapped assertions are evaluated by an AI model using the existing `BuildAIMapperPrompt` prompt. Mapped at confidence 50.
- `verify` action on gaze-test-generator: After generating tests, reruns quality analysis and reports contract coverage improvement.

### Modified Capabilities

- `resolveExprRoot`: Now handles `*ast.TypeAssertExpr` by recursing into the expression being asserted. Fixes container unwrap tracing for patterns using type assertions on interface slices.
- `gaze quality --json`: When `--ai-mapper` is set, the JSON output includes AI-mapped assertions at confidence 50 alongside mechanical mappings.

### Removed Capabilities

(none)

## Impact

- **`internal/quality/mapping.go`**: 1-line fix to `resolveExprRoot` for TypeAssertExpr
- **`internal/quality/quality.go`**: Wire `AIMapperFunc` from `Options` into `MapAssertionsToEffects` call
- **`cmd/gaze/main.go`**: Add `--ai-mapper` flag to `quality` and `crap` commands; construct adapter-backed `AIMapperFunc`
- **`internal/quality/ai_mapper.go`**: No changes — existing infrastructure is used as-is
- **`.opencode/agents/gaze-test-generator.md`**: Add `verify` action description
- **Assertion mapping accuracy**: Expected to increase from 85.1% to 90%+ with AI mapper enabled
- **Container unwrap coverage**: TypeAssertExpr patterns that previously failed now trace correctly

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: PASS

The TypeAssertExpr fix directly reduces false negatives — type assertion chains that broke the container unwrap pass now trace correctly. The AI mapper adds a lowest-confidence (50) fallback that only fires when all mechanical passes fail, providing additional coverage without overriding correct mechanical mappings. AI mappings are clearly annotated in output (confidence 50 distinguishes them from mechanical mappings at 55-75).

### II. Minimal Assumptions

**Assessment**: PASS

The TypeAssertExpr fix requires no user changes — it corrects an AST walking gap. The AI mapper is opt-in (`--ai-mapper` flag) and defaults to off, requiring no setup for users who don't want AI assistance. The test generator verify action works with existing `gaze quality` output.

### III. Actionable Output

**Assessment**: PASS

Previously-unmapped assertions become mapped (either mechanically via the TypeAssertExpr fix or via AI), increasing contract coverage percentages. The verify action provides concrete before/after coverage deltas so users can measure the impact of generated tests.

### IV. Testability

**Assessment**: PASS

The TypeAssertExpr fix is testable by adding a type-assertion case to the existing container unwrap test fixture. The AI mapper is already tested (7 tests in `ai_mapper_test.go`) — wiring it up requires integration tests with a mock adapter. The verify action is testable by running the agent on a package with known coverage gaps.
