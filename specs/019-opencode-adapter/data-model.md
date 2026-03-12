# Data Model: OpenCode AI Adapter for gaze report

**Branch**: `019-opencode-adapter`
**Date**: 2026-03-12

This feature introduces no new persistent data. All entities are ephemeral runtime constructs. This document captures the adapter's type model and the lifecycle of the temporary agent file.

---

## Entity: `OpenCodeAdapter`

**Package**: `internal/aireport`
**File**: `adapter_opencode.go`

| Field | Type | Notes |
|-------|------|-------|
| `config` | `AdapterConfig` | Embedded adapter config (Name, Model, Timeout, OllamaHost) |

**Implements**: `AIAdapter` (Format method), `AdapterValidator` (ValidateBinary method)

**Relationships**:
- Created by `NewAdapter(cfg AdapterConfig)` when `cfg.Name == "opencode"`
- Consumed by `aireport.Run(RunnerOptions)` via the `AIAdapter` interface
- `AdapterConfig.Model` is optional; when empty, no `-m` flag is passed to opencode

---

## Entity: `AdapterConfig` (existing, extended)

**Package**: `internal/aireport`
**File**: `adapter.go`

No new fields. The existing `Model` field serves the opencode adapter without modification.

| Field | Type | Meaning for opencode |
|-------|------|----------------------|
| `Name` | `string` | `"opencode"` |
| `Model` | `string` | Optional; passed as `-m <name>` to `opencode run` when non-empty |
| `Timeout` | `time.Duration` | Applied via `exec.CommandContext(ctx, ...)` — inherited from caller |
| `OllamaHost` | `string` | Unused by opencode adapter |

**Allowlist update**: `validAdapters` map in `adapter.go` gains `"opencode": true`.

---

## Entity: Ephemeral Agent File (runtime only)

**Not a Go type** — a filesystem artifact created and destroyed within a single `Format()` call.

| Attribute | Value |
|-----------|-------|
| Base directory | `os.MkdirTemp("", "gaze-opencode-*")` |
| Subdirectory | `.opencode/agents/` (created inside temp dir) |
| Filename | `gaze-reporter.md` |
| Full path | `<tmpDir>/.opencode/agents/gaze-reporter.md` |
| Permissions | `0600` (owner read/write only) |
| Content | Empty YAML frontmatter (`---\n---\n`) + system prompt body |
| Lifetime | Created at start of `Format()`; removed by `defer os.RemoveAll(tmpDir)` |

**Why this structure**: `opencode run --dir <tmpDir> --agent gaze-reporter` resolves the agent file at `<tmpDir>/.opencode/agents/gaze-reporter.md`. The subdirectory path `.opencode/agents/` must exist for opencode to locate the agent.

---

## Entity: `validAdapters` map (existing, extended)

**Package**: `internal/aireport`
**File**: `adapter.go`

| Key | Value | Status |
|-----|-------|--------|
| `"claude"` | `true` | Existing |
| `"gemini"` | `true` | Existing |
| `"ollama"` | `true` | Existing |
| `"opencode"` | `true` | **New** |

---

## Process Flow: `OpenCodeAdapter.Format(ctx, systemPrompt, payload)`

```
Format(ctx, systemPrompt, payload io.Reader) (string, error)
│
├── exec.LookPath("opencode")          → error if not on PATH (FR-007)
├── os.MkdirTemp("", "gaze-opencode-*") → tmpDir
├── os.MkdirAll(tmpDir/.opencode/agents, 0700)
├── os.WriteFile(tmpDir/.opencode/agents/gaze-reporter.md,
│               "---\n---\n" + systemPrompt, 0600)
├── args = ["run", "--dir", tmpDir, "--agent", "gaze-reporter",
│           "--format", "default", ""]
│   + optional ["-m", cfg.Model]
├── exec.CommandContext(ctx, opencodePath, args...)
│   cmd.Stdin = payload
│   cmd.Stdout → stdoutPipe (bounded by io.LimitReader 64 MiB)
│   cmd.Stderr → stderrBuf
├── cmd.Start()
├── io.ReadAll(io.LimitReader(stdoutPipe, 64 MiB)) → outBytes
├── cmd.Wait()
│   ├── waitErr != nil → error (exit error + truncated stderr ≤512B)
│   └── readErr != nil → error
├── strings.TrimSpace(string(outBytes)) == "" → FR-009 error
├── defer os.RemoveAll(tmpDir)
└── return string(outBytes), nil
```

---

## State Transitions: Temp Directory Lifecycle

```
[Not created]
    │  Format() called
    ▼
[tmpDir created]
    │  subdirectory + agent file written
    ▼
[Agent file ready]
    │  subprocess started
    ▼
[Subprocess running]  ──── ctx cancelled ──►  [Subprocess killed]
    │                                               │
    │  subprocess exits 0                           │
    ▼                                               ▼
[Output read]                              [Error returned]
    │                                               │
    └─────────────────────┬─────────────────────────┘
                          │  defer os.RemoveAll(tmpDir)
                          ▼
                   [tmpDir deleted]
```

The `defer os.RemoveAll(tmpDir)` runs in all exit paths — success, subprocess error, read error, and context cancellation.

---

## No New Persistent State

This feature introduces no database tables, no configuration file changes, no new environment variables, and no new persistent files. The only state is the `validAdapters` map entry and the `NewAdapter` switch case, both compile-time constants.
