# Tasks: Quality Include Unexported

## 1. Add flag to quality command

- [x] 1.1 Add `includeUnexported bool` field to `qualityParams` struct in `cmd/gaze/main.go` (~line 670).
- [x] 1.2 Register `--include-unexported` flag in `newQualityCmd()` in `cmd/gaze/main.go` (~line 930) with description `"include unexported functions"` matching the `analyze` command's wording.
- [x] 1.3 Wire the flag in `runQuality`: change `analysis.Options{IncludeUnexported: false}` to `analysis.Options{IncludeUnexported: p.includeUnexported}` at `cmd/gaze/main.go:692`.

## 2. Add package main auto-detection

- [x] 2.1 In `runQuality` (`cmd/gaze/main.go`), after `loader.Load(p.pkgPath)`, check if `result.Pkg.Name() == "main"`. If so, set `opts.IncludeUnexported = true` and log `"package main detected, including unexported functions"`.
- [x] 2.2 In `runQualityForPackage` (`internal/aireport/runner_steps.go` ~line 138), after loading the package, check if it's `package main` and set `IncludeUnexported = true` on the `analysis.Options`. Write diagnostic to `stderr`.
- [x] 2.3 In `analyzePackageCoverage` (`internal/crap/contract.go` ~line 149), after the `analysis.LoadAndAnalyze` call, add auto-detection. Since `LoadAndAnalyze` takes the options upfront, the detection needs to happen before the call. Use a pre-flight package name check: load package metadata with `packages.NeedName` mode, check if `pkg.Name == "main"`, then set `IncludeUnexported = true` on the `analysis.Options` before calling `LoadAndAnalyze`.

## 3. Tests

- [x] 3.1 Add a `package main` test fixture at `internal/quality/testdata/src/mainpkg/` with a few unexported functions that have test files. The functions should have clear contractual side effects (return values, error returns) so quality analysis produces meaningful reports.
- [x] 3.2 Add `TestRunQuality_IncludeUnexported_PackageMain` in `cmd/gaze/main_test.go` that runs `runQuality` against the `mainpkg` fixture without `--include-unexported` and verifies unexported functions are found (auto-detect).
- [x] 3.3 Add `TestRunQuality_IncludeUnexported_LibraryPackage` that runs against a non-main package without the flag and verifies only exported functions appear.

## 4. Verification

- [x] 4.1 Run `go build ./cmd/gaze && go test -race -count=1 -short ./cmd/gaze/... ./internal/aireport/... ./internal/crap/...` to verify all tests pass.
- [x] 4.2 Run `go vet ./...` to verify no issues.
