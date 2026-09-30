package diff

import (
	"strings"
	"testing"
)

func TestGitDiffer_GitNotFound(t *testing.T) {
	// Deterministically remove git from PATH so exec.LookPath fails
	// regardless of the host environment.
	t.Setenv("PATH", t.TempDir())
	_, err := (&GitDiffer{}).Diff("HEAD")
	if err == nil || !strings.Contains(err.Error(), "git binary not found: --gate-on-change requires git") {
		t.Fatalf("got err %v, want 'git binary not found: --gate-on-change requires git'", err)
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

func TestDiffArgs(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want []string
	}{
		{
			name: "staged",
			ref:  "staged",
			want: []string{"-c", "core.quotepath=false", "diff", "--staged", "--unified=0", "--no-ext-diff", "--no-color"},
		},
		{
			name: "ref",
			ref:  "HEAD",
			want: []string{"-c", "core.quotepath=false", "diff", "HEAD", "--unified=0", "--no-ext-diff", "--no-color"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffArgs(tt.ref)
			if len(got) != len(tt.want) {
				t.Fatalf("diffArgs(%q) = %v, want %v", tt.ref, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("diffArgs(%q)[%d] = %q, want %q", tt.ref, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGitDiffer_RejectsLeadingDashRef(t *testing.T) {
	gd := &GitDiffer{}
	_, err := gd.Diff("--output=/tmp/evil")
	if err == nil {
		t.Fatal("expected error for leading-dash ref")
	}
	if !strings.Contains(err.Error(), "invalid ref") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "invalid ref")
	}
}
