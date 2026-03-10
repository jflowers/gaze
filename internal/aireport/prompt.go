package aireport

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
)

// defaultPrompt is the embedded gaze-reporter.md with YAML frontmatter
// already stripped. It is used when no local .opencode/agents/gaze-reporter.md
// is found in the working directory.
//
//go:embed assets/agents/gaze-reporter.md
var defaultPromptRaw string

// LoadPrompt returns the system prompt to use with the AI adapter.
//
// It first looks for .opencode/agents/gaze-reporter.md relative to workdir.
// If found, it reads that file and strips any YAML frontmatter before
// returning the content.
//
// If the local file is absent, it falls back to the embedded default
// (the gaze-reporter.md bundled at build time with frontmatter stripped).
func LoadPrompt(workdir string) (string, error) {
	localPath := filepath.Join(workdir, ".opencode", "agents", "gaze-reporter.md")
	data, err := os.ReadFile(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			return stripFrontmatter(defaultPromptRaw), nil
		}
		return "", err
	}
	return stripFrontmatter(string(data)), nil
}

// stripFrontmatter removes the YAML frontmatter block from content.
//
// Frontmatter is defined as the block starting with the first line "---"
// through the next line "---" (inclusive). If the content does not begin
// with a frontmatter block the content is returned unchanged.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	// Find the closing delimiter after the opening "---\n".
	rest := content[3:]
	// Skip the optional newline after the opening "---".
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 0 && rest[0] == '\r' && len(rest) > 1 && rest[1] == '\n' {
		rest = rest[2:]
	}

	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		// No closing delimiter; return original content.
		return content
	}
	// Skip past "\n---" and any trailing newline.
	after := rest[idx+4:]
	if len(after) > 0 && after[0] == '\n' {
		after = after[1:]
	} else if len(after) > 0 && after[0] == '\r' && len(after) > 1 && after[1] == '\n' {
		after = after[2:]
	}
	return after
}
