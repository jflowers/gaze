# Classification

Once Gaze detects a function's [side effects](side-effects.md), the next question is: *which of these effects are part of the function's contract?* Classification answers this by assigning each side effect one of three labels — [contractual](../reference/glossary.md#contractual), [ambiguous](../reference/glossary.md#ambiguous), or [incidental](../reference/glossary.md#incidental) — based on weighted evidence from five mechanical signal analyzers.

Classification is the bridge between raw side effect detection and meaningful quality metrics. Only [contractual](../reference/glossary.md#contractual) effects count toward [contract coverage](../reference/glossary.md#contract-coverage). Only [incidental](../reference/glossary.md#incidental) effects count toward [over-specification](../reference/glossary.md#over-specification-score). [Ambiguous](../reference/glossary.md#ambiguous) effects are excluded from both metrics.

## The Three Labels

| Label | Meaning | Metric Impact |
|---|---|---|
| **Contractual** | The effect is part of the function's behavioral contract — callers depend on it | Counted in contract coverage (denominator and potentially numerator) |
| **Incidental** | The effect is an implementation detail — callers should not depend on it | Counted in over-specification if asserted on |
| **Ambiguous** | Insufficient evidence to classify — could be either | Excluded from both metrics |

## Confidence Scoring

Each side effect receives a **confidence score** from 0 to 100. The score determines the label:

- **Score >= 75** (default contractual threshold): **Contractual**
- **Score >= 50 and < 75**: **Ambiguous**
- **Score < 50** (default incidental threshold): **Incidental**

These thresholds are configurable via `.gaze.yaml` (see [Configuring Thresholds](#configuring-thresholds) below).

### How the Score Is Computed

The confidence score starts at a **base value** that depends on the effect's [tier](../reference/glossary.md#tier), then accumulates evidence from five signal analyzers, applies a contradiction penalty if conflicting signals exist, and clamps to the 0–100 range.

#### Step 1: Base + Tier Boost

Every effect starts at a base confidence of **50**. A tier-based boost is added:

| Tier | Boost | Effective Starting Score | Rationale |
|---|---|---|---|
| P0 | +25 | **75** | P0 effects (returns, errors, mutations) are definitionally contractual — they are a function's direct observable outputs |
| P1 | +10 | **60** | P1 effects (channels, writers, globals) are frequently contractual but context-dependent |
| P2–P4 | +0 | **50** | Higher-tier effects genuinely depend on context for classification |

This means P0 effects reach the default contractual threshold (75) with no additional signals. A `ReturnValue` effect is contractual by default — you need *negative* evidence to push it below the threshold.

#### Step 2: Signal Accumulation

Each of the five signal analyzers contributes a weighted signal (positive or negative). Signals with zero weight or empty source are skipped. The weights are added to the running score.

#### Step 3: Contradiction Penalty

If both positive and negative signals are present (e.g., the function name suggests contractual but the godoc says "logs"), a **contradiction penalty of -20** is applied. This pushes conflicting evidence toward the ambiguous range, reflecting genuine uncertainty.

#### Step 4: Clamping

The final score is clamped to the range [0, 100].

## The Five Signal Analyzers

### 1. Interface Satisfaction (max weight: +30)

Checks whether the function's receiver type satisfies any interface defined in the module. When a method appears in an interface, its side effects are strong contractual evidence — the interface defines the contract.

**Example:** If `(*Store).Save` satisfies `Repository.Save`, the `ReceiverMutation` effect of `Save` receives a +30 signal.

**Weight:** +30 when the method satisfies an interface that declares it; 0 otherwise.

### 2. API Surface Visibility (max weight: +20)

Evaluates whether the side effect is observable through the exported API. Three dimensions contribute independently:

| Dimension | Weight | Condition |
|---|---|---|
| Exported function | +8 | The function itself is exported (starts with uppercase) |
| Exported return type | +6 | At least one return type is exported |
| Exported receiver type | +6 | The receiver type is exported |

The total is capped at +20. An exported method on an exported type with exported return types receives the full +20.

**Weight:** 0 to +20 depending on how many dimensions match.

### 3. Caller Dependency (max weight: +15)

Scans all packages in the module for call sites that reference the target function. More callers means more code depends on the function's behavior, strengthening the contractual case.

| Caller Count | Weight |
|---|---|
| 0 | 0 (no signal) |
| 1 | +5 |
| 2–3 | +10 |
| 4+ | +15 |

**Weight:** 0 to +15 based on the number of distinct packages that call the function.

### 4. Naming Convention (max weight: +10 / -10, sentinel: +30)

Matches the function name against Go community naming conventions. Certain prefixes strongly imply contractual or incidental behavior.

**Contractual prefixes** (weight: +10 when the effect type matches the prefix's implied effects):

| Prefix | Implied Effect Types |
|---|---|
| `Get`, `Fetch`, `Load`, `Read` | `ReturnValue`, `ErrorReturn` |
| `Save`, `Write`, `Update` | `ReceiverMutation`, `PointerArgMutation`, `ErrorReturn` |
| `Set` | `ReceiverMutation`, `PointerArgMutation` |
| `Delete`, `Remove` | `ReceiverMutation`, `ErrorReturn` |
| `Handle`, `Process` | All effect types |
| `Compute`, `Analyze`, `Classify`, `Parse`, `Build`, `New` | `ReturnValue`, `ErrorReturn` |

**Incidental prefixes** (weight: -10):
`log`, `Log`, `debug`, `Debug`, `trace`, `Trace`, `print`, `Print`

**Sentinel error naming** (weight: +30): Variables with the `Err` prefix and `SentinelError` type receive a boosted +30 weight. Sentinel errors are unambiguously contractual by convention — they are exported, named with the `Err` prefix, and exist solely to be matched by callers. The higher weight ensures sentinels reach the contractual threshold even without other signals (since package-level variables cannot receive interface, visibility, or godoc signals).

### 5. GoDoc Comment (max weight: +15 / -15)

Parses the function's documentation comment for behavioral declarations.

**Contractual keywords** (weight: +15 when the effect type matches, +5 when a contractual keyword is found but the effect type doesn't directly match):

| Keyword | Implied Effect Types |
|---|---|
| `returns` | `ReturnValue`, `ErrorReturn` |
| `writes`, `modifies`, `updates`, `sets`, `persists`, `stores` | `ReceiverMutation`, `PointerArgMutation` |
| `deletes`, `removes` | `ReceiverMutation` |

**Incidental keywords** (weight: -15):
`logs`, `prints`, `traces`, `debugs`

## Worked Example

Consider an exported method `(*Store).Save` that has two detected side effects:

1. **`ErrorReturn`** (P0): The function returns an error
2. **`ReceiverMutation`** (P0): The function mutates `s.data`

For the `ErrorReturn` effect:

| Step | Value | Running Score |
|---|---|---|
| Base | 50 | 50 |
| Tier boost (P0) | +25 | 75 |
| Interface signal (`Repository.Save`) | +30 | 105 |
| Visibility signal (exported function + exported receiver) | +14 | 119 |
| Caller signal (3 callers) | +10 | 129 |
| Naming signal (`Save` prefix implies `ErrorReturn`) | +10 | 139 |
| GoDoc signal ("persists" keyword) | +5 | 144 |
| Contradiction penalty | 0 (no negative signals) | 144 |
| Clamp to [0, 100] | | **100** |

**Result:** Contractual (confidence 100 >= 75)

For a `LogWrite` effect on a function named `logRequest`:

| Step | Value | Running Score |
|---|---|---|
| Base | 50 | 50 |
| Tier boost (P2) | +0 | 50 |
| Naming signal (`log` prefix) | -10 | 40 |
| GoDoc signal ("logs" keyword) | -15 | 25 |
| Contradiction penalty | 0 (only negative signals) | 25 |

**Result:** Incidental (confidence 25 < 50)

## Configuring Thresholds

The default thresholds (contractual >= 75, incidental < 50) can be adjusted in `.gaze.yaml`:

```yaml
classification:
  thresholds:
    contractual: 75   # Score at or above this = contractual
    incidental: 50     # Score below this = incidental
                       # Scores in [incidental, contractual) = ambiguous
```

Lowering the contractual threshold makes more effects contractual (stricter contract coverage requirements). Raising the incidental threshold makes more effects incidental (more lenient contract coverage).

## What's Next

- [Scoring](scoring.md) — how classification feeds into CRAP and GazeCRAP scores
- [Quality Assessment](quality.md) — how contract coverage and over-specification are computed from classified effects
- [Side Effects](side-effects.md) — the full taxonomy of 37 effect types
