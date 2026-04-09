# JSON Schema Reference

Three Gaze commands produce structured JSON output via `--format=json`: [`analyze`](cli/analyze.md), [`crap`](cli/crap.md), and [`quality`](cli/quality.md). The [`report`](cli/report.md) command also supports `--format=json`, which outputs the combined analysis payload.

Gaze embeds JSON Schemas (Draft 2020-12) for the analyze and quality outputs. Use `gaze schema` to print the analyze schema.

## Analyze Output

**Command**: `gaze analyze <package> --format=json`
**Schema ID**: `https://github.com/unbound-force/gaze/analysis-report.schema.json`

### Top-Level Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | `string` | Yes | Schema version (semver) |
| `results` | `AnalysisResult[]` | Yes | Array of per-function analysis results |

### AnalysisResult

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `target` | `FunctionTarget` | Yes | Function metadata (package, name, signature, location) |
| `side_effects` | `SideEffect[]` | Yes | Detected side effects |
| `metadata` | `Metadata` | Yes | Analysis metadata (version, timing) |

### FunctionTarget

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `package` | `string` | Yes | Full import path |
| `function` | `string` | Yes | Function or method name. `<package>` indicates package-level declarations (e.g., sentinel errors). |
| `receiver` | `string` | No | Receiver type for methods (e.g., `*Store`) |
| `signature` | `string` | Yes | Full function signature |
| `location` | `string` | Yes | Source position (`file:line:col`) |

### SideEffect

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Stable identifier (`se-XXXXXXXX`) |
| `type` | `string` | Yes | One of 37 [side effect types](../concepts/side-effects.md) (e.g., `ReturnValue`, `ErrorReturn`, `ReceiverMutation`) |
| `tier` | `string` | Yes | Priority tier: `P0`, `P1`, `P2`, `P3`, or `P4` |
| `location` | `string` | Yes | Source position |
| `description` | `string` | Yes | Human-readable explanation |
| `target` | `string` | Yes | Affected entity (field, variable, type, etc.) |
| `classification` | `Classification` | No | Only present when `--classify` is used |

### Classification

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `label` | `string` | Yes | `contractual`, `incidental`, or `ambiguous` |
| `confidence` | `int` | Yes | Confidence score (0–100) |
| `signals` | `Signal[]` | Yes | Evidence signals that contributed to the score |
| `reasoning` | `string` | No | Human-readable summary |

### Signal

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | `string` | Yes | Signal source (e.g., `interface`, `caller`, `naming`, `godoc`, `readme`) |
| `weight` | `int` | Yes | Numeric contribution to confidence (can be negative) |
| `source_file` | `string` | No | File path (verbose mode only) |
| `excerpt` | `string` | No | Short quote from source (verbose mode only) |
| `reasoning` | `string` | No | Explanation (verbose mode only) |

### Annotated Example

```json
{
  "version": "0.9.0",
  "results": [
    {
      "target": {
        "package": "github.com/example/pkg",
        "function": "Save",
        "receiver": "*Store",
        "signature": "func (*Store).Save(ctx context.Context, item Item) error",
        "location": "store.go:42:1"
      },
      "side_effects": [
        {
          "id": "se-a1b2c3d4",
          "type": "ErrorReturn",
          "tier": "P0",
          "location": "store.go:55:3",
          "description": "returns error value",
          "target": "error",
          "classification": {
            "label": "contractual",
            "confidence": 92,
            "signals": [
              { "source": "interface", "weight": 15 },
              { "source": "naming", "weight": 5 }
            ],
            "reasoning": "Error return is part of the function's behavioral contract"
          }
        }
      ],
      "metadata": {
        "gaze_version": "0.9.0",
        "go_version": "go1.25.0",
        "duration_ms": 142,
        "timestamp": "2026-04-08T10:30:00Z",
        "warnings": null
      }
    }
  ]
}
```

---

## CRAP Output

**Command**: `gaze crap <packages> --format=json`

The CRAP JSON output is not covered by a formal embedded schema but follows a stable structure.

### Top-Level Structure

| Field | Type | Description |
|-------|------|-------------|
| `scores` | `Score[]` | Per-function CRAP scores |
| `summary` | `Summary` | Aggregate statistics |

### Score

