## ADDED Requirements

### Requirement: gaze quality MUST support --ai-mapper flag

The `gaze quality` command MUST accept an optional `--ai-mapper=<backend>` flag where `<backend>` is one of `claude`, `gemini`, `ollama`, or `opencode`. When set, the command MUST construct an `AIMapperFunc` callback that delegates to the specified AI adapter and pass it to `quality.Options.AIMapperFunc`.

#### Scenario: AI mapper enabled with Claude backend
- **GIVEN** a user runs `gaze quality --ai-mapper=claude ./...`
- **WHEN** an assertion cannot be mapped by any of the 4 mechanical passes
- **THEN** the AI mapper calls the Claude adapter with `BuildAIMapperPrompt` and maps the assertion at confidence 50 if the AI returns a valid effect ID

#### Scenario: AI mapper disabled by default
- **GIVEN** a user runs `gaze quality ./...` without `--ai-mapper`
- **WHEN** the pipeline runs
- **THEN** no AI calls are made and behavior is identical to today

#### Scenario: Invalid backend name
- **GIVEN** a user runs `gaze quality --ai-mapper=invalid ./...`
- **WHEN** the command parses the flag
- **THEN** the command returns an error listing the valid backend names

### Requirement: gaze crap MUST propagate AI mapper to quality pipeline

When `--ai-mapper` is set on `gaze crap`, the AI mapper MUST be propagated through `BuildContractCoverageFunc` to the quality pipeline, so that contract coverage calculations benefit from AI-assisted assertion mappings.

#### Scenario: AI mapper improves contract coverage in CRAP output
- **GIVEN** a function with 4 contractual effects where only 1 maps mechanically (25% coverage)
- **WHEN** `gaze crap --ai-mapper=claude ./...` runs and the AI maps 1 additional assertion
- **THEN** contract coverage increases to 50% (2/4) and GazeCRAP decreases accordingly

### Requirement: AI mapper MUST reuse existing adapter infrastructure

The AI mapper MUST delegate to the existing `internal/aireport` adapter implementations (`ClaudeAdapter`, `GeminiAdapter`, `OllamaAdapter`, `OpenCodeAdapter`) rather than creating new subprocess/HTTP integration code. The adapter's `Format` method MUST be called with the AI mapper prompt as the system prompt and the assertion context as the payload.

#### Scenario: AI mapper uses Claude adapter subprocess
- **GIVEN** `--ai-mapper=claude` is specified and `claude` binary is on PATH
- **WHEN** the AI mapper callback is invoked for an unmapped assertion
- **THEN** the callback calls `ClaudeAdapter.Format` with the mapper prompt and parses the response with `ParseAIMapperResponse`

### Requirement: AI-mapped assertions MUST be distinguishable in output

Assertions mapped by the AI mapper MUST carry confidence 50 in the JSON output. This is lower than all mechanical passes (55-75), ensuring AI mappings are identifiable and do not override mechanical results.

#### Scenario: AI mapping visible in JSON output
- **GIVEN** an assertion mapped by the AI mapper
- **WHEN** the quality JSON output is generated
- **THEN** the assertion's mapping entry has `confidence: 50` and the mapping is included in contract coverage calculations

## MODIFIED Requirements

(none)

## REMOVED Requirements

(none)
