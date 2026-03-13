## 1. Dependency Upgrade

- [x] 1.1 Run `go get golang.org/x/tools@v0.43.0` to upgrade the SSA dependency
- [x] 1.2 Run `go mod tidy` to regenerate `go.sum` and update transitive deps (`x/mod`, `x/sync`, `x/sys`)
- [x] 1.3 Verify `go.mod` declares `go 1.25.0`, `golang.org/x/tools v0.43.0`, and updated transitive versions

## 2. CI Matrix Expansion

- [x] 2.1 Add `strategy.matrix.go-version: ["1.24", "1.25"]` to the `unit-and-integration` job in `.github/workflows/test.yml`
- [x] 2.2 Replace the hardcoded `go-version: "1.24"` in the `unit-and-integration` job's `setup-go` step with `${{ matrix.go-version }}`
- [x] 2.3 Update the `unit-and-integration` job `name` to include `(Go ${{ matrix.go-version }})` for readability in the Actions UI
- [x] 2.4 Add `strategy.matrix.go-version: ["1.24", "1.25"]` to the `e2e` job in `.github/workflows/test.yml`
- [x] 2.5 Replace the hardcoded `go-version: "1.24"` in the `e2e` job's `setup-go` step with `${{ matrix.go-version }}`
- [x] 2.6 Update the `e2e` job `name` to include `(Go ${{ matrix.go-version }})` for readability in the Actions UI

## 3. Local Verification (CI Parity Gate)

- [x] 3.1 Run `go build ./...` and confirm it succeeds
- [x] 3.2 Run `go test -race -count=1 -short ./...` and confirm all tests pass
- [x] 3.3 Run `golangci-lint run` and confirm no new lint errors

## 4. Constitution Alignment Verification

- [x] 4.1 Confirm Principle I (Accuracy): `gaze report` produces correct output after the upgrade — no panic, no silent degradation
- [x] 4.2 Confirm Principle II (Minimal Assumptions): no new annotation or user-code restructuring is required; minimum Go version assumption is explicit in `go.mod`
- [x] 4.3 Confirm Principle III (Actionable Output): output format and content are unchanged
- [x] 4.4 Confirm Principle IV (Testability): CI matrix now covers Go 1.25; existing test suite validates the upgrade
<!-- scaffolded by unbound vdev -->