| Field | Type | Description |
|-------|------|-------------|
| `package` | `string` | Go package name |
| `function` | `string` | Function or method name |
| `file` | `string` | Source file path |
| `line` | `int` | Line number of function declaration |
| `complexity` | `int` | Cyclomatic complexity |
| `line_coverage` | `float64` | Line coverage percentage (0–100) |
| `crap` | `float64` | Classic CRAP score |
| `contract_coverage` | `float64?` | Contract coverage percentage (omitted when unavailable) |
| `gaze_crap` | `float64?` | GazeCRAP score (omitted when unavailable) |
| `quadrant` | `string?` | Quadrant classification: `Q1_Safe`, `Q2_ComplexButTested`, `Q3_SimpleButUnderspecified`, `Q4_Dangerous` |
| `fix_strategy` | `string?` | Remediation action: `decompose`, `add_tests`, `add_assertions`, `decompose_and_test` (only for CRAPload functions) |
| `contract_coverage_reason` | `string?` | Diagnostic reason for contract coverage value (e.g., when all effects are ambiguous) |
| `effect_confidence_range` | `[int, int]?` | Min/max classification confidence across all side effects |

### Summary

| Field | Type | Description |
|-------|------|-------------|
| `total_functions` | `int` | Total analyzed functions |
| `avg_complexity` | `float64` | Average cyclomatic complexity |
| `avg_line_coverage` | `float64` | Average line coverage |
| `avg_crap` | `float64` | Average CRAP score |
| `crapload` | `int` | Count of functions at or above CRAP threshold |
| `crap_threshold` | `float64` | CRAP score threshold used |
| `gaze_crapload` | `int?` | Count of functions at or above GazeCRAP threshold |
| `gaze_crap_threshold` | `float64?` | GazeCRAP threshold used |
| `avg_gaze_crap` | `float64?` | Average GazeCRAP score |
| `avg_contract_coverage` | `float64?` | Average contract coverage |
| `quadrant_counts` | `map[string]int?` | Count of functions per quadrant |
| `fix_strategy_counts` | `map[string]int?` | Count of functions per fix strategy |
| `worst_crap` | `Score[]` | Top functions by CRAP score |
| `worst_gaze_crap` | `Score[]?` | Top functions by GazeCRAP score |
| `recommended_actions` | `RecommendedAction[]?` | Prioritized remediation list (top 20) |
| `ssa_degraded_packages` | `string[]?` | Packages where SSA construction failed |

### Annotated Example

```json
{
  "scores": [
    {
      "package": "github.com/example/pkg",
      "function": "(*Store).Save",
      "file": "store.go",
      "line": 42,
      "complexity": 8,
      "line_coverage": 85.0,
      "crap": 10.2,
      "contract_coverage": 66.7,
      "gaze_crap": 14.1,
      "quadrant": "Q1_Safe",
      "fix_strategy": null
    }
  ],
  "summary": {
    "total_functions": 142,
    "avg_complexity": 4.2,
    "avg_line_coverage": 78.5,
    "avg_crap": 8.7,
    "crapload": 5,
    "crap_threshold": 15.0,
    "gaze_crapload": 3,
    "gaze_crap_threshold": 15.0,
    "avg_gaze_crap": 9.1,
    "avg_contract_coverage": 72.3,
    "quadrant_counts": {
      "Q1_Safe": 120,
      "Q2_ComplexButTested": 12,
      "Q3_SimpleButUnderspecified": 7,
      "Q4_Dangerous": 3
    },
    "fix_strategy_counts": {
      "add_tests": 2,
      "add_assertions": 2,
      "decompose": 1
    },
    "worst_crap": [],
    "worst_gaze_crap": [],
    "recommended_actions": [],
    "ssa_degraded_packages": []
  }
}
```

---

## Quality Output

**Command**: `gaze quality <package> --format=json`
**Schema ID**: `https://github.com/unbound-force/gaze/quality-report.schema.json`

### Top-Level Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `quality_reports` | `QualityReport[]` | Yes | Per-test quality assessments |
| `quality_summary` | `PackageSummary` | Yes | Package-level aggregate statistics |

### QualityReport

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `test_function` | `string` | Yes | Name of the test function |
| `test_location` | `string` | Yes | Source position (`file:line`) |
| `target_function` | `FunctionTarget` | Yes | The function being tested |
| `contract_coverage` | `ContractCoverage` | Yes | Coverage of contractual side effects |
| `over_specification` | `OverSpecificationScore` | Yes | Assertions on incidental effects |
| `ambiguous_effects` | `SideEffectRef[]?` | No | Effects excluded from metrics due to ambiguous classification |
| `unmapped_assertions` | `AssertionMapping[]?` | No | Assertions that could not be linked to any side effect |
| `assertion_count` | `int` | No | Total detected assertion sites in the test function |
| `assertion_detection_confidence` | `int` | Yes | Fraction of assertions successfully pattern-matched (0–100) |
| `metadata` | `Metadata` | Yes | Analysis metadata |

