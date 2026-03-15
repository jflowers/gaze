## Context

The assertion mapping pipeline has four mechanical passes:
1. Direct identity matching (confidence 75)
2. Indirect root resolution (confidence 65)
3. Helper parameter bridge (confidence 70)
4. Inline call matching (confidence 60)

These cover 86.4% of assertions. The remaining 13.6% are semantically valid but structurally disconnected — the assertion verifies the target's effect through an indirect path that static analysis cannot trace (calling a reader method to verify a writer method, passing values through third-party functions, etc.).

## Goals / Non-Goals

### Goals
- Add AI as a 5th (final) fallback pass for unmapped assertions
- Opt-in via callback on `quality.Options` — zero impact when not enabled
- Confidence 50 — clearly distinguished from mechanical mappings
- Model-agnostic — callback receives text, returns an effect ID
- Testable via mock injection

### Non-Goals
- Making AI a required dependency
- Using AI for assertions that mechanical analysis CAN map (waste of API calls)
- Changing confidence levels of existing mechanical passes
- Adding a new AI adapter (reuse existing adapter infrastructure from `aireport`)

## Decisions

### D1: AIMapperFunc callback signature

```go
// AIMapperContext contains the context needed for AI-assisted
// assertion mapping.
type AIMapperContext struct {
    AssertionSource string              // The assertion expression source code
    AssertionKind   AssertionKind       // Pattern type (stdlib_comparison, testify_equal, etc.)
    TestFuncSource  string              // Full test function source for context
    TargetFunc      string              // Target function qualified name
    SideEffects     []taxonomy.SideEffect // Available side effects to match against
}

// AIMapperFunc evaluates whether an unmapped assertion verifies a
// side effect of the target function. Returns the matched effect ID
// or empty string if no match. Called only for assertions that all
// mechanical passes failed to map.
type AIMapperFunc func(ctx AIMapperContext) (effectID string, err error)
```

The callback receives everything the AI needs in a single struct: the assertion code, the test function body for context, and the list of side effects to choose from. It returns either an effect ID (match found) or empty string (no match).

**Rationale**: A struct parameter is extensible — we can add fields later without changing the signature. The function returns `(string, error)` so callers can distinguish "no match" from "AI failed."

### D2: Integration point in MapAssertionsToEffects

After all mechanical passes fail and before classifying as unmapped:

```go
if mapping == nil && opts.AIMapperFunc != nil {
    // AI fallback for structurally disconnected assertions.
    mapping = tryAIMapping(site, targetFunc, effects, testPkg, opts.AIMapperFunc)
}
```

The `tryAIMapping` function constructs the `AIMapperContext` from the assertion site, target function, and effects, then calls the callback.

**Rationale**: The AI fallback is a natural extension of the existing pass chain. It runs after all fast/free passes and only for the ~14% of assertions that remain unmapped.

### D3: Thread AIMapperFunc through Assess

`MapAssertionsToEffects` currently doesn't receive `Options`. The cleanest approach is to add an `aiMapper AIMapperFunc` parameter (or pass the full `Options`). Since `MapAssertionsToEffects` is called in a tight loop inside `Assess`, I'll add a single parameter rather than threading the full Options struct.

### D4: Source code extraction for AI context

The AI needs readable source code, not AST nodes. Use `token.FileSet` to extract:
- Assertion expression: `format.Node` on `site.Expr`
- Test function body: `format.Node` on `site.FuncDecl` (or the test's `FuncDecl`)
- Target function name: `targetFunc.Name()` with package path

### D5: Prompt structure

The AI receives a structured prompt:

```
Given this Go test assertion:
  [assertion source]

In this test function:
  [test function source]

The target function [target name] produces these side effects:
  1. [effect type]: [description] (ID: [id])
  2. [effect type]: [description] (ID: [id])
  ...

Does this assertion verify any of these side effects? If yes, respond with ONLY the effect ID. If no, respond with NONE.
```

The prompt is deliberately constrained — the AI must respond with an effect ID or "NONE." This prevents hallucination and makes parsing trivial.

### D6: Error handling — silent degradation

If `AIMapperFunc` returns an error, the assertion remains unmapped (same as if AI returned "no match"). A warning is logged to `opts.Stderr` but the pipeline continues. AI failures should never crash the quality analysis.

## Risks / Trade-offs

### R1: Latency

Each unmapped assertion requires an AI API call. For a typical package with 5-10 unmapped assertions, this adds 5-10 seconds of latency. Mitigation: AI is opt-in, and the number of calls is bounded by the unmapped count (small by design — mechanical passes handle 86%+).

### R2: Non-determinism

AI responses may vary between runs. Mitigation: confidence 50 signals that these mappings are AI-evaluated. CI pipelines can use `--ai` for more accurate results while accepting non-determinism, or omit `--ai` for deterministic mechanical-only results.

### R3: Cost

API calls cost money. Mitigation: only unmapped assertions trigger calls (not all assertions). For a 66-assertion package, that's ~9 calls, not 66.
