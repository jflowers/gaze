// Package diff parses git diff output to identify changed files and
// line ranges. It provides a pure-function parser (Parse) and a
// Differ interface for dependency injection of the git invocation.
package diff

import (
	"regexp"
	"strconv"
	"strings"
)

// LineRange represents a contiguous range of changed lines.
type LineRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// FileChange represents a single file's changes in a diff.
type FileChange struct {
	Path   string      `json:"path"`
	Ranges []LineRange `json:"ranges"`
}

// Differ produces diff output for a given reference.
type Differ interface {
	Diff(ref string) (string, error)
}

var (
	fileHeaderRe = regexp.MustCompile(`^\+\+\+ (?:b/|/dev/null)(.+)?$`)
	diffLineRe   = regexp.MustCompile(`^diff --git a/.+ b/(.+)$`)
	hunkHeaderRe = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)
)

// Parse parses `git diff --unified=0` output into structured
// FileChange records. It extracts file paths from `+++ b/` headers
// (or `diff --git` lines for binary/deleted files) and line ranges
// from `@@ ... +N,M @@` hunk headers.
func Parse(diffOutput string) []FileChange {
	if strings.TrimSpace(diffOutput) == "" {
		return nil
	}

	var changes []FileChange
	var current *FileChange
	var pendingPath string

	flushPending := func() {
		if pendingPath != "" && current == nil {
			current = &FileChange{Path: pendingPath}
			pendingPath = ""
		}
	}

	for _, line := range strings.Split(diffOutput, "\n") {
		if m := diffLineRe.FindStringSubmatch(line); m != nil {
			flushPending()
			if current != nil {
				changes = append(changes, *current)
				current = nil
			}
			pendingPath = m[1]
			continue
		}

		if m := fileHeaderRe.FindStringSubmatch(line); m != nil {
			path := m[1]
			if path == "" {
				path = pendingPath
			}
			if path == "" {
				if current != nil {
					changes = append(changes, *current)
					current = nil
				}
				pendingPath = ""
				continue
			}
			if current != nil {
				changes = append(changes, *current)
			}
			current = &FileChange{Path: path}
			pendingPath = ""
			continue
		}

		if current == nil {
			continue
		}

		if m := hunkHeaderRe.FindStringSubmatch(line); m != nil {
			start, _ := strconv.Atoi(m[1])
			count := 1
			if m[2] != "" {
				count, _ = strconv.Atoi(m[2])
			}
			if count == 0 {
				continue
			}
			end := start + count - 1
			current.Ranges = append(current.Ranges, LineRange{Start: start, End: end})
		}
	}

	flushPending()
	if current != nil {
		changes = append(changes, *current)
	}

	return changes
}
