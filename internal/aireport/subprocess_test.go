package aireport

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSubprocess_Success(t *testing.T) {
	ctx := context.Background()
	// Use "echo" which is available on all platforms.
	outBytes, _, err := runSubprocess(ctx, "echo", []string{"hello", "world"}, "", nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	result := strings.TrimSpace(string(outBytes))
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestRunSubprocess_BinaryNotFound(t *testing.T) {
	ctx := context.Background()
	_, _, err := runSubprocess(ctx, "gaze-nonexistent-binary-xyz", nil, "", nil)
	if err == nil {
		t.Fatal("expected error for missing binary, got nil")
	}
	if !strings.Contains(err.Error(), "gaze-nonexistent-binary-xyz") {
		t.Errorf("expected error to contain binary name, got: %v", err)
	}
	if !strings.Contains(err.Error(), "not found on PATH") {
		t.Errorf("expected error to mention PATH, got: %v", err)
	}
}

func TestRunSubprocess_NonZeroExit(t *testing.T) {
	ctx := context.Background()
	// "false" exits with status 1 on Unix. On Windows this test would
	// need a different binary, but gaze targets darwin/linux only.
	_, _, err := runSubprocess(ctx, "false", nil, "", nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
	if !strings.Contains(err.Error(), "false") {
		t.Errorf("expected error to contain binary name 'false', got: %v", err)
	}
	if !strings.Contains(err.Error(), "exited with error") {
		t.Errorf("expected error to mention 'exited with error', got: %v", err)
	}
}

func TestRunSubprocess_CmdDir(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a marker file in tmpDir.
	marker := filepath.Join(tmpDir, "marker.txt")
	if err := os.WriteFile(marker, []byte("found"), 0644); err != nil {
		t.Fatalf("creating marker file: %v", err)
	}

	// Use "cat marker.txt" with cmdDir set — should find the file.
	outBytes, _, err := runSubprocess(ctx, "cat", []string{"marker.txt"}, tmpDir, nil)
	if err != nil {
		t.Fatalf("expected nil error with cmdDir, got: %v", err)
	}
	if string(outBytes) != "found" {
		t.Errorf("expected 'found', got %q", string(outBytes))
	}
}

func TestRunSubprocess_StderrReturnedOnSuccess(t *testing.T) {
	ctx := context.Background()
	// "sh -c" writes to both stdout and stderr.
	outBytes, stderrBytes, err := runSubprocess(ctx, "sh", []string{"-c", "echo stdout-msg; echo stderr-msg >&2"}, "", nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if got := strings.TrimSpace(string(outBytes)); got != "stdout-msg" {
		t.Errorf("expected stdout 'stdout-msg', got %q", got)
	}
	if got := strings.TrimSpace(string(stderrBytes)); got != "stderr-msg" {
		t.Errorf("expected stderr 'stderr-msg', got %q", got)
	}
}

func TestRunSubprocess_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// "sleep 60" should be killed by context cancellation.
	_, _, err := runSubprocess(ctx, "sleep", []string{"60"}, "", nil)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestRunSubprocess_NonZeroExitRedactsSensitiveEnvValue(t *testing.T) {
	ctx := context.Background()
	// Plant an obviously fake secret under a name carrying the KEY marker.
	// t.Setenv puts it in the real process environment, so the sh child
	// inherits it (runSubprocess never sets cmd.Env) and echoes it to
	// stderr while the scrub inside runSubprocess reads the same value
	// from os.Environ() — the full end-to-end path.
	const secret = "test-secret-api-key-xyz123"
	t.Setenv("GAZE_TEST_API_KEY", secret)

	_, _, err := runSubprocess(ctx, "sh", []string{"-c",
		`echo "gaze-test-stderr-context: auth failed for ${GAZE_TEST_API_KEY}" >&2; exit 1`,
	}, "", nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
	if !strings.Contains(err.Error(), "[REDACTED:GAZE_TEST_API_KEY]") {
		t.Errorf("expected redaction marker naming the env var, got: %v", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("expected sensitive env value to be scrubbed from error, got: %v", err)
	}
	// The scrub must be surgical, not a wholesale nuke: surrounding
	// diagnostic text survives (scrubbing preserves diagnosability).
	if !strings.Contains(err.Error(), "gaze-test-stderr-context: auth failed for") {
		t.Errorf("expected surrounding stderr text to survive redaction, got: %v", err)
	}
}

func TestRunSubprocess_EmptyStdoutSurfacesRedactedStderr(t *testing.T) {
	ctx := context.Background()
	// The CONFIG_CONTENT name marker fires even though the fake value (26
	// bytes) is below the length threshold, mirroring the real
	// OPENCODE_CONFIG_CONTENT secret shape from the reusable workflow.
	const secret = "test-config-content-abc123"
	t.Setenv("GAZE_TEST_CONFIG_CONTENT", secret)

	// Exit 0, empty stdout, secret on stderr — the adapter empty-output
	// precondition (FR-009/FR-016).
	outBytes, stderrBytes, err := runSubprocess(ctx, "sh", []string{"-c",
		`echo "gaze-test-stderr-context: rejected config ${GAZE_TEST_CONFIG_CONTENT}" >&2`,
	}, "", nil)
	if err != nil {
		t.Fatalf("expected nil error for exit-0 run, got: %v", err)
	}
	if len(outBytes) != 0 {
		t.Errorf("expected empty stdout, got %q", string(outBytes))
	}
	// Mirror the adapters: the success-path stderr bytes are surfaced
	// through formatStderrSuffix when stdout comes back empty.
	suffix := formatStderrSuffix(stderrBytes)
	if !strings.Contains(suffix, "[REDACTED:GAZE_TEST_CONFIG_CONTENT]") {
		t.Errorf("expected redaction marker naming the env var, got: %q", suffix)
	}
	if strings.Contains(suffix, secret) {
		t.Errorf("expected sensitive env value to be scrubbed from stderr suffix, got: %q", suffix)
	}
	if !strings.Contains(suffix, "gaze-test-stderr-context: rejected config") {
		t.Errorf("expected surrounding stderr text to survive redaction, got: %q", suffix)
	}
}

func TestRunSubprocess_RedactsSecretStraddlingTruncationBoundary(t *testing.T) {
	ctx := context.Background()
	// Center the fake secret on the truncation boundary: it starts before
	// byte maxAdapterStderrBytes and extends past it, so a truncation that
	// ran before the scrub would leak a prefix of the secret into the
	// surfaced error (design D3: the scrub must precede truncation).
	const secret = "test-secret-straddling-the-truncation-boundary-0123456789abcdef"
	t.Setenv("GAZE_TEST_STRADDLE_SECRET", secret)

	// Sanity-check the straddle arithmetic against the live constants.
	headLen := maxAdapterStderrBytes - len(secret)/2
	if headLen < 0 || headLen >= maxAdapterStderrBytes || headLen+len(secret) <= maxAdapterStderrBytes {
		t.Fatalf("test setup error: secret must start before byte %d and extend past it (starts at %d, length %d)",
			maxAdapterStderrBytes, headLen, len(secret))
	}
	head := strings.Repeat("x", headLen)
	tail := strings.Repeat("y", 64)
	// printf (not echo) emits exactly head+secret+tail bytes with no
	// trailing newline, keeping the boundary arithmetic exact. Braces are
	// required around the variable: it is immediately followed by the
	// tail padding, so a bare $VAR would expand the wrong name.
	script := `printf '%s' "` + head + `${GAZE_TEST_STRADDLE_SECRET}` + tail + `" >&2; exit 1`

	_, _, err := runSubprocess(ctx, "sh", []string{"-c", script}, "", nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
	errMsg := err.Error()
	// Setup self-check: the scrubbed stderr must still exceed the byte cap
	// so the truncation path — and thus the straddle — is really exercised.
	if !strings.Contains(errMsg, "... (truncated)") {
		t.Fatalf("test setup error: expected stderr truncation to engage, got: %v", errMsg)
	}
	// The marker replacing the secret may itself be cut by the byte cap, so
	// only the secret's absence is asserted. The prefix scan starts at
	// redactMinReplacementLength: shorter prefixes are outside the scrub's
	// guarantee (the garble guard) and collide with the "... (truncated)"
	// suffix text (e.g. "te" appears inside "truncated").
	for k := redactMinReplacementLength; k <= len(secret); k++ {
		if strings.Contains(errMsg, secret[:k]) {
			t.Fatalf("surfaced error leaks a %d-byte prefix of the sensitive value: %v", k, errMsg)
		}
	}
}

func TestFormatStderrSuffix_NonEmpty(t *testing.T) {
	got := formatStderrSuffix([]byte("Error: Invalid API key."))
	want := "\nstderr: Error: Invalid API key."
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestFormatStderrSuffix_Empty(t *testing.T) {
	got := formatStderrSuffix([]byte(""))
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFormatStderrSuffix_WhitespaceOnly(t *testing.T) {
	got := formatStderrSuffix([]byte("  \n\t  "))
	if got != "" {
		t.Errorf("expected empty string for whitespace-only stderr, got %q", got)
	}
}

func TestFormatStderrSuffix_Truncation(t *testing.T) {
	// Create a string longer than maxAdapterStderrBytes (512).
	long := strings.Repeat("x", 600)
	got := formatStderrSuffix([]byte(long))
	if !strings.Contains(got, "... (truncated)") {
		t.Errorf("expected truncation marker, got %q", got)
	}
	// The prefix should be "\nstderr: " + 512 chars + "... (truncated)".
	expectedPrefix := "\nstderr: " + strings.Repeat("x", 512)
	if !strings.HasPrefix(got, expectedPrefix) {
		t.Errorf("expected prefix with 512 chars, got length %d", len(got))
	}
}
