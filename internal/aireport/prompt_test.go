package aireport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadPrompt_LocalFileWithFrontmatter verifies that a local
// .opencode/agents/gaze-reporter.md with YAML frontmatter is loaded and
// stripped before returning.
func TestLoadPrompt_LocalFileWithFrontmatter(t *testing.T) {
	workdir := t.TempDir()
	agentsDir := filepath.Join(workdir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := "---\ndescription: test\n---\n\n# Real Content\n\nSome instructions."
	if err := os.WriteFile(filepath.Join(agentsDir, "gaze-reporter.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadPrompt(workdir)
	if err != nil {
		t.Fatalf("LoadPrompt: %v", err)
	}

	if strings.Contains(got, "description: test") {
		t.Errorf("expected frontmatter stripped, got: %q", got)
	}
	if !strings.Contains(got, "# Real Content") {
		t.Errorf("expected body content preserved, got: %q", got)
	}
	if strings.HasPrefix(got, "---") {
		t.Errorf("expected no leading frontmatter delimiter, got: %q", got)
	}
}

// TestLoadPrompt_LocalFileWithoutFrontmatter verifies that a local file
// without frontmatter is returned as-is.
func TestLoadPrompt_LocalFileWithoutFrontmatter(t *testing.T) {
	workdir := t.TempDir()
	agentsDir := filepath.Join(workdir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := "# Just Instructions\n\nNo frontmatter here."
	if err := os.WriteFile(filepath.Join(agentsDir, "gaze-reporter.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadPrompt(workdir)
	if err != nil {
		t.Fatalf("LoadPrompt: %v", err)
	}
	if got != content {
		t.Errorf("expected content unchanged, got: %q", got)
	}
}

// TestLoadPrompt_LocalFileAbsent_FallsBackToEmbedded verifies that when the
// local file is absent, LoadPrompt returns the embedded default (non-empty).
func TestLoadPrompt_LocalFileAbsent_FallsBackToEmbedded(t *testing.T) {
	workdir := t.TempDir() // no .opencode/agents/gaze-reporter.md

	got, err := LoadPrompt(workdir)
	if err != nil {
		t.Fatalf("LoadPrompt: %v", err)
	}
	if strings.TrimSpace(got) == "" {
		t.Error("expected non-empty embedded default prompt")
	}
}

// TestLoadPrompt_EmbeddedDefaultHasNoFrontmatter verifies that the embedded
// default prompt (after stripping) does not begin with "---".
func TestLoadPrompt_EmbeddedDefaultHasNoFrontmatter(t *testing.T) {
	if strings.HasPrefix(strings.TrimSpace(defaultPromptRaw), "---") {
		// defaultPromptRaw embeds the raw file including frontmatter;
		// LoadPrompt strips it. Verify LoadPrompt strips correctly.
		workdir := t.TempDir()
		got, err := LoadPrompt(workdir)
		if err != nil {
			t.Fatalf("LoadPrompt: %v", err)
		}
		if strings.HasPrefix(got, "---") {
			t.Errorf("embedded default prompt begins with frontmatter after LoadPrompt: %q", got[:50])
		}
	}
}

// TestStripFrontmatter_RemovesBlock verifies the stripping logic directly.
func TestStripFrontmatter_RemovesBlock(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantIn  string
		wantOut string
	}{
		{
			name:    "standard YAML block",
			input:   "---\nkey: value\n---\n\n# Body",
			wantIn:  "",
			wantOut: "# Body",
		},
		{
			name:    "no frontmatter",
			input:   "# Just content",
			wantIn:  "",
			wantOut: "# Just content",
		},
		{
			name:    "opening delimiter only",
			input:   "---\nno closing",
			wantIn:  "",
			wantOut: "---\nno closing",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := stripFrontmatter(tc.input)
			if tc.wantOut != "" && !strings.Contains(got, tc.wantOut) {
				t.Errorf("expected %q in output, got: %q", tc.wantOut, got)
			}
			if strings.HasPrefix(got, "---") && tc.wantOut != "---\nno closing" {
				t.Errorf("result starts with frontmatter delimiter: %q", got)
			}
		})
	}
}
