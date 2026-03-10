package aireport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteStepSummary appends content to the GitHub Actions Step Summary file
// at path. It validates that the path is absolute and refers to a file
// (via os.Lstat), then opens the file for append-write.
//
// On any validation or write error, WriteStepSummary emits a warning to
// stderr and returns nil — Step Summary write failure must not abort the
// command (FR-008).
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

	info, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		_, _ = fmt.Fprintf(stderr,
			"warning: could not stat GITHUB_STEP_SUMMARY %q: %v; skipping Step Summary write\n", path, err)
		return
	}
	if info != nil && !info.Mode().IsRegular() {
		_, _ = fmt.Fprintf(stderr,
			"warning: GITHUB_STEP_SUMMARY %q is not a regular file; skipping Step Summary write\n", path)
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		_, _ = fmt.Fprintf(stderr,
			"warning: could not open GITHUB_STEP_SUMMARY %q for writing: %v; skipping Step Summary write\n",
			path, err)
		return
	}
	defer f.Close()

	if _, err := fmt.Fprint(f, content); err != nil {
		_, _ = fmt.Fprintf(stderr,
			"warning: failed to write to GITHUB_STEP_SUMMARY %q: %v\n", path, err)
	}
}
