## ADDED Requirements

### Requirement: [MI-REQ-001] Canonical v2 Module Identity

The project MUST declare `github.com/unbound-force/gaze/v2` as its Go module
path and MUST use that path for every module-local Go import.

#### Scenario: [MI-SC-001] Build uses the v2 identity
- **GIVEN** a clean checkout of the migrated repository
- **WHEN** the Go toolchain loads and builds all packages
- **THEN** the main module MUST be identified as `github.com/unbound-force/gaze/v2`
- **AND** project packages MUST NOT import one another through the unsuffixed v1 path

#### Scenario: [MI-SC-002] Tests use the same identity
- **GIVEN** tests and fixtures that refer to project package paths
- **WHEN** the full test suite runs after the migration
- **THEN** package expectations and coverage fixtures MUST use the `/v2` module path
- **AND** the tests MUST NOT depend on a second copy of the project under the v1 identity

### Requirement: [MI-REQ-002] Non-Module URLs Remain Stable

Repository URLs, GitHub links, Homebrew references, GoReleaser metadata, and
JSON Schema identifiers MUST remain unsuffixed unless the value specifically
represents a Go module or package import path.

#### Scenario: [MI-SC-003] Repository link is audited
- **GIVEN** a string beginning with `https://github.com/unbound-force/gaze`
- **WHEN** the migration determines whether to append `/v2`
- **THEN** the string MUST remain unchanged when it identifies the repository or a repository resource
- **AND** `/v2` MUST be added only when the string identifies the Go module or package path

### Requirement: [MI-REQ-003] Major-Version Compatibility Boundary

The v2 module MUST NOT add a compatibility shim for the unsuffixed v1 import
path. Existing v1 consumers continue using `github.com/unbound-force/gaze`,
while v2 consumers MUST use `github.com/unbound-force/gaze/v2`.

#### Scenario: [MI-SC-004] Consumer selects a major version
- **GIVEN** a consumer that imports Gaze packages
- **WHEN** the consumer selects v2 behavior
- **THEN** the consumer MUST import `github.com/unbound-force/gaze/v2/...`
- **AND** the v2 repository MUST NOT masquerade as the unsuffixed v1 module

### Requirement: [MI-REQ-004] Module Migration Regression Ratchet

Automated regression checks MUST require `go list -m` to return the exact `/v2`
identity, reject current module-local imports through the unsuffixed path, verify
package and coverage fixtures, and preserve all existing build, race, unit,
integration, E2E, lint, security, and embedded-copy synchronization gates.
The retained guidance audit MUST also pass.

#### Scenario: [MI-SC-005] Mechanical migration is regression-tested
- **GIVEN** the module declaration, imports, and fixtures have migrated
- **WHEN** CI-equivalent regression checks run
- **THEN** every current module/package reference MUST use the `/v2` identity
- **AND** all pre-existing quality gates MUST pass without weakened thresholds or flags
