package aireport

import (
	"sort"
	"strings"
)

// sensitiveEnvNameMarkers lists the substrings that mark an environment
// variable name as sensitive: a variable whose name contains any marker has
// its value redacted by redactEnvValues regardless of value length. Markers
// are matched case-insensitively against the upper-cased variable name, so
// lowercase and mixed-case names (e.g. opencode_config_content) match too.
// The slice is deliberately extensible — broadening the match (PASSWORD,
// CREDENTIAL, PRIVATE) is a one-line addition — and MUST be treated as
// read-only at runtime.
var sensitiveEnvNameMarkers = []string{
	"KEY",
	"TOKEN",
	"SECRET",
	"CONFIG_CONTENT",
}

// redactValueLengthThreshold is the value length at or above which
// redactEnvValues scrubs an inherited env-var value regardless of its name.
// It catches oddly-named secrets — e.g. a strict-JSON config blob exported
// under a name carrying no sensitive marker. Over-redaction of long mundane
// values (PATH, GITHUB_SHA, workspace paths) is accepted: it is bounded by
// the stderr snippet cap and the marker names what was removed.
const redactValueLengthThreshold = 32

// redactMinReplacementLength is the minimum value length redactEnvValues
// requires before it will replace a sensitive value (the garble guard).
// Values below this length — including empty values, which
// strings.ReplaceAll cannot handle — are never replaced, even when the
// variable name matches a sensitive marker: replacing a 1-3 character value
// everywhere would destroy the surrounding error text. Sub-minimum-length
// secrets are a documented limitation.
const redactMinReplacementLength = 4

// isSensitiveEnvName reports whether name contains any of the
// sensitiveEnvNameMarkers, matched case-insensitively. Both sides are
// upper-cased before the substring check so the marker slice stays correct
// even if extended with non-uppercase entries.
func isSensitiveEnvName(name string) bool {
	upper := strings.ToUpper(name)
	for _, marker := range sensitiveEnvNameMarkers {
		if strings.Contains(upper, strings.ToUpper(marker)) {
			return true
		}
	}
	return false
}

// redactEnvValues returns s with every sensitive value inherited from
// environ replaced by a "[REDACTED:<VAR>]" marker naming the environment
// variable the value came from. A value is sensitive when its variable name
// contains a sensitiveEnvNameMarker (case-insensitive) or when its length
// reaches redactValueLengthThreshold regardless of name.
//
// Replacement mechanics:
//   - Values shorter than redactMinReplacementLength are never replaced
//     (garble guard; also covers empty values).
//   - Identical values shared by several variables are replaced once — the
//     first variable name in environ order wins, keeping output
//     deterministic.
//   - Replacements run longest value first, so when one value is a
//     substring of another the longer one is replaced before the shorter
//     can match inside it, avoiding partial or nested artifacts.
//
// Each environ entry is split at the FIRST '=' only, so values containing
// embedded '=' characters (strict JSON, base64 padding) are captured whole;
// entries containing no '=' at all are malformed and skipped.
//
// Callers pass os.Environ(), read after cmd.Wait() returns. That snapshot
// equals the environment the child process inherited, because gaze blocks
// inside cmd.Wait() for the child's entire lifetime and never mutates its
// own environment during adapter execution — a documented assumption
// (design D4 of the redact-env-values-from-adapter-stderr change).
func redactEnvValues(s string, environ []string) string {
	// A redaction pairs a sensitive env value with the name of the first
	// variable (in environ order) that carries it.
	type redaction struct {
		name  string
		value string
	}

	var redactions []redaction
	seen := make(map[string]bool)
	for _, entry := range environ {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			// Malformed entry (no '='); nothing to derive from it.
			continue
		}
		name, value := parts[0], parts[1]
		if len(value) < redactMinReplacementLength {
			// Garble guard: below-minimum and empty values are never
			// replaced.
			continue
		}
		if !isSensitiveEnvName(name) && len(value) < redactValueLengthThreshold {
			// Not sensitive: name carries no marker and the value is
			// below the length threshold.
			continue
		}
		if seen[value] {
			// Dedupe: the first name in environ order already claimed
			// this value.
			continue
		}
		seen[value] = true
		redactions = append(redactions, redaction{name: name, value: value})
	}

	// Longest value first so overlapping values cannot leave partial or
	// nested artifacts; the stable sort keeps environ order for equal
	// lengths, making replacement order fully deterministic.
	sort.SliceStable(redactions, func(i, j int) bool {
		return len(redactions[i].value) > len(redactions[j].value)
	})

	for _, r := range redactions {
		s = strings.ReplaceAll(s, r.value, "[REDACTED:"+r.name+"]")
	}
	return s
}
