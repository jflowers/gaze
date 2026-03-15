## 1. Prompt Updates

- [x] 1.1 Add Remediation Breakdown instruction (item 8) to CRAP mode section in `internal/scaffold/assets/agents/gaze-reporter.md`.
- [x] 1.2 Add Fix Strategy Awareness block to the Top 5 Recommendations section, mapping each strategy to specific guidance with Read tool instruction for testability assessment.
- [x] 1.3 Sync `.opencode/agents/gaze-reporter.md` and `internal/aireport/assets/agents/gaze-reporter.md` with the embedded asset to satisfy scaffold and aireport drift detection tests.

## 2. Verification

- [x] 2.1 Run `TestEmbeddedAssetsMatchSource` to verify drift parity.
- [x] 2.2 Run `go build ./...` and `go vet ./...`.
- [x] 2.3 Run `go test -race -count=1 -short ./...`.
