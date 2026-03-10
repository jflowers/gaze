# Quickstart: AI-Powered CI Quality Report (`gaze report`)

**Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)

---

## What is `gaze report`?

`gaze report` produces the same rich quality report that the `/gaze` OpenCode command generates interactively, but from the command line — in CI or locally. It:

1. Runs all four gaze analysis operations (`crap`, `quality`, `analyze --classify`, `docscan`)
2. Passes the results to your chosen AI CLI (`claude`, `gemini`, or `ollama`) with a formatting prompt
3. Writes the formatted markdown report to stdout
4. When running in GitHub Actions, also appends the report to the Step Summary tab

---

## Prerequisites

- `gaze` installed (`brew install unbound-force/tap/gaze` or `go install github.com/unbound-force/gaze/cmd/gaze@latest`)
- One of the supported AI CLIs installed and authenticated:
  - **claude**: `npm install -g @anthropic-ai/claude-code` + `ANTHROPIC_API_KEY` env var
  - **gemini**: `npm install -g @google/gemini-cli` + `GEMINI_API_KEY` env var
  - **ollama**: `brew install ollama` + `ollama serve` running + model pulled

---

## GitHub Actions — Full Report (P1 Use Case)

```yaml
- name: Run Gaze Quality Report
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: gaze report ./... --ai=claude
```

After this step runs, the formatted quality report appears in the **Summary** tab of the workflow run.

---

## GitHub Actions — With Quality Gate (P2 Use Case)

```yaml
- name: Run Gaze Quality Report
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: |
    gaze report ./... --ai=claude \
      --max-crapload=10 \
      --max-gaze-crapload=5 \
      --min-contract-coverage=70
```

If any threshold is breached, the step exits with code 1 and a summary line appears on stderr:

```
CRAPload: 13/10 (FAIL) | GazeCRAPload: 3/5 (PASS) | AvgContractCoverage: 65/70 (FAIL)
```

---

## Local Use (P3 Use Case)

```bash
# Full report to terminal
gaze report ./... --ai=claude

# Gemini instead of claude
GEMINI_API_KEY=... gaze report ./... --ai=gemini

# Ollama (local model, no internet required)
ollama serve &
ollama pull llama3.2
gaze report ./... --ai=ollama --model=llama3.2

# Raw JSON output (no AI required)
gaze report ./... --format=json | jq .summary
```

---

## Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--ai` | *(required)* | AI adapter: `claude`, `gemini`, or `ollama` |
| `--model` | *(adapter default)* | Model name (required for `ollama`) |
| `--ai-timeout` | `10m` | Max wait time for AI response |
| `--max-crapload` | *(disabled)* | Fail if CRAPload exceeds N (0 = fail if any > threshold) |
| `--max-gaze-crapload` | *(disabled)* | Fail if GazeCRAPload exceeds N |
| `--min-contract-coverage` | *(disabled)* | Fail if avg contract coverage < N% |
| `--format` | `text` | `text` = AI-formatted report; `json` = raw analysis data |

---

## Environment Variables

| Variable | Description |
|---|---|
| `ANTHROPIC_API_KEY` | Required for `--ai=claude` |
| `GEMINI_API_KEY` | Required for `--ai=gemini` |
| `OLLAMA_HOST` | Override ollama server URL (default: `http://localhost:11434`) |
| `GITHUB_STEP_SUMMARY` | Set automatically by GitHub Actions; when present, report is also written here |

---

## Custom Formatting Prompt

By default, `gaze report` uses the embedded prompt from the gaze binary (derived from the gaze-reporter agent). To customize the formatting, add or edit `.opencode/agents/gaze-reporter.md` in your project root (created by `gaze init`). The local file takes precedence over the embedded default.

---

## `--format=json` Mode

In JSON mode, no AI CLI is invoked. The raw combined analysis data is written to stdout as a single JSON object:

```bash
gaze report ./... --format=json > analysis.json
```

Output schema:
```json
{
  "crap":     { "scores": [...], "summary": {...} },
  "quality":  { "quality_reports": [...], "quality_summary": {...} },
  "classify": { "version": "...", "results": [...] },
  "docscan":  [ { "path": "...", "content": "...", "priority": 1 } ],
  "errors":   { "crap": null, "quality": null, "classify": null, "docscan": null }
}
```

A non-null `errors` field value indicates that pipeline step failed (partial report). The other steps' data is still present.

---

## Acceptance Test Verification (for implementers)

| SC | Verified by |
|---|---|
| SC-001 | `TestSC001_GithubActionsReport` — sets `GITHUB_STEP_SUMMARY` to temp file, verifies file written with expected section markers |
| SC-002 | `TestSC002_ReportStructure` — uses FakeAdapter returning known markdown, verifies section header patterns `🔍`, `📊`, `🧪`, `🏥` present |
| SC-003 | `TestSC003_ThresholdTiming` — measures threshold evaluation duration: < 1ms |
| SC-004 | `TestSC004_PartialFailure` — simulates CRAP analysis failure, verifies report produced with `> ⚠️` warning and exit 0 |
| SC-005 | `BenchmarkReportAnalysis` — measures analysis phase on gaze module itself; enforced ≤ 5m |
| SC-006 | Manual verification — run with claude, gemini, ollama; verify all four section markers present |
