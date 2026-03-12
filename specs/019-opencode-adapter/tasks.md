# Tasks: OpenCode AI Adapter for gaze report

**Input**: Design documents from `specs/019-opencode-adapter/`
**Branch**: `019-opencode-adapter`
**Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)

**Organization**: Tasks are grouped by user story. All implementation tasks also include their test tasks because the spec explicitly requires the fake-binary test coverage strategy (plan.md Coverage Strategy, constitution Principle IV: Testability).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2)
- All paths are relative to the repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the one new directory needed before any source files can be written.

- [x] T001 Create directory `internal/aireport/testdata/fake_opencode/` (required before T002)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The fake binary must exist before adapter tests can compile it. It must be written before any test file references it.

**⚠️ CRITICAL**: T002 must be complete before any test file in Phase 3 or Phase 4 can pass.

- [x] T002 Create fake opencode test binary at `internal/aireport/testdata/fake_opencode/main.go` — accepts `run` subcommand, `--dir` (verifies `.opencode/agents/gaze-reporter.md` exists there and starts with `---`), `--agent`, `--format`, `-m` (echoed in output), `--exit-error` (exits 1), `--empty-output` (writes nothing); reads stdin; writes `"# Fake OpenCode Report\n\n🔍 CRAP Analysis\n\n📊 Quality\n\n🧪 Classification\n\n🏥 Health\n"` to stdout on success

**Checkpoint**: Fake binary written — test compilation can proceed.

---

## Phase 3: User Story 1 — Generate Report with OpenCode (Priority: P1) 🎯 MVP

**Goal**: `gaze report --ai=opencode` invokes the `opencode run` subprocess, delivers the system prompt as `.opencode/agents/gaze-reporter.md` in a temp dir, delivers the payload via stdin, reads plain-text stdout, and returns a formatted markdown report.

**Independent Test**: `go test -race -count=1 -run TestOpenCodeAdapter ./internal/aireport/...` passes all 10 adapter tests using the fake binary. `go test -race -count=1 -run TestSC006_CrossAdapterStructure ./cmd/gaze/...` passes with `"opencode"` included.

### Implementation for User Story 1

- [x] T003 [US1] Create `internal/aireport/adapter_opencode.go` — define `OpenCodeAdapter` struct with `config AdapterConfig` field; add compile-time interface check `var _ AdapterValidator = &OpenCodeAdapter{}`; implement `ValidateBinary()` calling `exec.LookPath("opencode")` with error `"opencode not found on PATH (FR-007): %w"`; implement `Format(ctx, systemPrompt, payload)` per plan.md Implementation Design: LookPath defense, `os.MkdirTemp("", "gaze-opencode-*")`, `os.MkdirAll(.opencode/agents, 0700)`, write `"---\n---\n"+systemPrompt` to `gaze-reporter.md` at `0600`, build args `["run", "--dir", tmpDir, "--agent", "gaze-reporter", "--format", "default", ""]` with optional `-m model`, `exec.CommandContext(ctx, ...)`, stdin=payload, bounded stdout pipe (`maxAdapterOutputBytes`), stderr buffer, `cmd.Start()`, `io.ReadAll(LimitReader)`, `cmd.Wait()`, stderr truncation at `maxAdapterStderrBytes`, `strings.TrimSpace` empty-output check with `"opencode returned empty output (FR-009): ..."`, `defer os.RemoveAll(tmpDir)`

- [x] T004 [US1] Update `internal/aireport/adapter.go` — add `"opencode": true` to `validAdapters` map; add `case "opencode": return &OpenCodeAdapter{config: cfg}, nil` to `NewAdapter()` switch; update error string to `"must be one of \"claude\", \"gemini\", \"ollama\", or \"opencode\""`

- [x] T005 [P] [US1] Create `internal/aireport/adapter_opencode_test.go` — implement all 10 tests using `buildFakeOpenCode(t)` / `withOpenCodeOnPath(t, bin)` helpers (same pattern as `adapter_claude_test.go`): `TestOpenCodeAdapter_SuccessfulInvocation`, `TestOpenCodeAdapter_AgentFileWrittenToTempDir` (fake binary verifies agent file exists — a passing call proves delivery), `TestOpenCodeAdapter_FrontmatterWritten` (fake binary reads agent file, asserts `---` prefix), `TestOpenCodeAdapter_ModelFlagPassed` (cfg.Model="test-model"; assert output contains "test-model"), `TestOpenCodeAdapter_NoModelFlag_Succeeds` (cfg.Model=""; assert success and no `-m` in args), `TestOpenCodeAdapter_NotOnPath_ReturnsError` (no Short guard; assert "FR-007" in error), `TestOpenCodeAdapter_NonZeroExit_ReturnsError` (shell wrapper injects `--exit-error`), `TestOpenCodeAdapter_EmptyOutput_ReturnsError` (shell wrapper injects `--empty-output`; assert "FR-009" in error), `TestOpenCodeAdapter_TempDirCleanedUp` (count `gaze-opencode-*` entries before/after; assert no leak), `TestOpenCodeAdapter_ContextCancellation` (pre-cancelled ctx; assert error); all subprocess tests guarded with `if testing.Short() { t.Skip(...) }`

**Checkpoint**: Run `go test -race -count=1 -run TestOpenCodeAdapter ./internal/aireport/...` — all 10 tests pass. Run `go build ./...` — compiles cleanly.

---

## Phase 4: User Story 2 — Consistent Behavior with Other Adapters (Priority: P2)

**Goal**: The `opencode` adapter is wired into the full CLI pipeline — `--ai=opencode` is accepted by `gaze report`, error messages list it, usage examples show it, and `TestSC006_CrossAdapterStructure` verifies structural parity with all other adapters.

