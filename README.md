# Gaze

**Test quality analysis via side effect detection for Go.**

Line coverage tells you which lines ran. It does not tell you whether your tests actually verified anything.

A function can have 90% line coverage and tests that assert on nothing contractually meaningful — logging calls, goroutine lifecycle, internal stdout writes — while leaving the return values, error paths, and state mutations completely unverified. That function is dangerous to change, and traditional coverage metrics will not warn you.

Gaze fixes this by working from first principles:

1. **Detect** every observable side effect a function produces (return values, error returns, mutations, I/O, channel sends, etc.).
2. **Classify** each effect as *contractual* (part of the function's public obligation), *incidental* (an implementation detail), or *ambiguous*.
3. **Measure** whether your tests actually assert on the contractual effects — and flag the ones they don't.

This produces three actionable metrics:

### Contract Coverage

The percentage of a function's contractual side effects that at least one test assertion verifies.

```
ContractCoverage% = (contractual effects asserted on / total contractual effects) × 100
```

A function with 90% line coverage but 20% contract coverage has tests that run code without checking correctness. Gaze surfaces the specific effects that have no assertion — the exact gaps you need to close.

| Range | Status |
|-------|--------|
| ≥ 80% | Good |
| 50–79% | Warning |
| < 50%  | Bad |

### Over-Specification Score

The count and ratio of test assertions that target *incidental* effects — implementation details that are not part of the function's contract.

```
OverSpec.Ratio = incidental assertions / total mapped assertions
```

Tests that assert on log output, goroutine lifecycle, or internal stdout will break during refactoring even when the function's actual contract is preserved. Gaze identifies each over-specified assertion and explains why it is fragile.

| Count | Status |
|-------|--------|
| 0 | Good |
| 1–3 | Warning |
| > 3 | Bad |

### GazeCRAP

A composite risk score that replaces line coverage with contract coverage in the CRAP (Change Risk Anti-Patterns) formula.

```
CRAP(m)     = complexity² × (1 − lineCoverage/100)³ + complexity
GazeCRAP(m) = complexity² × (1 − contractCoverage)³ + complexity
```

Functions are placed in a quadrant based on both scores:

| | Low GazeCRAP | High GazeCRAP |
|---|---|---|
| **Low CRAP** | Safe | Simple but underspecified |
| **High CRAP** | Complex but tested | **Dangerous** |

The **Dangerous** quadrant — complex functions whose tests don't verify their contracts — is the highest-priority target for remediation.

---

Gaze requires no annotations, no test framework changes, and no restructuring of your code. It analyzes your existing Go packages as-is.

## Installation

### Homebrew (recommended)

```bash
brew install unbound-force/tap/gaze
```

### Go Install

```bash
go install github.com/unbound-force/gaze/cmd/gaze@latest
```

### Build from Source

```bash
git clone https://github.com/unbound-force/gaze.git
cd gaze
go build -o gaze ./cmd/gaze
```

Requires Go 1.25.0 or later.

### macOS Code Signing

Homebrew binaries are code-signed with an Apple Developer ID certificate and notarized by Apple's notary service. macOS Gatekeeper trusts the binary on first run -- no security overrides needed.

**For maintainers**: Signing requires 5 GitHub secrets (Apple Developer ID certificate + App Store Connect API key). See [quickstart guide](specs/014-macos-notarization/quickstart.md) for setup instructions. When secrets are not configured, the release pipeline produces unsigned binaries without error.

## Commands

### `gaze analyze` -- Side Effect Detection

Analyze a Go package to detect all observable side effects each function produces.

```bash
# Analyze all exported functions in a package
gaze analyze ./internal/analysis

# Analyze a specific function
gaze analyze -f ParseConfig ./internal/config

# Include unexported functions
gaze analyze --include-unexported ./internal/loader

# JSON output
gaze analyze --format=json ./internal/analysis

# Classify side effects as contractual, incidental, or ambiguous
gaze analyze --classify ./internal/analysis

# Verbose classification with full signal breakdown
gaze analyze --verbose ./internal/analysis

# Interactive TUI for browsing results
gaze analyze -i ./internal/analysis

# Use a config file with custom thresholds
gaze analyze --classify --config=.gaze.yaml --contractual-threshold=90 ./internal/analysis
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--function` | `-f` | Analyze a specific function (default: all exported) |
| `--format` | | Output format: `text` or `json` (default: `text`) |
| `--include-unexported` | | Include unexported functions |
| `--interactive` | `-i` | Launch interactive TUI for browsing results |
| `--classify` | | Classify side effects as contractual, incidental, or ambiguous |
| `--verbose` | `-v` | Print full signal breakdown (implies `--classify`) |
| `--config` | | Path to `.gaze.yaml` config file (default: search CWD) |
| `--contractual-threshold` | | Override contractual confidence threshold (default: from config or 80) |
| `--incidental-threshold` | | Override incidental confidence threshold (default: from config or 50) |

**Detected side effect types:**

| Tier | Effects |
|------|---------|
| P0 | `ReturnValue`, `ErrorReturn`, `SentinelError`, `ReceiverMutation`, `PointerArgMutation` |
| P1 | `SliceMutation`, `MapMutation`, `GlobalMutation`, `WriterOutput`, `HTTPResponseWrite`, `ChannelSend`, `ChannelClose`, `DeferredReturnMutation` |
| P2 | `FileSystemWrite`, `FileSystemDelete`, `FileSystemMeta`, `DatabaseWrite`, `DatabaseTransaction`, `GoroutineSpawn`, `Panic`, `CallbackInvocation`, `LogWrite`, `ContextCancellation` |
| P3* | `StdoutWrite`, `StderrWrite`, `EnvVarMutation`, `MutexOp`, `WaitGroupOp`, `AtomicOp`, `TimeDependency`, `ProcessExit`, `RecoverBehavior` |
| P4* | `ReflectionMutation`, `UnsafeMutation`, `CgoCall`, `FinalizerRegistration`, `SyncPoolOp`, `ClosureCaptureMutation` |

*P3 and P4 types are defined in the taxonomy but detection is not yet implemented.*

Example output:

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

### `gaze crap` -- CRAP Score Analysis

Compute CRAP scores by combining cyclomatic complexity with test coverage.

```bash
# Analyze all packages
gaze crap ./...

# Use an existing coverage profile
gaze crap --coverprofile=cover.out ./...

# Custom thresholds
gaze crap --crap-threshold=20 ./...
gaze crap --gaze-crap-threshold=20 ./...

# CI mode: fail if too many crappy functions
gaze crap --max-crapload=5 ./...

# JSON output
gaze crap --format=json ./...
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--format` | Output format: `text` or `json` (default: `text`) |
| `--coverprofile` | Path to existing coverage profile (default: generate one) |
| `--crap-threshold` | CRAP score threshold (default: 15) |
| `--gaze-crap-threshold` | GazeCRAP score threshold, used when contract coverage is available (default: 15) |
| `--max-crapload` | Fail if CRAPload exceeds this count (0 = no limit) |
| `--max-gaze-crapload` | Fail if GazeCRAPload exceeds this count (0 = no limit) |

**CRAP formula:**

```text
CRAP(m) = complexity^2 * (1 - coverage/100)^3 + complexity
```

A function with complexity 5 and 0% coverage has CRAP = 30. The same function with 100% coverage has CRAP = 5. The default threshold is 15.

Example output:

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

### `gaze quality` -- Test Quality Assessment

Assess how well a package's tests assert on the contractual side effects of the functions they test. Reports Contract Coverage (ratio of contractual effects that are asserted on) and Over-Specification Score (assertions on incidental implementation details).

```bash
# Analyze test quality for a package
gaze quality ./internal/analysis

# Target a specific function
gaze quality --target=LoadAndAnalyze ./internal/analysis

# Verbose output with detailed assertion and mapping information
gaze quality --verbose ./internal/analysis

# JSON output
gaze quality --format=json ./internal/analysis

# CI mode: enforce minimum contract coverage
gaze quality --min-contract-coverage=80 --max-over-specification=3 ./internal/analysis

# Custom classification thresholds
gaze quality --config=.gaze.yaml --contractual-threshold=90 ./internal/analysis
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--format` | | Output format: `text` or `json` (default: `text`) |
| `--target` | | Restrict analysis to tests that exercise this function |
| `--verbose` | `-v` | Show detailed assertion and mapping information |
| `--config` | | Path to `.gaze.yaml` config file (default: search CWD) |
| `--contractual-threshold` | | Override contractual confidence threshold (default: from config or 80) |
| `--incidental-threshold` | | Override incidental confidence threshold (default: from config or 50) |
| `--min-contract-coverage` | | Fail if contract coverage is below this percentage (0 = no limit) |
| `--max-over-specification` | | Fail if over-specification count exceeds this (0 = no limit) |

#### How Target Inference Works

Gaze automatically determines which function each test exercises by walking the test function's SSA (Static Single Assignment) call graph. A function is considered a "target" when:

1. It is in the same package as the test (minus the `_test` suffix)
2. It is not a test, benchmark, init, or closure function
3. It is called from the test body (including `t.Run` sub-tests)

**Common causes of "multiple target functions detected" warnings:**

- **Setup helpers called from test body**: Functions like `writeCredentials()`, `createTempDir()`, or `newMockServer()` are detected as targets because they're called from the test body and live in the same package. To fix: move setup to `TestMain` or extract to a `testutil_test.go` file with functions that don't match the target package path.

- **Multi-function integration tests**: Tests that call `Create()`, then `Read()`, then `Delete()` in sequence exercise multiple targets. Gaze reports all detected targets and the contract coverage is diluted across them. To fix: use sub-tests (`t.Run`) where each sub-test exercises one function.

**What does NOT affect target detection:**

- `t.Helper()` — this affects stack frame reporting in test failures, not SSA call graph analysis. Marking a function as a helper does **not** prevent it from being detected as a target.
- Function naming conventions — Gaze does **not** use `TestFoo` → `Foo` name matching. It relies entirely on SSA call graph analysis.
- Assertion libraries — the assertion library used (stdlib, testify, go-cmp) does not affect target detection.

**Best practices for clear target inference:**

1. Name test functions to match their target: `TestFoo_Scenario` exercises `Foo`
2. Keep setup in helpers that accept `*testing.T` — Gaze detects these as helpers, not targets, when they only set up fixtures
3. Use `t.Run` sub-tests for integration tests that call multiple functions
4. Use `--target=FuncName` flag to restrict analysis to a specific target when automatic detection is ambiguous

### `gaze schema` -- JSON Schema Output

Print the JSON Schema (Draft 2020-12) that documents the structure of `gaze analyze --format=json` output. Useful for validating output or generating client types.

```bash
gaze schema
```

### `gaze docscan` -- Documentation Scanner

Scan the repository for Markdown documentation files and output a prioritized list as JSON. Useful as input to the gaze-reporter agent's full mode for document-enhanced classification.

Files are prioritized by proximity to the target package:

1. Same directory as the target package (highest relevance)
2. Module root
3. Other locations

```bash
# Scan from current directory
gaze docscan

# Scan for a specific package
gaze docscan ./internal/analysis

# Use a config file
gaze docscan --config=.gaze.yaml ./internal/analysis
```

### `gaze init` -- OpenCode Integration Setup

Scaffold OpenCode agent and command files into the current project directory for AI-assisted quality reporting.

```bash
# Initialize OpenCode integration
gaze init

# Overwrite existing files
gaze init --force
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Overwrite existing OpenCode files |

This creates 2 files in `.opencode/`:

- `.opencode/agents/gaze-reporter.md` -- Quality report agent
- `.opencode/command/gaze.md` -- `/gaze` command

### `gaze report` -- AI-Powered Quality Report

Orchestrate all four gaze analysis operations (CRAP, quality, classification, docscan) and pipe the combined JSON payload to an external AI CLI for formatting into a human-readable markdown report. Optionally appends the report to `$GITHUB_STEP_SUMMARY` for visibility in the GitHub Actions UI.

```bash
# Generate a report using claude
gaze report ./... --ai=claude

# Use gemini with a specific model
gaze report ./... --ai=gemini --model=gemini-2.5-pro

# Use local ollama (model required)
gaze report ./... --ai=ollama --model=llama3.2

# Use opencode (uses your configured default model)
gaze report ./... --ai=opencode

# Use opencode with a specific model
gaze report ./... --ai=opencode --model=claude-3-5-sonnet

# JSON output (no AI required)
gaze report ./... --format=json

# CI mode: fail if quality thresholds are breached
gaze report ./... --ai=claude --max-crapload=10 --max-gaze-crapload=5 --min-contract-coverage=60
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--ai` | AI adapter: `claude`, `gemini`, `ollama`, or `opencode` (required in text mode) |
| `--model` | Model name (required for ollama; optional for claude/gemini/opencode) |
| `--ai-timeout` | AI adapter timeout (default: `10m`) |
| `--format` | Output format: `text` or `json` (default: `text`) |
| `--coverprofile` | Path to a pre-generated coverage profile (skips internal go test run) |
| `--max-crapload` | Fail if project CRAPload exceeds N (zero is a live threshold) |
| `--max-gaze-crapload` | Fail if GazeCRAPload exceeds N (zero is a live threshold) |
| `--min-contract-coverage` | Fail if average contract coverage is below N% |

**GitHub Actions example:**

```yaml
- name: Gaze quality report
  run: gaze report ./... --ai=claude --max-crapload=15 --min-contract-coverage=50
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
    GITHUB_STEP_SUMMARY: ${{ github.step_summary }}
```

When `$GITHUB_STEP_SUMMARY` is set (as in GitHub Actions), the formatted report is appended to the workflow step summary. Write failures are non-fatal — the command exits 0 with a warning on stderr.

#### Using a pre-generated coverage profile

If your CI workflow already runs `go test -coverprofile`, pass that profile to `gaze report` to avoid running tests twice:

```bash
go test -race -count=1 -coverprofile=coverage.out ./...
gaze report ./... --ai=claude --coverprofile=coverage.out
```

With this setup, tests run **once** per CI job. The coverage data used for CRAP scores is the same high-quality profile produced by the race-detecting test run.

**GitHub Actions example with pre-generated profile:**

```yaml
- name: Test
  run: go test -race -count=1 -coverprofile=coverage.out ./...

- name: Gaze Report
  run: gaze report ./... --ai=claude --coverprofile=coverage.out
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
    GITHUB_STEP_SUMMARY: ${{ github.step_summary }}
```

### `gaze self-check` -- Self-Analysis

Run CRAP analysis on Gaze's own source code, serving as both a dogfooding exercise and a code quality gate.

```bash
# Run self-check
gaze self-check

# JSON output
gaze self-check --format=json

# CI mode: enforce limits
gaze self-check --max-crapload=5 --max-gaze-crapload=3
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--format` | Output format: `text` or `json` (default: `text`) |
| `--max-crapload` | Fail if CRAPload exceeds this count (0 = no limit) |
| `--max-gaze-crapload` | Fail if GazeCRAPload exceeds this count (0 = no limit) |

### CI Integration

Use threshold flags for CI enforcement. Gaze exits non-zero when limits are exceeded and prints a one-line summary to stderr:

```bash
gaze crap --max-crapload=5 --max-gaze-crapload=3 ./...
# stderr: CRAPload: 2/5 (PASS) | GazeCRAPload: 1/3 (PASS)
```

Without threshold flags, Gaze always exits 0 (report-only mode).

## Output Formats

The `analyze`, `crap`, `quality`, and `self-check` commands support `--format=text` (default) and `--format=json`.

JSON output conforms to documented schemas. Use `gaze schema` to print the analysis report schema. The schemas are embedded in the binary at `internal/report/schema.go`.

## OpenCode Integration

After running `gaze init`, use the `/gaze` command in OpenCode for AI-assisted quality reporting:

```text
/gaze ./...                     # Full report: CRAP + quality + classification
/gaze crap ./internal/store     # CRAP scores only
/gaze quality ./pkg/api         # Test quality metrics only
```

The `gaze-reporter` agent runs gaze CLI commands with `--format=json`, interprets the output, and produces human-readable markdown summaries with actionable recommendations.

## Architecture

```text
cmd/gaze/              CLI entry point (cobra)
internal/
  analysis/            Side effect detection engine
    analyzer.go        Main analysis orchestrator
    returns.go         Return value analysis (AST)
    sentinel.go        Sentinel error detection (AST)
    mutation.go        Receiver/pointer mutation (SSA)
    p1effects.go       P1-tier effects (AST)
    p2effects.go       P2-tier effects (AST)
  taxonomy/            Side effect type system and stable IDs
  classify/            Contractual classification engine
  config/              Configuration file handling (.gaze.yaml)
  loader/              Go package loading wrapper
  report/              JSON and text formatters for analysis output
  crap/                CRAP score computation and reporting
  quality/             Test quality assessment (contract coverage)
  docscan/             Documentation file scanner
  scaffold/            OpenCode file scaffolding (embed.FS)
```

## Known Limitations

- **Direct function body only.** Gaze analyzes the immediate function body. Transitive side effects (effects produced by called functions) are out of scope for v1.
- **P3-P4 side effects not yet detected.** The taxonomy defines types for stdout/stderr writes, environment mutations, mutex operations, reflection, unsafe, and other P3-P4 effects, but detection logic is not yet implemented for these tiers.
- **GazeCRAP accuracy is limited.** The quality pipeline is wired into the CRAP command and GazeCRAP scores are computed when contract coverage data is available. However, assertion-to-side-effect mapping accuracy is currently ~86% (target: 90%), primarily affecting cross-target assertions and go-cmp patterns (tracked as GitHub Issue #6).
- **No CGo or unsafe analysis.** Functions using `cgo` or `unsafe.Pointer` are not analyzed for their specific side effects.
- **Single package loading.** The `analyze` command processes one package at a time. Use shell loops or scripting for multi-package analysis.

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
