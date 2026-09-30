## ADDED Requirements

### Requirement: [DOC-REQ-001] Current v2 Go Installation Command

Active documentation MUST use
`go install github.com/unbound-force/gaze/v2/cmd/gaze@latest` for corrected
stable v2 releases and MUST retain Homebrew and GitHub Releases as alternatives.

#### Scenario: [DOC-SC-001] User follows Go installation guidance
- **GIVEN** a corrected stable v2 release is available
- **WHEN** a user follows the documented Go installation method
- **THEN** the command MUST use the `/v2` module path

### Requirement: [DOC-REQ-002] Synchronized Gaze-Owned Guidance

The source reporter prompt and its embedded scaffold/runtime copies MUST contain
the same corrected installation guidance.

#### Scenario: [DOC-SC-002] Embedded prompt copies are verified
- **GIVEN** the three Gaze-owned prompt copies
- **WHEN** synchronization tests run
- **THEN** their installation guidance MUST match

### Requirement: [DOC-REQ-003] Historical Records Stay Historical

Completed historical specifications MUST remain unchanged unless they are
consumed as current executable guidance.

#### Scenario: [DOC-SC-003] Old command exists in a historical spec
- **GIVEN** a completed historical artifact records the old command
- **WHEN** active guidance is migrated
- **THEN** the historical artifact MUST remain unchanged

### Requirement: [DOC-REQ-004] Website Documentation Is Tracked

A human-created website issue MUST track the user-facing installation update
before merge.

#### Scenario: [DOC-SC-004] Website follow-up is recorded
- **GIVEN** installation behavior changes
- **WHEN** implementation is handed off
- **THEN** the website issue URL MUST be recorded
