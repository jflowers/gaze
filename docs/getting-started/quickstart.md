# Quickstart

Get from zero to meaningful analysis in under 10 minutes. This guide walks you through installing Gaze and running its three core commands on your own Go project.

## 1. Install Gaze

The fastest method is Homebrew:

```bash
brew install unbound-force/tap/gaze
```

Verify it works:

```bash
gaze --version
```

For other installation methods (Go install, build from source), see the [Installation guide](installation.md).

## 2. Analyze Side Effects

Navigate to your Go project and run:

```bash
gaze analyze ./...
```

This scans every exported function and reports the observable [side effects](../reference/glossary.md) each one produces.

**Example output:**

```text
=== ParseConfig ===
    func ParseConfig(path string) (*Config, error)
    internal/config/config.go:15:1

    TIER  TYPE         DESCRIPTION
    ----  ----         -----------
    P0    ReturnValue  returns *Config at position 0
    P0    ErrorReturn  returns error at position 1

    Summary: P0: 2
```

**What to look for:**

- **Tier** indicates priority. [P0 effects](../reference/glossary.md) (return values, error returns, mutations) are almost always contractual -- they are what callers depend on.
- **Type** names the specific effect. See the [side effects reference](../concepts/side-effects.md) for the full taxonomy.
- Functions with many effects across multiple tiers are more complex to test thoroughly.

To also see how effects are classified (contractual, incidental, or ambiguous), add `--classify`:

```bash
gaze analyze --classify ./internal/config
```

## 3. Compute CRAP Scores

[CRAP](../reference/glossary.md) (Change Risk Anti-Patterns) combines cyclomatic complexity with test coverage to identify risky functions:

```bash
gaze crap ./...
```

> **Note:** If you don't pass `--coverprofile`, Gaze runs `go test -coverprofile` automatically. This may take a few minutes on large projects. To skip the wait, generate a coverage profile first and pass it in:
>
> ```bash
> go test -coverprofile=coverage.out ./...
> gaze crap --coverprofile=coverage.out ./...
> ```

**Example output:**

```text
CRAP    COMPLEXITY  COVERAGE  FUNCTION       FILE
----    ----------  --------  --------       ----
30.0 *  5           0.0%      ParseConfig    internal/config/config.go:15
5.0     5           100.0%    FormatOutput   internal/report/text.go:20

--- Summary ---
Functions analyzed:  2
Avg complexity:     5.0
Avg line coverage:  50.0%
Avg CRAP score:     17.5
CRAP threshold:     15
CRAPload:           1 (functions at or above threshold)
```

**What to look for:**

- Functions marked with `*` exceed the [CRAP threshold](../reference/glossary.md) (default: 15). These are your riskiest functions to change.
- **[CRAPload](../reference/glossary.md)** is the count of functions above the threshold -- the single number that summarizes project risk.
- High complexity + low coverage = high CRAP score. The fix is to either reduce complexity (refactor) or increase coverage (add tests).
- When [contract coverage](../reference/glossary.md) data is available, the output also shows **GazeCRAP** scores and [quadrant](../reference/glossary.md) assignments (Q1 Safe through Q4 Dangerous).

For the full scoring model, see [Scoring](../concepts/scoring.md).

## 4. Assess Test Quality

The `quality` command goes deeper than CRAP -- it maps your test assertions to the specific side effects they verify:

```bash
gaze quality ./internal/config
```

> **Note:** `quality` operates on a single package (not `./...`). Point it at the package you want to assess.

**Example output:**

```text
=== ParseConfig ===
    Contract Coverage: 50.0%
    Over-Specification: 0

    EFFECT                    STATUS
    ------                    ------
    ReturnValue (*Config)     ASSERTED
    ErrorReturn (error)       NOT ASSERTED

    Gaps:
      - ErrorReturn: no test assertion verifies the error return
```

**What to look for:**

- **[Contract Coverage](../reference/glossary.md)** is the percentage of contractual effects that your tests assert on. Higher is better. Below 50% is a warning sign.
- **[Over-Specification](../reference/glossary.md)** counts assertions on incidental effects (implementation details). Ideally zero -- these assertions break during refactoring without catching real bugs.
- **Gaps** list the specific contractual effects that no test verifies. These are the exact things you need to add assertions for.

For details on how Gaze maps assertions to effects, see [Test Quality](../concepts/quality.md).

## 5. What Next?

You've now seen Gaze's three core commands. Here's where to go from here:

| Goal | Guide |
|------|-------|
| Set up CI quality gates | [CI Integration](../guides/ci-integration.md) |
| Fix the issues Gaze found | [Improving Scores](../guides/improving-scores.md) |
| Generate AI-powered reports | [AI Reports](../guides/ai-reports.md) |
| Understand the scoring model | [Scoring](../concepts/scoring.md) |
| Look up all flags for a command | [CLI Reference](../reference/cli/analyze.md) |
| Learn the full side effect taxonomy | [Side Effects](../concepts/side-effects.md) |

### Quick Reference: Common Flags

```bash
# JSON output (any command)
gaze analyze --format=json ./...
gaze crap --format=json ./...

# CI mode: fail on threshold violations
gaze crap --max-crapload=5 ./...
gaze quality --min-contract-coverage=80 ./internal/config

# Analyze a specific function
gaze analyze -f ParseConfig ./internal/config

# Include unexported functions
gaze analyze --include-unexported ./internal/config
```

For the complete flag reference for each command, see the [CLI Reference](../reference/cli/analyze.md) pages.
