package aireport

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestWriteStepSummary_EmptyPath verifies that an empty path emits a warning
// and returns without writing anything.
func TestWriteStepSummary_EmptyPath(t *testing.T) {
	var stderr strings.Builder
	WriteStepSummary("", "content", &stderr)
	if !strings.Contains(stderr.String(), "warning") {
		t.Errorf("expected warning for empty path, got: %q", stderr.String())
	}
}

// TestWriteStepSummary_RelativePath verifies that a relative path emits a
// warning and returns without writing anything.
func TestWriteStepSummary_RelativePath(t *testing.T) {
	var stderr strings.Builder
	WriteStepSummary("relative/path/summary.md", "content", &stderr)
	if !strings.Contains(stderr.String(), "warning") {
		t.Errorf("expected warning for relative path, got: %q", stderr.String())
	}
}

// TestWriteStepSummary_ExistingFile verifies that content is appended to an
// existing file at a valid absolute path.
func TestWriteStepSummary_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "summary.md")

	existing := "# Existing content\n"
	if err := os.WriteFile(p, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	var stderr strings.Builder
	WriteStepSummary(p, "# Appended\n", &stderr)

	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), existing) {
		t.Errorf("expected existing content preserved, got: %s", data)
	}
	if !strings.Contains(string(data), "# Appended") {
		t.Errorf("expected appended content, got: %s", data)
	}
	if stderr.String() != "" {
		t.Errorf("expected no warning for valid path, got: %q", stderr.String())
	}
}

// TestWriteStepSummary_NonExistentFile verifies that a new file is created
// when the path does not exist.
func TestWriteStepSummary_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "new-summary.md")

	var stderr strings.Builder
	WriteStepSummary(p, "# New content\n", &stderr)

	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("expected file to be created: %v", err)
	}
	if !strings.Contains(string(data), "# New content") {
		t.Errorf("expected written content, got: %s", data)
	}
	if stderr.String() != "" {
		t.Errorf("expected no warning, got: %q", stderr.String())
	}
}

// TestWriteStepSummary_UnwritablePath verifies that an unwritable path emits
// a warning and returns without error.
func TestWriteStepSummary_UnwritablePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod not reliable on Windows")
	}
	dir := t.TempDir()
	// Make the directory unwritable so creating a file inside it fails.
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	p := filepath.Join(dir, "summary.md")
	var stderr strings.Builder
	WriteStepSummary(p, "content", &stderr)

	if !strings.Contains(stderr.String(), "warning") {
		t.Errorf("expected warning for unwritable path, got: %q", stderr.String())
	}
}

// TestWriteStepSummary_SymlinkPath verifies that a symlink target is rejected
// by O_NOFOLLOW, a warning is emitted, and the function returns nil (no error).
func TestWriteStepSummary_SymlinkPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks may require elevated privileges on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "real-file.md")
	if err := os.WriteFile(target, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "symlink.md")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	var stderr strings.Builder
	WriteStepSummary(link, "content", &stderr)

	// O_NOFOLLOW should cause ELOOP; warning emitted, no panic.
	if !strings.Contains(stderr.String(), "warning") {
		t.Errorf("expected warning for symlink path, got: %q", stderr.String())
	}
	// The target file must not have been written to.
	data, _ := os.ReadFile(target)
	if strings.Contains(string(data), "content") {
		t.Errorf("expected symlink target NOT written, but it was")
	}
}
