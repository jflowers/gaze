## 1. Change runSubprocess return signature

- [x] 1.1 Change `runSubprocess` in `internal/aireport/adapter.go` to return `([]byte, []byte, error)` — stdout, stderr, error. On success (exit 0), return `(outBytes, stderrBuf.Bytes(), nil)`. On non-zero exit, preserve existing behavior: return `(nil, nil, error)` with stderr embedded in the error message.
- [x] 1.2 Update `subprocess_test.go` tests to handle the new 3-value return from `runSubprocess`.

## 2. Update adapters to surface stderr in empty-output errors

- [x] 2.1 Update `OpenCodeAdapter.Format` in `adapter_opencode.go`: capture stderr from `runSubprocess`, and when stdout is empty, append truncated stderr (at `maxAdapterStderrBytes`) to the FR-009 error message if stderr contains non-whitespace content.
- [x] 2.2 Update `ClaudeAdapter.Format` in `adapter_claude.go`: same pattern — capture stderr and include in FR-016 empty-output error when stderr is non-empty.
- [x] 2.3 Update `GeminiAdapter.Format` in `adapter_gemini.go`: same pattern — capture stderr and include in FR-016 empty-output error when stderr is non-empty.

## 3. Update adapter tests

- [x] 3.1 Update `adapter_opencode_test.go`: fix all call sites for new `runSubprocess` return signature. Update the empty-output test to verify stderr appears in the error message.
- [x] 3.2 Update `adapter_claude_test.go`: fix all call sites for new `runSubprocess` return signature.
- [x] 3.3 Update `adapter_gemini_test.go`: fix all call sites for new `runSubprocess` return signature.

## 4. Verification

- [x] 4.1 Run `go build ./cmd/gaze` to verify compilation.
- [x] 4.2 Run `go test -race -count=1 -short ./internal/aireport/...` to verify all adapter tests pass.
- [x] 4.3 Run `go vet ./...` and `golangci-lint run` to verify no lint issues.
- [x] 4.4 Verify constitution alignment: Observable Quality principle is satisfied — error messages now include diagnostic context from subprocess stderr.
