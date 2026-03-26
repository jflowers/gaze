## Context

Gaze's assertion mapping pipeline has four mechanical passes (direct identity at 75, helper bridge at 70, indirect root at 65, inline call at 60, container unwrap at 55) followed by an AI mapper stub at confidence 50. The container unwrap pass (spec 034) traces through field access, index operations, and transformation calls like `json.Unmarshal`. However, `resolveExprRoot` — used in both the forward-tracing loop and the assertion-level fallback — does not handle `*ast.TypeAssertExpr`, breaking the chain for real-world patterns like `result.Content[0].(mcp.TextContent).Text`.

The `AIMapperFunc` infrastructure (`ai_mapper.go`) is fully implemented with `BuildAIMapperPrompt`, `ParseAIMapperResponse`, and 7 passing tests — but no CLI command wires it up. The existing AI adapter infrastructure in `internal/aireport/` (Claude, Gemini, Ollama, OpenCode) provides the subprocess/HTTP integration layer.

## Goals / Non-Goals

### Goals

- Fix `resolveExprRoot` to handle `*ast.TypeAssertExpr` (complete the container unwrap chain)
- Wire `AIMapperFunc` to `gaze quality` and `gaze crap` via an opt-in `--ai-mapper` flag
- Reuse the existing `internal/aireport` adapter infrastructure for AI backend communication
- Add a `verify` action to the gaze-test-generator agent for coverage feedback loops

### Non-Goals

- Adding AI to classification (determinism is critical for CRAP ratchets)
- Adding AI to effect detection (non-deterministic detection breaks accuracy guarantees)
- Creating new AI adapters — reusing existing Claude/Gemini/Ollama/OpenCode adapters
- Making AI mapper the default — it must be opt-in via flag

## Decisions

### D1: TypeAssertExpr in resolveExprRoot

Add `case *ast.TypeAssertExpr: return resolveExprRoot(e.X, info)` to the switch statement in `resolveExprRoot` (`mapping.go:680-714`). TypeAssertExpr.X is the expression being asserted (e.g., in `x.(T)`, X is `x`). Recursing into X follows the same pattern as SelectorExpr and IndexExpr.

**Constitution alignment**: Accuracy (I) — reduces false negatives for type assertion chains. Testability (IV) — testable by adding a TypeAssertExpr case to the existing container unwrap fixture.

### D2: AI Mapper Adapter Architecture

Create a thin `AIMapperAdapter` function that bridges `AIMapperFunc` to the existing `aireport.AIAdapter` interface. The function:

1. Receives `AIMapperContext` (assertion source, test source, target name, effects)
2. Calls `BuildAIMapperPrompt(ctx)` to produce the prompt string
3. Calls `adapter.Format(ctx, prompt)` on the selected AI adapter
4. Calls `ParseAIMapperResponse(response, ctx.Effects)` to extract the effect ID
5. Returns the effect ID (or empty string for "NONE")

The adapter selection uses the same pattern as `gaze report --ai=<backend>`:

- `claude` → `ClaudeAdapter`
- `gemini` → `GeminiAdapter`
- `ollama` → `OllamaAdapter`
- `opencode` → `OpenCodeAdapter`

**Design decision**: Reuse `aireport.AIAdapter` rather than creating a new adapter interface. The AI mapper prompt is a string, the response is a string — the adapter just needs to format input and return output. The existing adapters handle binary lookup, subprocess management, and error formatting.

**Constitution alignment**: Composability (II) — the AI mapper is opt-in and defaults to off. Observable Quality (III) — AI-mapped assertions carry confidence 50 and are distinguishable in JSON output.

### D3: CLI Flag Propagation

The `--ai-mapper=<backend>` flag is added to the `quality` command in `cmd/gaze/main.go`. The flag value flows through:

1. `runQuality` params → `quality.Options.AIMapperFunc`
2. `quality.Assess` passes the callback to `MapAssertionsToEffects`
3. `mapAssertionsToEffectsImpl` calls `tryAIMapping` when all mechanical passes fail
4. `tryAIMapping` calls the `AIMapperFunc` callback

For `gaze crap`, the flag propagates through `BuildContractCoverageFunc` → `analyzePackageCoverage` → `quality.Assess`. The existing `crap` command already passes `quality.Options` to the coverage builder.

**No changes to `quality.Options`** — the `AIMapperFunc` field already exists (line 54 of `quality.go`). Only the CLI layer needs to construct and set it.

### D4: Test Generator Verify Action

Add a `verify` action to `.opencode/agents/gaze-test-generator.md` that:

1. Records the current contract coverage for the target package (from the quality JSON passed as input)
2. After test generation, runs `gaze quality --format=json <package>` on the target package
3. Compares before/after contract coverage percentages
4. Reports the improvement delta (e.g., "Contract coverage: 25% → 67% (+42%)")

This is an agent-level change only — no production code modifications. The agent uses the existing `gaze quality` CLI command for measurement.

### D5: Container Unwrap Fixture Enhancement

Add a type-assertion variant to the `containerunwrap` test fixture that mirrors the real MCP pattern:

```go
type Content interface{ content() }
type TextContent struct { Text string }
func (TextContent) content() {}
type ResultWithInterface struct { Content []Content }
func WrapWithInterface(key, value string) *ResultWithInterface { ... }
```

Test:

```go
result := WrapWithInterface("key", "value")
text := result.Content[0].(TextContent).Text
var data map[string]any
json.Unmarshal([]byte(text), &data)
if data["key"] != "value" { t.Error(...) }
```

This validates that the `resolveExprRoot` TypeAssertExpr fix enables the full container unwrap chain through interface slice type assertions.

## Risks / Trade-offs

### AI Mapper Latency

Each unmapped assertion triggers an AI call (1-5 seconds). For a package with 20 unmapped assertions, this adds 20-100 seconds. Mitigation: the flag is opt-in, and only fires when mechanical passes fail. Users choose to trade latency for mapping accuracy.

### AI Mapper Non-Determinism

Different runs may produce different AI mappings (confidence 50). This means contract coverage can fluctuate slightly between runs when `--ai-mapper` is enabled. Mitigation: AI mappings are the lowest confidence tier and are clearly distinguishable in output. Mechanical mappings (confidence 55-75) are deterministic and take precedence. CRAP ratchets should be evaluated against mechanical-only results.

### AI Mapper Cost

Each AI call costs ~$0.01-0.05 depending on the backend. For a module with 50 unmapped assertions, this is $0.50-2.50 per run. Mitigation: opt-in flag with no default. Users control when to incur costs.

### TypeAssertExpr False Positives

Adding TypeAssertExpr to `resolveExprRoot` could theoretically resolve through type assertions to reach unrelated variables. Mitigation: `resolveExprRoot` is only used in the mapping pipeline where the tracked variable set is already scoped to the target function's return values. The risk of false positives is minimal — type assertions narrow types, they don't introduce new variables.
