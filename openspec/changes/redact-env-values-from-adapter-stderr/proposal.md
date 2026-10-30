# Redact Env Values from Adapter Stderr

## Why

The `internal/aireport` subprocess adapters surface AI-CLI stderr in
error messages to make adapter failures diagnosable: `runSubprocess`
embeds up to `maxAdapterStderrBytes` (512) of child stderr in its
non-zero-exit error, and `formatStderrSuffix` appends a bounded snippet
to the claude/gemini/opencode empty-output errors (FR-009/FR-016). The
cap bounds volume but does not scrub content.

The reusable gaze AI report workflow (PR #300) delivers the consumer's
entire opencode configuration — provider, model, permissions, and the
literal API key — as a single strict-JSON secret, exported verbatim as
`OPENCODE_CONFIG_CONTENT` into the report step's environment. gaze
spawns the AI CLI inheriting that environment (`runSubprocess` never
sets `cmd.Env`), producing this leak path if a pinned CLI release ever
echoed config fragments to stderr (for example, in a JSON parse
error):

child stderr → `runSubprocess` / `formatStderrSuffix` error string →
`runTextPath` "AI formatting failed" error → CLI stderr → GitHub
Actions job log.

GitHub's automatic secret masking registers only the full secret value
— the whole JSON document — so a bare `apiKey` substring echoed alone
would appear unmasked in CI logs.

Today there is no demonstrated leak: gaze itself never reads, logs, or
writes the config (pure env passthrough), and opencode-ai@1.19.0 shows
no config-echo behavior on stderr. This change is the defense-in-depth
follow-up from the PR #300 divisor-adversary review (issue #301): gaze
is the process that passes the environment to the child, so it is also
the right place to scrub inherited secret values from any child stderr
it surfaces.

## What Changes

### Stderr redaction at the runSubprocess choke point

`runSubprocess` scrubs child-process stderr against sensitive
**inherited env-var values** before that stderr is embedded in an error
or returned to callers. A value is scrubbed when its env-var name
contains `KEY`, `TOKEN`, `SECRET`, or `CONFIG_CONTENT`
(case-insensitive), or when the value is at or above a length threshold
regardless of name; scrubbed values are replaced with a
`[REDACTED:<VAR>]` marker. JSON-pattern scrubbing was evaluated in the
issue and rejected as brittle. Redaction runs on the full stderr before
the 512-byte truncation, so a secret straddling the truncation boundary
cannot leak a prefix. All three subprocess adapters (claude, gemini,
opencode) and both surfacing paths (non-zero-exit error and
empty-output suffix) inherit the scrub through the single choke point.
The environment remains pure passthrough — gaze still never reads the
config beyond the redaction set it derives from its own inherited
environment.

### CI consumer documentation

`docs/guides/ci-integration.md` gains a defense-in-depth subsection for
reusable-workflow consumers: registering the bare API key as a second
secret referenced in an unused env var of the caller's job (so GitHub
masks verbatim occurrences anywhere in logs), and preferring a
single-line strict-JSON secret value (GitHub's multiline secret masking
is unreliable). Documentary only — no workflow, contract, or CI
configuration changes.

## Capabilities

### New Capabilities

- `adapter-stderr-redaction`: subprocess-adapter stderr surfaced in
  errors is scrubbed against sensitive inherited env-var values;
  consumer-side masking mitigations are documented in the CI
  integration guide

### Modified Capabilities

- None

### Removed Capabilities

- None

## Impact

- `internal/aireport/adapter.go` — modified: redaction helper and
  `runSubprocess` wiring
- `internal/aireport/` test files — new table-driven unit tests for the
  pure helper and subprocess-level integration tests (`t.Setenv` plus
  the existing `sh`-based patterns)
- `docs/guides/ci-integration.md` — modified: defense-in-depth
  subsection in the Reusable Workflow section
- Observable error-message change: surfaced stderr snippets may contain
  `[REDACTED:<VAR>]` markers where inherited env values were scrubbed;
  the 512-byte snippet cap and all other error text are unchanged
- No API changes: all new symbols are unexported; no CLI, workflow,
  protocol, or CI configuration changes

Website Documentation Gate: assessed N/A — internal error-output
hardening plus in-repo consumer guidance; no new command, flag,
installation step, or user-facing workflow change. Recorded as an open
question in design.md so a human can override at PR review.

## Constitution Alignment

Assessed against the Gaze project constitution
(`.specify/memory/constitution.md`, v1.3.0).

### I. Accuracy

**Assessment**: N/A

No side-effect detection, classification, or metric behavior changes.
The redaction operates only on adapter-failure error text.

### II. Minimal Assumptions

**Assessment**: PASS

The redaction set is derived solely from gaze's own inherited
environment — the exact environment gaze passes to the child — so it
assumes nothing about the host project, the AI CLI, or the stderr
content. The name markers and length threshold are explicit,
documented, tunable constants; sub-minimum-length and
oddly-named-sub-threshold values are a documented limitation rather
than a silent assumption.

### III. Actionable Output

**Assessment**: PASS

Errors stay actionable: the `[REDACTED:<VAR>]` marker names the env
var whose value was scrubbed, so a consumer can tell exactly what was
removed and why, while secret material stays out of CI logs. The
512-byte snippet cap and surrounding diagnostics are unchanged.

### IV. Testability

**Assessment**: PASS

The scrub is a pure function taking the snippet and an explicit
environ slice, unit-testable with a synthetic environ (table-driven).
It is wired through a single choke point covered by subprocess-level
integration tests using `t.Setenv` and the existing `sh`-based test
patterns — no external services required.
