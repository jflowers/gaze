# Quickstart: OpenCode AI Adapter for gaze report

**Branch**: `019-opencode-adapter`
**Date**: 2026-03-12

A quick reference for developers implementing and testing the `opencode` adapter.

---

## What Changes

| File | Type | Change |
|------|------|--------|
| `internal/aireport/adapter_opencode.go` | New | `OpenCodeAdapter` struct + `Format()` + `ValidateBinary()` |
| `internal/aireport/adapter_opencode_test.go` | New | 10 test cases using fake binary pattern |
| `internal/aireport/testdata/fake_opencode/main.go` | New | Fake `opencode` binary for tests |
| `internal/aireport/adapter.go` | Modified | Add `"opencode"` to `validAdapters` + `NewAdapter` switch |
| `cmd/gaze/main.go` | Modified | Update `--ai` help text, error message, usage examples |
| `cmd/gaze/main_test.go` | Modified | Extend `TestSC006_CrossAdapterStructure` with `"opencode"` |

---

## Implementing `OpenCodeAdapter`

### File: `internal/aireport/adapter_opencode.go`

Structural outline (full implementation in tasks):

```go
package aireport

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

// OpenCodeAdapter invokes the opencode CLI to format the analysis payload.
// The system prompt is written as .opencode/agents/gaze-reporter.md in a
// temporary directory passed via --dir. This mirrors the Gemini adapter's
// temp-dir pattern adapted for opencode's agent file convention.
type OpenCodeAdapter struct {
    config AdapterConfig
}

var _ AdapterValidator = &OpenCodeAdapter{}

func (a *OpenCodeAdapter) ValidateBinary() error {
    if _, err := exec.LookPath("opencode"); err != nil {
        return fmt.Errorf("opencode not found on PATH (FR-007): %w", err)
    }
    return nil
}

func (a *OpenCodeAdapter) Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error) {
    opencodePath, err := exec.LookPath("opencode")
    if err != nil {
        return "", fmt.Errorf("opencode not found on PATH (FR-007): %w", err)
    }

    tmpDir, err := os.MkdirTemp("", "gaze-opencode-*")
    if err != nil {
        return "", fmt.Errorf("creating temp dir for opencode agent: %w", err)
    }
    defer func() { _ = os.RemoveAll(tmpDir) }()

    agentsDir := filepath.Join(tmpDir, ".opencode", "agents")
    if err := os.MkdirAll(agentsDir, 0700); err != nil {
        return "", fmt.Errorf("creating .opencode/agents dir: %w", err)
    }

    agentFile := filepath.Join(agentsDir, "gaze-reporter.md")
    content := "---\n---\n" + systemPrompt
    if err := os.WriteFile(agentFile, []byte(content), 0600); err != nil {
        return "", fmt.Errorf("writing agent file: %w", err)
    }

    args := []string{"run", "--dir", tmpDir, "--agent", "gaze-reporter", "--format", "default", ""}
    if a.config.Model != "" {
        args = append(args, "-m", a.config.Model)
    }

    cmd := exec.CommandContext(ctx, opencodePath, args...)
    cmd.Stdin = payload

    stdoutPipe, err := cmd.StdoutPipe()
    if err != nil {
        return "", fmt.Errorf("creating stdout pipe for opencode: %w", err)
    }
    var stderrBuf bytes.Buffer
    cmd.Stderr = &stderrBuf

    if err := cmd.Start(); err != nil {
        return "", fmt.Errorf("starting opencode: %w", err)
    }

    outBytes, readErr := io.ReadAll(io.LimitReader(stdoutPipe, maxAdapterOutputBytes))
    waitErr := cmd.Wait()

    if waitErr != nil {
        stderrSnippet := stderrBuf.String()
        if len(stderrSnippet) > maxAdapterStderrBytes {
            stderrSnippet = stderrSnippet[:maxAdapterStderrBytes] + "... (truncated)"
        }
        return "", fmt.Errorf("opencode exited with error: %w\nstderr: %s", waitErr, stderrSnippet)
    }
    if readErr != nil {
        return "", fmt.Errorf("reading opencode output: %w", readErr)
    }

    result := string(outBytes)
    if strings.TrimSpace(result) == "" {
        return "", fmt.Errorf("opencode returned empty output (FR-009): ensure the opencode CLI is working correctly")
    }
    return result, nil
}
```

