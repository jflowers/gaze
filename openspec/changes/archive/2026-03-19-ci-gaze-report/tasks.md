# Tasks: CI Gaze Report

## 1. Add coverprofile to test step

- [x] 1.1 In `.github/workflows/test.yml`, change the Test step's `run` command from `go test -race -count=1 -short ./...` to `go test -race -count=1 -short -coverprofile=coverage.out ./...`. This applies to both Go versions (the coverprofile is harmless on Go 1.25 even though it won't be consumed).

## 2. Add gaze report step

- [x] 2.1 In `.github/workflows/test.yml`, add a new step after "Test" in the `unit-and-integration` job named "Gaze quality report" with `if: matrix.go-version == '1.24'`. The step should run: `go run ./cmd/gaze report ./... --coverprofile=coverage.out --max-crapload=16 --max-gaze-crapload=5`.

## 3. Verification

- [x] 3.1 Verify the workflow YAML is valid by checking indentation, step ordering, and `if` condition syntax. Ensure the `if` condition uses string comparison: `matrix.go-version == '1.24'`.
- [ ] 3.2 Commit, push, and verify CI passes (the gaze report step runs on Go 1.24 and thresholds are satisfied).
