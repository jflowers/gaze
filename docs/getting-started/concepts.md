# Why Line Coverage Isn't Enough

Before diving into Gaze's commands, it helps to understand the problem it solves and the mental model behind it.

## The Problem with Line Coverage

Line coverage measures which lines of code *executed* during a test run. It does not measure whether the test *verified* anything.

A function can have 90% line coverage and tests that assert on nothing contractually meaningful -- logging calls, goroutine lifecycle, internal stdout writes -- while leaving the return values, error paths, and state mutations completely unverified. That function is dangerous to change, and traditional coverage metrics will not warn you.

Consider a function that parses a config file and returns a struct:

```go
func ParseConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    return &cfg, nil
}
```

A test that calls `ParseConfig("testdata/valid.json")` and checks only that `err == nil` achieves high line coverage. But it never verifies the returned `*Config` -- the function's primary [contract](../reference/glossary.md). If someone changes the parsing logic and the struct comes back with wrong values, this test still passes.

Line coverage says "tested." The function is not tested.

## What Contract-Level Analysis Adds

Gaze works from first principles to close this gap:

1. **Detect** every observable [side effect](../reference/glossary.md) a function produces -- return values, error returns, mutations, I/O, channel sends, and more.
2. **Classify** each effect as [contractual](../reference/glossary.md) (part of the function's public obligation), [incidental](../reference/glossary.md) (an implementation detail), or [ambiguous](../reference/glossary.md).
3. **Measure** whether your tests actually assert on the contractual effects -- and flag the specific ones they don't.

This produces metrics that answer the question line coverage cannot: *are your tests checking the right things?*

## The Three Core Metrics

### Contract Coverage

The percentage of a function's [contractual](../reference/glossary.md) side effects that at least one test assertion verifies.

```text
ContractCoverage% = (contractual effects asserted / total contractual effects) x 100
```

A function with 90% line coverage but 20% contract coverage has tests that run code without checking correctness. Gaze surfaces the specific effects that have no assertion -- the exact gaps you need to close.

For a deeper look at how effects are classified and scored, see [Classification](../concepts/classification.md).

### Over-Specification

The count and ratio of test assertions that target [incidental](../reference/glossary.md) effects -- implementation details that are not part of the function's contract.

Tests that assert on log output, goroutine lifecycle, or internal stdout writes will break during refactoring even when the function's actual contract is preserved. Gaze identifies each over-specified assertion and explains why it is fragile.

For details on how assertions are mapped to effects, see [Test Quality](../concepts/quality.md).

### GazeCRAP

A composite risk score that replaces line coverage with [contract coverage](../reference/glossary.md) in the [CRAP](../reference/glossary.md) (Change Risk Anti-Patterns) formula:

```text
CRAP(m)     = complexity^2 x (1 - lineCoverage/100)^3 + complexity
GazeCRAP(m) = complexity^2 x (1 - contractCoverage)^3  + complexity
```

Functions are placed in one of four [quadrants](../reference/glossary.md) based on both scores. The **Dangerous** quadrant -- complex functions whose tests don't verify their contracts -- is the highest-priority target for remediation.

For the full quadrant model, formulas, and fix strategies, see [Scoring](../concepts/scoring.md).

## How Gaze Works (Mental Model)

Think of Gaze as a three-stage pipeline:

```text
Your Go Code
     |
     v
 [1. Detect]     What does this function DO?
     |            (return values, errors, mutations, I/O, ...)
     v
 [2. Classify]   Which effects are part of the CONTRACT?
     |            (5 signal analyzers score each effect)
     v
 [3. Measure]    Do the TESTS verify the contract?
                  (map assertions to contractual effects)
```

**Stage 1** uses Go's AST (Abstract Syntax Tree) and SSA (Static Single Assignment) form to detect side effects without running your code. No annotations, no test framework changes, no restructuring required.

**Stage 2** applies five mechanical signal analyzers -- interface satisfaction, visibility, caller patterns, naming conventions, and GoDoc -- to score each effect's likelihood of being contractual.

**Stage 3** walks your test functions' call graphs, finds assertions, and maps them back to the detected effects. The ratio of asserted contractual effects to total contractual effects is your contract coverage.

For a deeper dive into the analysis pipeline, see [Analysis Pipeline](../concepts/analysis-pipeline.md).

## Next Steps

- [Quickstart](quickstart.md) -- install Gaze and produce your first analysis in under 10 minutes
- [Side Effects](../concepts/side-effects.md) -- the full taxonomy of 37 effect types across 5 tiers
- [Scoring](../concepts/scoring.md) -- CRAP, GazeCRAP, quadrants, and fix strategies
