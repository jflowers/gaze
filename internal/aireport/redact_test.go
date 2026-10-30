package aireport

import (
	"strings"
	"testing"
)

// TestRedactEnvValues exercises the pure redaction helper with a synthetic
// environ, per design D4 of the redact-env-values-from-adapter-stderr
// change: no t.Setenv and no subprocesses here — the end-to-end scrub
// paths through runSubprocess are covered by the TestRunSubprocess_*
// redaction tests in subprocess_test.go. Each case pins one clause of the
// helper's contract: the name-marker and length-threshold sensitivity
// rules (design D2), the garble guard, and the replacement mechanics
// (design D3).
func TestRedactEnvValues(t *testing.T) {
	// configJSON mirrors the reusable-workflow secret shape — a
	// strict-JSON opencode config carrying the apiKey — with an obviously
	// fake value, kept below redactValueLengthThreshold so only the
	// CONFIG_CONTENT name marker drives the redaction.
	const configJSON = `{"apiKey":"sk-fake"}`

	tests := []struct {
		name    string
		s       string
		environ []string
		want    string
	}{
		// Name-marker rule (design D2): a name containing any marker,
		// matched case-insensitively, gets its value redacted regardless
		// of value length. One row per marker pins every entry of
		// sensitiveEnvNameMarkers against accidental removal.
		{
			name: "KEY marker redacts value at every occurrence",
			s:    "Auth failed for sk-ant-fake-123456 (retry sk-ant-fake-123456)",
			environ: []string{
				"OPENAI_API_KEY=sk-ant-fake-123456",
			},
			want: "Auth failed for [REDACTED:OPENAI_API_KEY] (retry [REDACTED:OPENAI_API_KEY])",
		},
		{
			name: "TOKEN marker redacts value",
			s:    "403 for ghp_fakeToken123456",
			environ: []string{
				"GITHUB_TOKEN=ghp_fakeToken123456",
			},
			want: "403 for [REDACTED:GITHUB_TOKEN]",
		},
		{
			name: "SECRET marker redacts value",
			s:    "db connect failed: fake-db-secret-99",
			environ: []string{
				"CLIENT_SECRET=fake-db-secret-99",
			},
			want: "db connect failed: [REDACTED:CLIENT_SECRET]",
		},
		{
			name: "CONFIG_CONTENT marker redacts below-threshold JSON value",
			s:    "opencode rejected config " + configJSON,
			environ: []string{
				"OPENCODE_CONFIG_CONTENT=" + configJSON,
			},
			want: "opencode rejected config [REDACTED:OPENCODE_CONFIG_CONTENT]",
		},
		{
			// Both sides are upper-cased for matching; the marker keeps
			// the variable's original (lowercase) spelling.
			name: "lowercase variable name matches marker case-insensitively",
			s:    "auth failed for short-but-secret",
			environ: []string{
				"gaze_test_api_key=short-but-secret",
			},
			want: "auth failed for [REDACTED:gaze_test_api_key]",
		},

		// Length-threshold rule (design D2): a markerless name still gets
		// its value redacted once the value reaches
		// redactValueLengthThreshold. The boundary is pinned from both
		// sides so an off-by-one cannot slip through.
		{
			name: "length threshold redacts markerless value at threshold",
			s:    "runner " + strings.Repeat("x", redactValueLengthThreshold) + " failed",
			environ: []string{
				"CI_RUNNER_ID=" + strings.Repeat("x", redactValueLengthThreshold),
			},
			want: "runner [REDACTED:CI_RUNNER_ID] failed",
		},
		{
			name: "markerless value below threshold is not redacted",
			s:    "runner " + strings.Repeat("y", redactValueLengthThreshold-1) + " ok",
			environ: []string{
				"CI_RUNNER_ID=" + strings.Repeat("y", redactValueLengthThreshold-1),
			},
			want: "runner " + strings.Repeat("y", redactValueLengthThreshold-1) + " ok",
		},

		// Garble guard (design D2): values shorter than
		// redactMinReplacementLength are never replaced, even under a
		// sensitive name — replacing a 1-3 character value everywhere
		// would destroy the surrounding error text. The boundary is
		// pinned from both sides.
		{
			name: "garble guard skips sub-minimum-length value despite marker name",
			s:    "mode " + strings.Repeat("a", redactMinReplacementLength-1) + " is invalid",
			environ: []string{
				"GAZE_MODE_KEY=" + strings.Repeat("a", redactMinReplacementLength-1),
			},
			want: "mode " + strings.Repeat("a", redactMinReplacementLength-1) + " is invalid",
		},
		{
			name: "minimum-length value under marker name is redacted",
			s:    "pin " + strings.Repeat("b", redactMinReplacementLength) + " rejected",
			environ: []string{
				"PIN_TOKEN=" + strings.Repeat("b", redactMinReplacementLength),
			},
			want: "pin [REDACTED:PIN_TOKEN] rejected",
		},
		{
			// The empty value is the degenerate garble-guard case:
			// strings.ReplaceAll with an empty old string would corrupt
			// the text between every rune.
			name: "empty value under sensitive name is skipped without garbling",
			s:    "plain diagnostic text",
			environ: []string{
				"OPENAI_API_KEY=",
			},
			want: "plain diagnostic text",
		},

		// Replacement mechanics (design D3).
		{
			// If the shorter value were replaced first it would match
			// inside the longer one, leaving "-prefix-abcdef" behind as
			// a partial leak.
			name: "overlapping values replace longest first",
			s:    "leaked secret-token-prefix-abcdef and secret-token too",
			environ: []string{
				"LONG_TOKEN=secret-token-prefix-abcdef",
				"SHORT_KEY=secret-token",
			},
			want: "leaked [REDACTED:LONG_TOKEN] and [REDACTED:SHORT_KEY] too",
		},
		{
			// Identical values are claimed once, by the first variable
			// in environ order.
			name: "duplicate value dedupes with first environ name winning",
			s:    "saw dup-secret-value-123 once",
			environ: []string{
				"PRIMARY_TOKEN=dup-secret-value-123",
				"SECONDARY_KEY=dup-secret-value-123",
			},
			want: "saw [REDACTED:PRIMARY_TOKEN] once",
		},
		{
			// Entries split at the FIRST '=' only; a value truncated at
			// an inner '=' would fail to match and leak. Strict-JSON and
			// base64-padded values routinely contain '='.
			name: "value containing embedded equals is captured whole",
			s:    "rejected tok=en=abc at handshake",
			environ: []string{
				"GAZE_TEST_TOKEN=tok=en=abc",
			},
			want: "rejected [REDACTED:GAZE_TEST_TOKEN] at handshake",
		},
		{
			// Entries with no '=' at all (including the empty string)
			// carry no derivable name/value pair and are skipped.
			name: "malformed environ entries are skipped",
			s:    "real-secret-value-1 leaks",
			environ: []string{
				"",
				"GAZE_BROKEN_ENTRY_NO_EQUALS",
				"GAZE_TEST_KEY=real-secret-value-1",
			},
			want: "[REDACTED:GAZE_TEST_KEY] leaks",
		},

		// Marker format and pass-through behavior.
		{
			// Full-string equality pins the exact format: brackets,
			// colon, and the unmodified variable name.
			name: "whole-value replacement produces exactly the [REDACTED:VAR] marker",
			s:    "sk-fake-987654321",
			environ: []string{
				"MY_API_KEY=sk-fake-987654321",
			},
			want: "[REDACTED:MY_API_KEY]",
		},
		{
			// The scrub replaces value occurrences only: marker-shaped
			// text already in the snippet is left alone, and a sensitive
			// value absent from the snippet changes nothing.
			name: "marker-lookalike text and absent values pass through untouched",
			s:    "child emitted [REDACTED:NOT_A_REAL_VAR] already",
			environ: []string{
				"GAZE_TEST_SECRET=absent-secret-value",
				"HOME=/home/user",
			},
			want: "child emitted [REDACTED:NOT_A_REAL_VAR] already",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactEnvValues(tt.s, tt.environ)
			if got != tt.want {
				t.Errorf("redactEnvValues(%q, environ) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}
