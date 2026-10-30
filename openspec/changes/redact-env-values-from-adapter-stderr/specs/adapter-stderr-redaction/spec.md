# Adapter Stderr Redaction Requirements

## ADDED Requirements

### Requirement: Surfaced adapter stderr MUST be scrubbed of sensitive inherited env values

Before any subprocess stderr content is embedded in an error message or
returned to a caller, `internal/aireport` MUST replace occurrences of
sensitive inherited environment-variable values with a redaction
marker. The scrub MUST cover both surfacing paths: the non-zero-exit
error from `runSubprocess`, and the success-path stderr bytes surfaced
via `formatStderrSuffix` in the claude, gemini, and opencode
empty-output errors.

#### Scenario: Non-zero exit echoes a config secret

- **GIVEN** the gaze process has an inherited env var carrying a
  sensitive value (e.g. `OPENCODE_CONFIG_CONTENT`) whose value appears
  in the AI CLI's stderr, and the CLI exits non-zero
- **WHEN** `runSubprocess` returns its error
- **THEN** the error message contains a redaction marker naming the
  env var and does not contain the secret value

#### Scenario: Empty-output error surfaces a success-path secret

- **GIVEN** the AI CLI exits 0 with empty stdout, and its stderr
  contains a sensitive inherited env value
- **WHEN** the adapter builds its empty-output error via
  `formatStderrSuffix`
- **THEN** the appended stderr snippet contains a redaction marker and
  does not contain the value

### Requirement: A value MUST be treated as sensitive by name marker or length threshold

An inherited environment-variable value MUST be treated as sensitive
when the variable name contains `KEY`, `TOKEN`, `SECRET`, or
`CONFIG_CONTENT` (case-insensitive), or when the value is at or above
the configured length threshold regardless of name.

#### Scenario: Oddly-named long secret is caught by the threshold

- **GIVEN** an inherited env var whose name matches no sensitive name
  marker but whose value length is at or above the threshold, and the
  value appears in surfaced stderr
- **WHEN** the stderr is surfaced in an error
- **THEN** the value is replaced with a redaction marker

#### Scenario: Lowercase sensitive name is matched

- **GIVEN** an inherited env var whose lowercase name contains a
  sensitive marker (e.g. `openai_api_key`), and its value appears in
  surfaced stderr
- **WHEN** the stderr is surfaced in an error
- **THEN** the value is replaced with a redaction marker

### Requirement: Scrubbing MUST run before size truncation

The scrub MUST be applied to the full stderr buffer before the
`maxAdapterStderrBytes` truncation, so that a sensitive value
straddling the truncation boundary cannot leak a prefix into the
surfaced snippet.

#### Scenario: Secret straddles the truncation boundary

- **GIVEN** child stderr longer than `maxAdapterStderrBytes` in which a
  sensitive value starts before the truncation offset and extends past
  it
- **WHEN** the snippet is surfaced
- **THEN** no prefix of the sensitive value appears in the snippet (the
  redaction marker that replaced it may itself be cut by the byte cap)

### Requirement: Scrubbing MUST preserve error diagnosability

The redaction marker MUST identify the env var whose value was
scrubbed (env var names are not secret). Values shorter than the
minimum replacement length MUST NOT be replaced, preventing garbled
error text from short common substrings; sub-minimum-length secrets
are a documented limitation of this scrub.

#### Scenario: Marker names the scrubbed variable

- **GIVEN** a sensitive env var `OPENCODE_CONFIG_CONTENT` whose value
  is scrubbed from surfaced stderr
- **WHEN** the surfaced error is inspected
- **THEN** the redaction marker identifies `OPENCODE_CONFIG_CONTENT`
  by name

#### Scenario: Short values are not replaced

- **GIVEN** an inherited env var matching a sensitive name marker whose
  value is shorter than the minimum replacement length
- **WHEN** stderr containing that short value is surfaced
- **THEN** the value is not replaced and the surrounding error text is
  unchanged

### Requirement: The CI integration guide MUST document consumer-side masking mitigations

`docs/guides/ci-integration.md` MUST document, for reusable-workflow
consumers: (a) registering the bare API key as a second secret
referenced in an unused env var of the caller's job, so GitHub masks
verbatim occurrences anywhere in logs, and (b) preferring a
single-line strict-JSON secret value, because GitHub's multiline secret
masking is unreliable. The guide SHOULD also describe gaze's automatic
adapter-stderr scrubbing and its marker format.

#### Scenario: Consumer reads the mitigation guidance

- **GIVEN** a consumer configuring the reusable workflow with the
  `opencode-config` secret
- **WHEN** they read the Reusable Workflow section of
  `docs/guides/ci-integration.md`
- **THEN** the section explains the second-secret registration
  pattern, the single-line strict-JSON preference, and gaze's stderr
  scrubbing behavior

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
