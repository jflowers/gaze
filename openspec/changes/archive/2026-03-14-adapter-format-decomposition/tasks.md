## 1. Create runSubprocess helper

- [x] 1.1 Add `runSubprocess` function to `internal/aireport/adapter.go` with signature `func runSubprocess(ctx context.Context, binaryName string, args []string, cmdDir string, payload io.Reader) ([]byte, error)`. Implement: `exec.LookPath`, `exec.CommandContext`, `StdoutPipe`, stderr buffer, `Start`, `io.ReadAll(io.LimitReader(..., maxAdapterOutputBytes))`, `Wait`, error handling with stderr truncation at `maxAdapterStderrBytes`.
- [x] 1.2 Add GoDoc comment on `runSubprocess` explaining its role as shared subprocess execution for the three CLI-based AI adapters.

## 2. Refactor adapter Format methods

- [x] 2.1 Refactor `ClaudeAdapter.Format` (`internal/aireport/adapter_claude.go`) to use `runSubprocess`. Keep: temp dir + `prompt.md` creation, arg construction (`-p`, `--system-prompt-file`, `--model`), empty-output check (FR-016). Remove: inline `LookPath`, `StdoutPipe`, stderr buffer, `Start`, `ReadAll`, `Wait`, error handling.
- [x] 2.2 Refactor `GeminiAdapter.Format` (`internal/aireport/adapter_gemini.go`) to use `runSubprocess` with `cmdDir` set to `tmpDir`. Keep: temp dir + `GEMINI.md` creation, arg construction (`-p`, `--output-format json`, `-m`), JSON unmarshal of `geminiOutput`, empty-output check (FR-016). Remove: inline subprocess blocks.
- [x] 2.3 Refactor `OpenCodeAdapter.Format` (`internal/aireport/adapter_opencode.go`) to use `runSubprocess`. Keep: temp dir + nested `.opencode/agents/gaze-reporter.md` with YAML frontmatter, arg construction (`run`, `--dir`, `--agent`, `--format default`, `-m`), empty-output check (FR-009). Remove: inline subprocess blocks.

## 3. Tests

- [x] 3.1 Create `internal/aireport/subprocess_test.go` with tests for `runSubprocess`.
- [x] 3.2 Add `TestRunSubprocess_Success` — verify successful binary execution returns stdout bytes.
- [x] 3.3 Add `TestRunSubprocess_BinaryNotFound` — verify error contains the binary name when binary doesn't exist.
- [x] 3.4 Add `TestRunSubprocess_NonZeroExit` — verify error includes stderr content when subprocess exits non-zero.
- [x] 3.5 Add `TestRunSubprocess_CmdDir` — verify subprocess runs in the specified working directory.
- [x] 3.6 Run all existing adapter tests to verify no regressions: `go test -race -count=1 ./internal/aireport/`
- [x] 3.7 Verify complexity reduction: run `gocyclo` on the three adapter files and confirm each Format method is at complexity 7 or below.

## 4. Documentation

- [x] 4.1 Update `AGENTS.md` Recent Changes with a summary of this change.

## 5. Verification

- [x] 5.1 Run `go test -race -count=1 -short ./...` and verify all tests pass.
- [x] 5.2 Run `go build ./...` and `go vet ./...` to verify no issues.
- [x] 5.3 Verify constitution alignment: (I) Autonomous Collaboration — no interface changes. (II) Composability — adapters remain standalone. (III) Observable Quality — no output changes. (IV) Testability — `runSubprocess` testable in isolation.