**Independent Test**: `go test -race -count=1 -run TestSC006_CrossAdapterStructure ./cmd/gaze/...` passes with `"opencode"` in the adapter loop. `go test -race -count=1 -run TestRunReport_InvalidAI ./cmd/gaze/...` passes (error message includes `"opencode"`). `go test -race -count=1 -short ./...` passes with no regressions.

### Implementation for User Story 2

- [x] T006 [US2] Update `cmd/gaze/main.go` — make three targeted string changes: (1) `--ai` flag description from `"AI adapter: claude, gemini, or ollama"` to `"AI adapter: claude, gemini, ollama, or opencode"`; (2) required-flag error string from `"must be one of \"claude\", \"gemini\", or \"ollama\""` to `"must be one of \"claude\", \"gemini\", \"ollama\", or \"opencode\""`; (3) add `gaze report ./... --ai=opencode` and `gaze report ./... --ai=opencode --model=claude-3-5-sonnet` to the command usage examples block (locate the existing examples block near `newReportCmd` at line ~1327)

- [x] T007 [US2] Update `cmd/gaze/main_test.go` — extend `TestSC006_CrossAdapterStructure` adapter loop: change `[]string{"claude", "gemini", "ollama"}` to `[]string{"claude", "gemini", "ollama", "opencode"}` (located at line ~1661)

**Checkpoint**: Run `go test -race -count=1 -run TestSC006_CrossAdapterStructure ./cmd/gaze/...` — passes for all four adapters. Run `go test -race -count=1 -short ./...` — no regressions.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, CI parity validation, and spec bookkeeping.

- [x] T008 [P] Update `README.md` if it contains a list of supported `--ai` providers — add `opencode` alongside `claude`, `gemini`, `ollama`; if no such list exists, skip this task

- [x] T009 [P] Update `AGENTS.md` Recent Changes section — add a bullet for `019-opencode-adapter` describing the new `opencode` adapter (mirrors the existing bullet pattern for spec 018)

- [x] T010 Add GoDoc comment to `OpenCodeAdapter` type in `internal/aireport/adapter_opencode.go` — ensure it explains the temp dir structure, the `--dir` / `--agent` delivery mechanism, and references `FR-007` and `FR-009`; also ensure `Format()` and `ValidateBinary()` have complete GoDoc comments matching the style of `ClaudeAdapter` and `GeminiAdapter`

- [x] T011 CI parity gate — run `go build ./...` and `go test -race -count=1 -short ./...` locally; confirm both pass with zero failures; this replicates the exact commands from `.github/workflows/test.yml`

- [x] T012 Mark all completed tasks in this file with `[x]` and update `specs/019-opencode-adapter/tasks.md` to reflect final state

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (directory must exist)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (fake binary must exist before test file compiles it)
  - T003, T004, T005 can proceed together once T002 is complete (different files, no inter-dependencies)
- **User Story 2 (Phase 4)**: Depends on T004 (adapter must be in allowlist before CLI wiring makes sense); can proceed once Phase 3 is underway
- **Polish (Phase 5)**: Depends on all implementation tasks (T003–T007) being complete

### User Story Dependencies

- **US1 (Phase 3)**: Depends on Phase 2 only. T003, T004, T005 are all independent of each other (different files).
- **US2 (Phase 4)**: T006 depends on T004 (allowlist must exist). T007 depends on T004 (adapter name must be valid for the test to pass). US2 does not depend on US1 test results but logically follows US1 implementation.

### Within Each Phase

- T003 (adapter implementation) before T005 (tests) for logical order, but both can be written in parallel since they are in different files — tests should be written to fail first per TDD discipline if preferred
- T004 (allowlist update) is a prerequisite for T006 and T007
- T008 and T009 are independent of each other and of T010

### Parallel Opportunities

- **Phase 3**: T003, T004, T005 — all different files, no blocking dependencies between them
- **Phase 4**: T006, T007 — different files within `cmd/gaze/`
- **Phase 5**: T008, T009, T010 — all different files, fully independent

---

## Parallel Example: User Story 1

```bash
# All three can be launched together once T002 is complete:
Task T003: "Create internal/aireport/adapter_opencode.go"
Task T004: "Update internal/aireport/adapter.go (allowlist + factory)"
Task T005: "Create internal/aireport/adapter_opencode_test.go (10 tests)"
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete Phase 1: Create directory (T001)
2. Complete Phase 2: Write fake binary (T002)
3. Complete Phase 3: Implement adapter + update allowlist + write tests (T003, T004, T005)
4. **STOP and VALIDATE**: `go test -race -count=1 -run TestOpenCodeAdapter ./internal/aireport/...`
5. If all 10 tests pass: MVP is working — `gaze report --ai=opencode` is functional

### Full Delivery (Both Stories)

1. MVP (above)
2. Phase 4: Wire into CLI and extend SC006 (T006, T007)
3. Phase 5: Documentation + CI parity gate (T008–T012)
4. Run full `go test -race -count=1 -short ./...` — zero failures
5. PR ready

---

## Notes

- [P] tasks = different files, no dependencies on each other within the same phase
- [Story] label maps task to specific user story for traceability
- All subprocess-compiling tests are guarded with `testing.Short()` — they run under `go test -race -count=1 ./...` (full suite) but are skipped under `go test -race -count=1 -short ./...` (CI fast suite)
- `TestOpenCodeAdapter_NotOnPath_ReturnsError` does NOT need `testing.Short()` guard — it uses no subprocess
- Commit after each phase checkpoint to preserve clean history
- The fake binary must verify the agent file starts with `---` to confirm frontmatter is written (FR-003 + clarification answer from session 2026-03-12)
- No changes to `runner.go`, `runner_steps.go`, `payload.go`, `threshold.go`, `output.go`, or any other existing files beyond the three listed
