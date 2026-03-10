//go:build windows

package aireport

import (
	"fmt"
	"io"
)

// WriteStepSummary is a no-op on Windows: O_NOFOLLOW is not available on
// Windows, so Step Summary writing is not supported. A warning is emitted
// to stderr and the function returns without error (FR-008).
func WriteStepSummary(path, content string, stderr io.Writer) {
	if path == "" {
		return
	}
	_, _ = fmt.Fprintln(stderr, "warning: GITHUB_STEP_SUMMARY write is not supported on Windows; skipping Step Summary write")
}
