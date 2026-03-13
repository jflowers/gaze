## ADDED Requirements

### Requirement: Go 1.25 SSA Compatibility

`gaze report` MUST complete the quality (SSA) analysis phase without panicking
when the target module's transitive dependency graph includes
`github.com/go-json-experiment/json`, when run under Go 1.25.

#### Scenario: gaze report succeeds on a Go 1.25 codebase with go-json-experiment dep
- **GIVEN** gaze is installed under Go 1.25 and the target module transitively
  depends on `github.com/go-json-experiment/json`
- **WHEN** the user runs `gaze report ./... --ai=opencode --coverprofile=coverage.out`
- **THEN** the command SHALL complete the quality analysis phase and produce a
  report without panicking

#### Scenario: gaze report CRAP step is unaffected
- **GIVEN** gaze is run under Go 1.25
- **WHEN** the user runs `gaze crap ./...`
- **THEN** the command MUST succeed (it does not use SSA and MUST NOT be
  regressed by this change)

---

### Requirement: Go 1.25 CI Coverage

The gaze CI pipeline MUST run its full test suite against Go 1.25 on every
push and pull request to `main`, in addition to the existing Go 1.24 run.

#### Scenario: CI runs unit and integration tests on Go 1.25
- **GIVEN** a pull request is opened or a commit is pushed to `main`
- **WHEN** the CI `unit-and-integration` job executes
- **THEN** it SHALL run `go test -race -count=1 -short ./...` under both Go 1.24
  and Go 1.25, and both runs MUST pass for the job to succeed

#### Scenario: CI runs e2e tests on Go 1.25
- **GIVEN** a pull request is opened or a commit is pushed to `main`
- **WHEN** the CI `e2e` job executes
- **THEN** it SHALL run `go test -race -count=1 -run TestRunSelfCheck -timeout 30m ./cmd/gaze/...`
  under both Go 1.24 and Go 1.25, and both runs MUST pass for the job to succeed

## MODIFIED Requirements

### Requirement: Minimum Go Version

Previously: gaze required Go 1.24.2 or later to build from source.

gaze MUST now require Go 1.25.0 or later to build from source, as declared in
`go.mod`. This requirement is imposed by the transitive dependency
`golang.org/x/tools@v0.43.0`.

#### Scenario: Building gaze from source on Go 1.25
- **GIVEN** a developer has Go 1.25.0 or later installed
- **WHEN** they run `go build ./cmd/gaze`
- **THEN** the build MUST succeed without errors

#### Scenario: Pre-built binary users are unaffected
- **GIVEN** a user installs gaze via a GoReleaser-distributed binary (Homebrew,
  GitHub Releases, etc.)
- **WHEN** they run `gaze report`
- **THEN** the binary MUST work regardless of the Go toolchain version installed
  on their system (binaries are self-contained)

## REMOVED Requirements

None.
<!-- scaffolded by unbound vdev -->