---

## Updating `adapter.go`

Two targeted changes:

```go
// In validAdapters map:
var validAdapters = map[string]bool{
    "claude":    true,
    "gemini":    true,
    "ollama":    true,
    "opencode":  true,  // NEW
}

// In NewAdapter switch:
case "opencode":
    return &OpenCodeAdapter{config: cfg}, nil

// In error message:
return nil, fmt.Errorf(
    "unknown AI adapter %q: must be one of \"claude\", \"gemini\", \"ollama\", or \"opencode\"",
    cfg.Name,
)
```

---

## Updating `cmd/gaze/main.go`

Three targeted changes:

```go
// 1. --ai flag help text:
cmd.Flags().StringVar(&adapterName, "ai", "", "AI adapter: claude, gemini, ollama, or opencode")

// 2. Required --ai error message:
"--ai is required in text mode: must be one of \"claude\", \"gemini\", \"ollama\", or \"opencode\""

// 3. Usage examples (in Use or Long field):
gaze report ./... --ai=opencode
gaze report ./... --ai=opencode --model=claude-3-5-sonnet
```

---

## Implementing `fake_opencode`

### File: `internal/aireport/testdata/fake_opencode/main.go`

The fake binary must:
1. Accept `run` as first positional arg (subcommand)
2. Parse `--dir`, `--agent`, `--format`, `-m`/`--model`, `--exit-error`, `--empty-output`
3. Verify `.opencode/agents/gaze-reporter.md` exists under `--dir`
4. Optionally verify the agent file starts with `---` (frontmatter check)
5. Read stdin (the JSON payload)
6. Write a fixed markdown report to stdout (or empty, or exit 1)

---

## Implementing `adapter_opencode_test.go`

Test structure mirrors `adapter_claude_test.go` and `adapter_gemini_test.go`:

```go
func buildFakeOpenCode(t *testing.T) string { ... }  // compile fake binary
func withOpenCodeOnPath(t *testing.T, bin string) { ... }  // PATH manipulation

func TestOpenCodeAdapter_SuccessfulInvocation(t *testing.T) { ... }
func TestOpenCodeAdapter_AgentFileWrittenToTempDir(t *testing.T) { ... }
func TestOpenCodeAdapter_FrontmatterWritten(t *testing.T) { ... }
func TestOpenCodeAdapter_ModelFlagPassed(t *testing.T) { ... }
func TestOpenCodeAdapter_NoModelFlag_Succeeds(t *testing.T) { ... }
func TestOpenCodeAdapter_NotOnPath_ReturnsError(t *testing.T) { ... }
func TestOpenCodeAdapter_NonZeroExit_ReturnsError(t *testing.T) { ... }
func TestOpenCodeAdapter_EmptyOutput_ReturnsError(t *testing.T) { ... }
func TestOpenCodeAdapter_TempDirCleanedUp(t *testing.T) { ... }
func TestOpenCodeAdapter_ContextCancellation(t *testing.T) { ... }
```

All subprocess tests use `testing.Short()` guard. `TestOpenCodeAdapter_NotOnPath_ReturnsError` does not need a subprocess and is NOT guarded.

---

## Extending `TestSC006_CrossAdapterStructure`

Add `"opencode"` to the adapter loop in `cmd/gaze/main_test.go`:

```go
for _, adapterName := range []string{"claude", "gemini", "ollama", "opencode"} {
```

No other change needed — the test already uses `FakeAdapter` and `runnerFunc` injection, bypassing the real adapter wiring.

---

## Running Tests

```bash
# Unit + integration (fast, no subprocess compilation):
go test -race -count=1 -short ./...

# Full suite including subprocess tests (adapter fake binaries):
go test -race -count=1 ./internal/aireport/... ./cmd/gaze/...

# Targeted adapter test only:
go test -race -count=1 -run TestOpenCodeAdapter ./internal/aireport/...

# SC006 cross-adapter structure test:
go test -race -count=1 -run TestSC006_CrossAdapterStructure ./cmd/gaze/...
```

---

## CI Parity Check

Before marking complete, run the exact CI commands from `.github/workflows/test.yml`:

```bash
go build ./...
go test -race -count=1 -short ./...
```

The e2e suite (`TestRunSelfCheck`) does not need to run locally for this feature — it tests the gaze-on-gaze self-analysis pipeline, which is unchanged.
