//go:build !windows

package aireport

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// WriteStepSummary appends content to the GitHub Actions Step Summary file
// at path. It validates that the path is absolute, then opens the file
// using O_NOFOLLOW to atomically refuse symlink following at the OS level.
//
// If the open fails because the path is a symlink (ELOOP), a warning is
// emitted to stderr and the function returns without error — Step Summary
// write failure must not abort the command (FR-008).
//
// On any other validation or write error, WriteStepSummary emits a warning
// to stderr and returns, allowing the command to continue.
func WriteStepSummary(path, content string, stderr io.Writer) {
	if path == "" {
		_, _ = fmt.Fprintln(stderr, "warning: GITHUB_STEP_SUMMARY is empty; skipping Step Summary write")
		return
	}
	if !filepath.IsAbs(path) {
		_, _ = fmt.Fprintf(stderr,
			"warning: GITHUB_STEP_SUMMARY %q is not an absolute path; skipping Step Summary write\n", path)
		return
	}

	// Open with O_NOFOLLOW: if the path is a symlink, the kernel returns
	// ELOOP (on Linux/macOS) atomically — no TOCTOU race possible.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE|syscall.O_NOFOLLOW, 0644)
	if err != nil {
		if isSymlinkError(err) {
			_, _ = fmt.Fprintf(stderr,
				"warning: GITHUB_STEP_SUMMARY %q is a symlink; skipping Step Summary write (O_NOFOLLOW)\n", path)
			return
		}
		_, _ = fmt.Fprintf(stderr,
			"warning: could not open GITHUB_STEP_SUMMARY %q for writing: %v; skipping Step Summary write\n",
			path, err)
		return
	}
	defer func() { _ = f.Close() }()

	if _, err := fmt.Fprint(f, content); err != nil {
		_, _ = fmt.Fprintf(stderr,
			"warning: failed to write to GITHUB_STEP_SUMMARY %q: %v\n", path, err)
	}
}

// isSymlinkError reports whether err is the result of O_NOFOLLOW rejecting
// a symlink (ELOOP on POSIX systems).
func isSymlinkError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.ELOOP
	}
	return false
}
