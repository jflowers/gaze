# Improving Scores

Every function in the [CRAPload](../reference/glossary.md) — those with a [CRAP](../reference/glossary.md) score at or above the threshold — carries a **fix strategy** label that tells you exactly what to do. Gaze assigns one of four strategies based on the function's complexity, line coverage, and [contract coverage](../reference/glossary.md).

## Prioritization

Start with the easiest wins and work toward the hardest:

1. **`add_tests`** — Zero-coverage functions where adding any test brings immediate improvement
2. **`add_assertions`** — Functions with tests that run the code but don't verify behavior (Q3)
3. **`decompose_and_test`** — Functions that need both structural changes and tests
4. **`decompose`** — Functions so complex that even 100% coverage can't help

This order matches how Gaze sorts the `recommended_actions` list in JSON output. The [gaze-test-generator agent](opencode-integration.md) follows the same priority when processing `/gaze fix` commands.

## `add_tests` — Write Tests for Zero-Coverage Functions

### When It Applies

A function receives the `add_tests` strategy when:

- Its CRAP score is **at or above** the threshold (default: 15)
- Its **line coverage is 0%** (no test executes this function)
- Its **complexity is below** the threshold (so adding coverage alone will bring CRAP below threshold)

### What to Do

1. Identify the function's side effects with [`gaze analyze`](../reference/cli/analyze.md)
2. Write a test that calls the function and asserts on its [contractual](../reference/glossary.md) effects (return values, error returns, mutations)
3. Focus on the P0 effects first — these are the function's core contract

### Example

**Before**: `ParseConfig` has complexity 5 and 0% coverage.

```text
CRAP: 30.0  Complexity: 5  Coverage: 0.0%  [add_tests]
```

CRAP formula: `5² × (1 - 0/100)³ + 5 = 25 × 1 + 5 = 30`

**After**: Add a test that calls `ParseConfig` and asserts on the returned `*Config` and `error`:

```go
func TestParseConfig_ValidFile(t *testing.T) {
    cfg, err := ParseConfig("testdata/valid.yaml")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if cfg.Threshold != 15 {
        t.Errorf("Threshold = %d, want 15", cfg.Threshold)
    }
}

func TestParseConfig_MissingFile(t *testing.T) {
    _, err := ParseConfig("nonexistent.yaml")
    if err == nil {
        t.Fatal("expected error for missing file")
    }
}
```

With 100% coverage: `5² × (1 - 100/100)³ + 5 = 0 + 5 = 5` — well below the threshold of 15.

The function moves from [Q4 Dangerous](../reference/glossary.md) to [Q1 Safe](../reference/glossary.md).

## `add_assertions` — Strengthen Existing Tests

### When It Applies

A function receives the `add_assertions` strategy when:

- Its CRAP score is **at or above** the threshold
- It is in the **Q3 (Simple But Underspecified)** [quadrant](../reference/glossary.md) — meaning it has adequate line coverage but inadequate contract coverage
- Tests execute the code but don't assert on the function's contractual side effects

This is the most common gap in well-tested codebases: tests that run code paths without verifying the observable results.

### What to Do

1. Run [`gaze quality --verbose`](../reference/cli/quality.md) to see which contractual effects lack assertions
2. Look at the `gap_hints` in JSON output — these list the specific unasserted effects
3. Add assertions to existing tests that target the missing effects

### Example

**Before**: `FormatReport` has complexity 8, 90% line coverage, but 0% contract coverage. Tests call it but only check that it doesn't panic — they never assert on the returned string.

```text
CRAP: 8.1   GazeCRAP: 520.2   Quadrant: Q3_SimpleButUnderspecified  [add_assertions]
```

The line coverage keeps CRAP low (8.1), but [GazeCRAP](../reference/glossary.md) is extremely high (520.2) because no contractual effects are verified.

**After**: Add assertions on the return value:

```go
func TestFormatReport_IncludesHeader(t *testing.T) {
    report := FormatReport(sampleData)
    if !strings.Contains(report, "Summary") {
        t.Error("report missing Summary header")
    }
    if !strings.Contains(report, "CRAPload") {
        t.Error("report missing CRAPload section")
    }
}
```

With contract coverage now at 80%: GazeCRAP drops from 520.2 to ~9.6. The function moves from Q3 to Q1 Safe.

## `decompose_and_test` — Split and Test

### When It Applies

A function receives the `decompose_and_test` strategy when:

- Its **complexity is at or above** the threshold (so even 100% coverage can't bring CRAP below threshold)
- Its **line coverage is 0%** (no tests exist)
- It needs both structural simplification and test coverage

### What to Do

1. Identify cohesive blocks of logic within the function
2. Extract each block into a separate, well-named helper function
3. Write tests for each extracted function individually
4. Keep the original function as a thin orchestrator that delegates to the helpers

### Example

**Before**: `ProcessBatch` has complexity 22 and 0% coverage.

```text
CRAP: 10670.0  Complexity: 22  Coverage: 0.0%  [decompose_and_test]
```

Even with 100% coverage, CRAP would be 22 (above the threshold of 15). And there are no tests at all.

**After**: Extract three helpers:

```go
// Before: one 80-line function with 22 branches
func ProcessBatch(items []Item) (*Result, error) { ... }

// After: orchestrator + focused helpers
func ProcessBatch(items []Item) (*Result, error) {
    validated, err := validateItems(items)
    if err != nil {
        return nil, fmt.Errorf("validation: %w", err)
    }
    transformed := transformItems(validated)
    return aggregateResults(transformed), nil
}

func validateItems(items []Item) ([]Item, error) { ... }  // complexity: 6
func transformItems(items []Item) []Item { ... }           // complexity: 4
func aggregateResults(items []Item) *Result { ... }        // complexity: 3
```

Each helper has low complexity and can be tested independently. The orchestrator has complexity ~3. All four functions get tests, bringing every CRAP score well below threshold.

## `decompose` — Reduce Complexity

### When It Applies

A function receives the `decompose` strategy when:

- Its **complexity is at or above** the threshold
- Its **line coverage is greater than 0%** (tests exist, but complexity is the problem)
- Even with 100% coverage, CRAP would still be at or above the threshold (since CRAP at 100% coverage equals the complexity value)

### What to Do

1. Identify independent branches or code paths within the function
2. Extract each into a helper function with a clear name and contract
3. Test the extracted helpers individually
4. Verify the original function's complexity drops below the threshold

### Example

**Before**: `WriteText` has complexity 18 and 85% line coverage.

```text
CRAP: 20.5  Complexity: 18  Coverage: 85.0%  [decompose]
```

Even at 100% coverage: `18² × 0 + 18 = 18` — still above the threshold of 15.

**After**: Extract four helpers:

```go
// Before: one function handling all report sections
func WriteText(w io.Writer, report *Report) error { ... }  // complexity: 18

// After: orchestrator delegates to section writers
func WriteText(w io.Writer, report *Report) error {
    writeScoreTable(w, report.Scores)
    writeSummarySection(w, report.Summary)
    writeQuadrantSection(w, report.Summary)
    writeWorstSection(w, report.Summary)
    return nil
}

func writeScoreTable(w io.Writer, scores []Score) { ... }       // complexity: 4
func writeSummarySection(w io.Writer, s Summary) { ... }        // complexity: 3
func writeQuadrantSection(w io.Writer, s Summary) { ... }       // complexity: 3
func writeWorstSection(w io.Writer, s Summary) { ... }          // complexity: 5
```

The orchestrator drops to complexity 2. Each helper is independently testable. Existing tests continue to pass because the external behavior is unchanged.

## Reading the Fix Strategy in Output

### Text Output

Fix strategies appear as labels in the worst offenders section:

```text
CRAP    COMPLEXITY  COVERAGE  FUNCTION       FILE
----    ----------  --------  --------       ----
30.0 *  5           0.0%      ParseConfig    config.go:15       [add_tests]
20.5 *  18          85.0%     WriteText      text.go:20         [decompose]
```

### JSON Output

Each score in the CRAPload includes a `fix_strategy` field:

```json
{
  "function": "ParseConfig",
  "crap": 30.0,
  "complexity": 5,
  "line_coverage": 0.0,
  "fix_strategy": "add_tests"
}
```

The `summary.fix_strategy_counts` object shows the distribution:

```json
{
  "fix_strategy_counts": {
    "add_tests": 3,
    "add_assertions": 2,
    "decompose": 1,
    "decompose_and_test": 1
  }
}
```

The `summary.recommended_actions` array provides a prioritized remediation list (up to 20 entries), sorted by fix strategy priority then CRAP score descending.

## General Tips

- **Start with `add_tests`** — these are the quickest wins. A single test for a zero-coverage function can drop its CRAP score dramatically.
- **Focus on P0 effects** — return values, error returns, and mutations are the function's core contract. Assert on these first.
- **Don't chase 100% contract coverage** — diminishing returns set in around 80%. Focus on the functions Gaze flags as dangerous.
- **Use [`gaze quality --verbose`](../reference/cli/quality.md)** to see exactly which effects are unasserted — don't guess.
- **Ratchet your thresholds** — once you've improved scores, tighten your CI thresholds to prevent regression. See [CI Integration](ci-integration.md) for threshold configuration.
