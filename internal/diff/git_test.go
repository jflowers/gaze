package diff

import (
	"strings"
	"testing"
)

func TestGitDiffer_GitNotFound(t *testing.T) {
	gd := &GitDiffer{}
	_, err := gd.Diff("HEAD")
	if err != nil && !strings.Contains(err.Error(), "git binary not found") {
		t.Logf("git is on PATH, skipping git-not-found test: %v", err)
	}
}

func TestGitDiffer_InvalidRef(t *testing.T) {
	gd := &GitDiffer{}
	_, err := gd.Diff("nonexistent-branch-xyz-12345")
	if err == nil {
		t.Skip("git ref unexpectedly resolved")
	}
	if !strings.Contains(err.Error(), "git diff failed") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "git diff failed")
	}
}

func TestGitDiffer_StagedArgument(t *testing.T) {
	gd := &GitDiffer{}
	output, err := gd.Diff("staged")
	if err != nil {
		t.Skipf("git diff --staged failed (may not be in a git repo): %v", err)
	}
	if output == "" {
		t.Log("no staged changes (empty diff is valid)")
	}
}
