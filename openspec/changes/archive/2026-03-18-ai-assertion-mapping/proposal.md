## Why

Assertion-to-effect mapping accuracy is 86.4% (57/66). The remaining 9 unmapped assertions are **semantically valid but structurally disconnected** from their target's side effects. Static analysis can trace variable flow but cannot understand semantic relationships like "calling `store.get()` after `store.set()` verifies the mutation."

Examples of the gap:
- A test calls `counter.increment(5)` then verifies via `counter.value()` — a different method that reads the state `increment` wrote
- A test uses `cmp.Diff(expected, result)` where `result` passes through a third-party function, losing type-checker identity
- A helper creates an internal variable from the target's return value, and assertions reference the helper's local, not a parameter

These patterns are common in real Go test code. Closing the gap requires understanding what the code *means*, not just tracing how variables flow — which is exactly what an AI model excels at.

## What Changes

Two layers of AI-assisted assertion mapping:

**Layer 1 — Library API**: Add an optional `AIMapperFunc` callback to `quality.Options` that is called as a final fallback for assertions that all mechanical passes fail to map. The AI receives the assertion expression, target function's side effects, and test body. Returns a side effect ID (mapped at confidence 50) or empty string. When nil, behavior is identical to today. Includes `BuildAIMapperPrompt` and `ParseAIMapperResponse` helpers for callers to build their own implementations.

**Layer 2 — Agent-level mapping via `/gaze` command**: Update the gaze-reporter agent prompt to evaluate unmapped assertions directly. When the agent runs `gaze quality --format=json`, it parses unmapped assertions from the output, reads the source files for context, evaluates semantic relationships, and reports AI-mapped assertions in the quality section with confidence indicators. This requires no changes to the gaze binary — the agent IS the AI mapper.

## Capabilities

### New Capabilities
- `AIMapperFunc` on `quality.Options`: Optional callback for AI-assisted assertion mapping. Signature: `func(ctx AIMapperContext) (effectID string, err error)`.
- `AIMapperContext` struct: Contains assertion source, target function name, side effects list, and test body source — everything the AI needs to evaluate the semantic relationship.
- `BuildAIMapperPrompt`: Constructs structured prompt from `AIMapperContext` for any AI backend.
- `ParseAIMapperResponse`: Extracts effect ID from AI response with NONE handling and embedded-ID search.
- gaze-reporter agent: Evaluates unmapped assertions during `/gaze` command execution by reading source files and making semantic judgments.

### Modified Capabilities
- `MapAssertionsToEffects`: After all mechanical passes fail, calls `AIMapperFunc` if non-nil. Maps at confidence 50 when AI returns a match.
- `Assess`: Threads `AIMapperFunc` from `Options` through to `MapAssertionsToEffects`.

### Removed Capabilities
- None.

## Impact

| File | Change |
|------|--------|
| `internal/quality/quality.go` | Add `AIMapperFunc` to `Options`, thread through to `MapAssertionsToEffects` |
| `internal/quality/mapping.go` | Add AI fallback after mechanical passes |
| `internal/quality/ai_mapper.go` | New file: `AIMapperContext` struct, prompt construction |
| `cmd/gaze/main.go` | Deferred — CLI wiring for `--ai` flag is a follow-up change |
| `internal/scaffold/assets/agents/gaze-reporter.md` | Add unmapped assertion evaluation instructions |
| `.opencode/agents/gaze-reporter.md` | Sync with embedded asset |
| `internal/aireport/assets/agents/gaze-reporter.md` | Sync with embedded asset |
| Tests | New tests with mock AI mapper, prompt builder, response parser |
| `AGENTS.md` | Update Recent Changes |

## Constitution Alignment

### I. Accuracy — PASS
AI mapping is additive — it only runs for assertions that mechanical analysis cannot map. It cannot reduce accuracy because it only adds new mappings (never removes or overrides mechanical ones). The lower confidence (50) signals to consumers that these are AI-evaluated.

### II. Minimal Assumptions — PASS
AI mapping is opt-in via `--ai` flag. Default behavior is unchanged. No new dependencies unless the user explicitly enables AI.

### III. Actionable Output — PASS
Previously unmapped assertions become mapped, improving contract coverage accuracy. The confidence level distinguishes AI-mapped from mechanically-mapped assertions.

### IV. Testability — PASS
The callback pattern makes the AI mapper testable in isolation — tests inject a mock function that returns predetermined results. No real AI calls needed in tests.
