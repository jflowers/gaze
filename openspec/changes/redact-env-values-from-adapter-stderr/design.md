# Design: Redact Env Values from Adapter Stderr

## Context

`runSubprocess` (`internal/aireport/adapter.go`) is the shared executor
for the three subprocess AI adapters (claude, gemini, opencode). It
never sets `cmd.Env`, so the AI CLI inherits gaze's full environment.
Stderr surfaces through two paths:

1. Non-zero exit: the error embeds up to `maxAdapterStderrBytes` (512)
   of raw child stderr.
2. Exit 0 with empty output: the adapter appends a
   `formatStderrSuffix` snippet (also capped at 512) to its
   empty-output error (FR-009/FR-016).

Both paths cap volume but do not scrub content. With the
config-as-single-secret pattern from the reusable gaze AI report
workflow (PR #300) — the entire strict-JSON opencode config including
the literal `apiKey`, exported verbatim as `OPENCODE_CONFIG_CONTENT` —
a pinned CLI release that echoed config fragments to stderr would put
bare key material into CI job logs. GitHub's automatic masking
registers only the full JSON document as the secret, so a bare
`apiKey` substring echoed alone appears unmasked. There is no
demonstrated leak today (gaze never reads/logs/writes the config;
opencode-ai@1.19.0 shows no echo behavior) — this is defense-in-depth
(issue #301, follow-up to the PR #300 divisor-adversary review).

## Goals / Non-Goals

### Goals

- Scrub sensitive inherited env-var values from every byte of child
  stderr gaze surfaces, on both surfacing paths
- Keep surfaced snippets diagnosable (the marker names the scrubbed
  env var)
- Close the truncation-boundary straddle case (scrub before truncation)
- Document consumer-side masking mitigations in the CI integration
  guide (second bare-key secret, single-line strict-JSON secret shape)

### Non-Goals

- Changing what gaze does with the environment — still pure
  passthrough; `runSubprocess` continues to never set `cmd.Env`
- JSON-pattern or content-shape scrubbing — evaluated in issue #301
  and rejected as brittle
- Redacting HTTP-path (ollama) adapter errors — no subprocess stderr
  involved
- Changing the reusable workflow contract or any CI configuration
- Runtime verification of GitHub's secret-masking behavior — platform
  behavior outside gaze's control; documented as guidance instead

## Decisions

### D1: Redact at the runSubprocess choke point

The scrub is applied once inside `runSubprocess`, to the stderr buffer
immediately after `cmd.Wait()` returns (the buffer is complete at that
point). Both the non-zero-exit error embed and the success-path
returned bytes use the scrubbed string, so `formatStderrSuffix` and all
three adapters need no changes, and any future consumer of
success-path stderr inherits redaction automatically.

Alternative rejected: scrubbing inside `formatStderrSuffix` only — it
would leave the non-zero-exit error path unredacted and require a
second choke point.

### D2: Hybrid sensitive-value classification — name markers plus a length threshold

A value is scrubbed when either rule fires:

- **Name markers**: the env-var name contains `KEY`, `TOKEN`,
  `SECRET`, or `CONFIG_CONTENT`, matched case-insensitively (the four
  markers named in issue #301; kept in an extensible slice constant).
  This catches short secrets in well-named vars — a 16-character API
  key in `OPENAI_API_KEY` would evade any length-only rule.
- **Length threshold**: any inherited value at or above the threshold
  (32) is scrubbed regardless of name. This catches oddly-named
  secrets — a config blob in a var not carrying a marker. Real secrets
  are typically far longer than ordinary CI env values; typical
  over-redactions at 32 (`PATH`, `GITHUB_SHA`, workspace paths) are
  harmless in a 512-byte error snippet.

The two rules cover each other's blind spots, which is why the hybrid
was chosen over either rule alone (the issue offered them as
alternatives; the hybrid is strictly stronger defense-in-depth).

**Garble guard**: values shorter than the minimum replacement length
(4 bytes) are never replaced, even when the name matches — replacing a
1-3 character value everywhere would destroy the error text.
Sub-minimum-length secrets are a documented limitation. All three
values (marker list, threshold, minimum length) are tunable constants
in one place.

### D3: Replacement mechanics — marker, dedupe, longest-first, before truncation

- **Marker**: `[REDACTED:<VAR_NAME>]`. Env-var names are not secret,
  and naming the var keeps the error actionable (Observable Quality:
  the operator can see what was removed). `***` was rejected as less
  diagnosable.
- **Dedupe**: identical values shared by multiple vars are replaced
  once; the first var name in `os.Environ()` order wins, keeping
  output deterministic.
- **Longest-first**: replacements run in descending value length, so
  overlapping values (one env value a substring of another) replace
  the longer one first, avoiding partial/nested artifacts.
- **Empty values skipped**: `strings.ReplaceAll` with an empty old
  string would corrupt the text.
- **Before truncation**: the scrub runs on the full stderr buffer, and
  the existing 512-byte truncation applies afterward. Truncating first
  could leak a prefix of a secret straddling the 512-byte boundary;
  scrubbing first makes that structurally impossible. A marker landing
  across the boundary may itself be cut by the byte cap — cosmetic
  only; no secret material survives.
- **Cost**: the scrub is O(vars × len(stderr)) over a buffer already
  fully in memory, on adapter-failure paths (and success paths with
  non-empty stderr) only — negligible in practice. A windowed scrub
  (only the first 512+max-value bytes) was considered and rejected:
  success-path stderr is returned in full, and simplicity favors the
  whole-buffer scrub.

### D4: Pure helper with environ injection for testability

The core is a pure function
`redactEnvValues(s string, environ []string) string` — the snippet and
an explicit environ slice in, redacted string out — unit-testable with
a synthetic environ (Testability: no external services, no real env
dependence). A thin call in `runSubprocess` passes `os.Environ()`.

The env snapshot is read after `cmd.Wait()`. This equals the
environment the child inherited because gaze blocks inside `cmd.Wait()`
for the child's lifetime and never mutates its own environment during
adapter execution — documented assumption, re-stated in the helper's
GoDoc.

### D5: Documentary mitigations only for the org-level items

Issue #301 items 2 and 3 (register the bare API key as a second secret
so GitHub masks verbatim occurrences; prefer single-line strict JSON
because multiline masking is unreliable) land as documentation in
`docs/guides/ci-integration.md`, not as workflow changes. A consumer
can reference the bare-key secret in an unused env var of their own
calling job without any change to the reusable workflow's contract —
adding a new secret input to the reusable workflow would expand a
published contract for a mitigation consumers can self-apply. The guide
keeps the pretty-printed config example for readability, with an
explicit note that the actual secret value should be single-line.

## Risks / Trade-offs

- **Over-redaction of long non-secret values** (`PATH`, `GITHUB_SHA`,
  workspace paths appear in stderr rarely; when they do, their removal
  costs little diagnostics): accepted — bounded by the 512-byte
  snippet, and the marker names what was removed.
- **Under-redaction**: sub-minimum-length (garble guard),
  sub-threshold oddly-named secrets, and transformed secrets (a child
  that base64-encodes or rewraps the value) are not caught — residual
  risk, mitigated on the platform side by the documented consumer
  guidance (second secret registration masks verbatim occurrences
  anywhere).
- **Error-path cost**: whole-buffer scrub on failure paths only;
  negligible in practice (see D3).
- **Env snapshot timing**: reading `os.Environ()` after `Wait()` is
  equivalent to the inherited environment because gaze cannot mutate
  its env while blocked in `Wait()` (see D4).

## Open Questions

- **Website Documentation Gate**: assessed N/A in the proposal
  (internal error-output hardening plus in-repo consumer guidance; no
  new command, flag, installation step, or workflow change). A human
  may disagree and want a tracking issue in `unbound-force/website` —
  flag at PR review.
- **Extending the name-marker list** (`PASSWORD`, `CREDENTIAL`,
  `PRIVATE`): deferred; the marker slice constant makes this a one-line
  future change if the org wants broader matching.

## Constitution Alignment

Reference: assessed in proposal.md against the Gaze constitution
v1.3.0 — Accuracy N/A (no analysis behavior changes); Minimal
Assumptions PASS (the redaction set derives solely from gaze's own
inherited environment, with explicit tunable constants and a documented
limitation); Actionable Output PASS (the `[REDACTED:<VAR>]` marker
keeps errors diagnosable — an Observable Quality concern); Testability
PASS (pure helper with injected environ plus subprocess-level
integration tests).
