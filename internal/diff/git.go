package diff

import (
	"bytes"
	"fmt"
	"os/exec"
)

// GitDiffer is a production Differ that shells out to git.
type GitDiffer struct{}

// Ensure GitDiffer satisfies Differ.
var _ Differ = (*GitDiffer)(nil)

// Diff runs `git diff <ref> --unified=0` or `git diff --staged
// --unified=0` when ref is "staged". Returns the raw diff output.
func (g *GitDiffer) Diff(ref string) (string, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", fmt.Errorf("git binary not found: --gate-on-change requires git")
	}

	var args []string
	if ref == "staged" {
		args = []string{"diff", "--staged", "--unified=0"}
	} else {
		args = []string{"diff", ref, "--unified=0"}
	}

	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff failed: %s: %w", stderr.String(), err)
	}

	return stdout.String(), nil
}
