package diff

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitDiffer is a production Differ that shells out to git.
type GitDiffer struct{}

// Ensure GitDiffer satisfies Differ.
var _ Differ = (*GitDiffer)(nil)

// gitDiffTimeout bounds the git subprocess to prevent indefinite hangs
// from stalled NFS, pathological repos, or misbehaving external diff
// drivers configured via .gitattributes.
const gitDiffTimeout = 30 * time.Second

// Diff runs `git diff <ref> --unified=0` or `git diff --staged
// --unified=0` when ref is "staged". Returns the raw diff output.
func (g *GitDiffer) Diff(ref string) (string, error) {
	if strings.HasPrefix(ref, "-") {
		return "", fmt.Errorf("invalid ref %q: refs must not start with '-'", ref)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return "", fmt.Errorf("git binary not found: --gate-on-change requires git")
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitDiffTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", diffArgs(ref)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git diff timed out after %s", gitDiffTimeout)
		}
		return "", fmt.Errorf("git diff failed: %s: %w", stderr.String(), err)
	}

	return stdout.String(), nil
}

// diffArgs returns the git arguments for a diff against the given ref,
// using `--staged` when ref is "staged". `--no-ext-diff` disables
// external diff drivers, `-c core.quotepath=false` disables path
// quoting, and `--no-color` disables ANSI escapes, keeping output in the
// exact format the parser expects regardless of user git config.
func diffArgs(ref string) []string {
	if ref == "staged" {
		return []string{"-c", "core.quotepath=false", "diff", "--staged", "--unified=0", "--no-ext-diff", "--no-color"}
	}
	return []string{"-c", "core.quotepath=false", "diff", ref, "--unified=0", "--no-ext-diff", "--no-color"}
}