### ContractCoverage

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `percentage` | `float64` | Yes | Coverage ratio (0–100) |
| `covered_count` | `int` | Yes | Number of contractual effects asserted on |
| `total_contractual` | `int` | Yes | Total contractual effects |
| `gaps` | `SideEffectRef[]?` | No | Contractual effects NOT asserted on |
| `gap_hints` | `string[]?` | No | Go code snippets suggesting how to assert on each gap (parallel to `gaps`) |
| `discarded_returns` | `SideEffectRef[]?` | No | Contractual return/error effects explicitly discarded (e.g., `_ = target()`) |
| `discarded_return_hints` | `string[]?` | No | Code snippets for discarded returns (parallel to `discarded_returns`) |

### OverSpecificationScore

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `count` | `int` | Yes | Number of incidental side effects asserted on |
| `ratio` | `float64` | Yes | Incidental assertions / total assertions (0–1) |
| `incidental_assertions` | `AssertionMapping[]?` | No | Details of incidental assertions |
| `suggestions` | `string[]?` | No | Actionable advice per incidental assertion |

### PackageSummary

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `total_tests` | `int` | Yes | Number of test functions analyzed |
| `average_contract_coverage` | `float64` | Yes | Mean coverage across all tests (0–100) |
| `total_over_specifications` | `int` | Yes | Total incidental assertions across all tests |
| `assertion_detection_confidence` | `int` | Yes | Package-level assertion detection confidence |
| `worst_coverage_tests` | `QualityReport[]?` | No | Bottom 5 tests by coverage |
| `ssa_degraded` | `bool` | No | `true` when SSA construction failed (partial results) |
| `ssa_degraded_packages` | `string[]` | No | Package paths where SSA construction failed |

### Annotated Example

```json
{
  "quality_reports": [
    {
      "test_function": "TestSave_Success",
      "test_location": "store_test.go:15",
      "target_function": {
        "package": "github.com/example/pkg",
        "function": "Save",
        "receiver": "*Store",
        "signature": "func (*Store).Save(ctx context.Context, item Item) error",
        "location": "store.go:42:1"
      },
      "contract_coverage": {
        "percentage": 66.7,
        "covered_count": 2,
        "total_contractual": 3,
        "gaps": [
          {
            "id": "se-d4e5f6a7",
            "type": "ReceiverMutation",
            "tier": "P0",
            "description": "mutates field Store.lastSaved",
            "target": "Store.lastSaved"
          }
        ],
        "gap_hints": [
          "if store.lastSaved.IsZero() { t.Error(\"expected lastSaved to be set\") }"
        ],
        "discarded_returns": null,
        "discarded_return_hints": null
      },
      "over_specification": {
        "count": 0,
        "ratio": 0.0,
        "incidental_assertions": null,
        "suggestions": null
      },
      "ambiguous_effects": null,
      "unmapped_assertions": null,
      "assertion_count": 3,
      "assertion_detection_confidence": 100,
      "metadata": {
        "gaze_version": "0.9.0",
        "go_version": "go1.25.0",
        "duration_ms": 89,
        "timestamp": "2026-04-08T10:30:00Z",
        "warnings": [
          "classification: mechanical signals only; run /gaze in full mode for document-enhanced results"
        ]
      }
    }
  ],
  "quality_summary": {
    "total_tests": 8,
    "average_contract_coverage": 75.0,
    "total_over_specifications": 1,
    "assertion_detection_confidence": 95,
    "ssa_degraded": false,
    "ssa_degraded_packages": []
  }
}
```

---

## Report JSON Output

**Command**: `gaze report <packages> --format=json`

The report JSON output combines the results of all four analysis operations into a single payload. It includes CRAP scores, quality reports, classification counts, and document scan results. This is the same payload that is sent to the AI adapter in text mode.

The structure is an internal format used by the report pipeline. For stable, well-defined schemas, prefer using `gaze analyze`, `gaze crap`, or `gaze quality` with `--format=json` individually.

## Retrieving Schemas Programmatically

```bash
# Print the analyze schema
gaze schema

# The quality schema is embedded but not exposed via CLI.
# Reference it at: internal/report/schema.go (QualitySchema constant)
```

## See Also

- [`gaze schema`](cli/schema.md) — print the embedded analyze JSON Schema
- [`gaze analyze`](cli/analyze.md) — side effect detection
- [`gaze crap`](cli/crap.md) — CRAP score computation
- [`gaze quality`](cli/quality.md) — test quality assessment
- [Glossary](glossary.md) — definitions of all domain terms used in JSON output
