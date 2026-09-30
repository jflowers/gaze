## ADDED Requirements

### Requirement: [VAL-REQ-001] Tag Major Matches Module Path

Before tag creation, the release workflow MUST derive the requested tag major
and require the exact canonical module path: `github.com/unbound-force/gaze`
for v0/v1, or `github.com/unbound-force/gaze/vN` for v2 and later.

#### Scenario: [VAL-SC-001] Matching v2 release proceeds
- **GIVEN** tag `v2.2.0` and module `github.com/unbound-force/gaze/v2`
- **WHEN** release validation runs
- **THEN** validation MUST pass

#### Scenario: [VAL-SC-002] Unsuffixed module blocks v2 release
- **GIVEN** a v2 tag and module `github.com/unbound-force/gaze`
- **WHEN** release validation runs
- **THEN** validation MUST fail before reusable preflight and tag creation
- **AND** the failure MUST report the expected and observed module paths

### Requirement: [VAL-REQ-002] Validation Is Deterministic and Unskippable

Module-major validation MUST run without network or GitHub API access and MUST
NOT be bypassed by existing release skip flags.

#### Scenario: [VAL-SC-003] Offline contract cases run
- **GIVEN** matching v1/v2, mismatched v2/v3, and malformed-tag inputs
- **WHEN** the offline validator tests run
- **THEN** all five observable status contracts MUST pass
