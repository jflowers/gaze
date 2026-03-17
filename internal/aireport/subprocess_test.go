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
