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

	for _, line := range strings.Split(diffOutput, "\n") {
		if path, ok := diffHeaderPath(line); ok {
			flushPending(&current, &pendingPath)
			flushCurrent(&changes, &current)
			pendingPath = path
			continue
		}

		if path, ok := fileHeaderPath(line, pendingPath); ok {
			if path == "" {
				flushCurrent(&changes, &current)
				pendingPath = ""
				continue
			}
			flushCurrent(&changes, &current)
			current = &FileChange{Path: path}
			pendingPath = ""
			continue
		}

		if current == nil {
			continue
		}

		if r, ok := parseHunkHeader(line); ok {
			current.Ranges = append(current.Ranges, r)
		}
	}

	flushPending(&current, &pendingPath)
	if current != nil {
		changes = append(changes, *current)
	}
	return changes
}

// diffHeaderPath extracts the target path from a `diff --git` header
// line. It returns the path and whether the line matched.
func diffHeaderPath(line string) (string, bool) {
	m := diffLineRe.FindStringSubmatch(line)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// fileHeaderPath extracts the file path from a `+++ b/` header line,
// falling back to pendingPath for `+++ /dev/null` deletion headers.
// It returns the resolved path and whether the line matched.
func fileHeaderPath(line, pendingPath string) (string, bool) {
	m := fileHeaderRe.FindStringSubmatch(line)
	if m == nil {
		return "", false
	}
	path := m[1]
	if path == "" {
		path = pendingPath
	}
	return path, true
}

// parseHunkHeader parses an `@@ -a,b +c,d @@` hunk header line. It
// returns the added-line range it describes and whether the line was
// a hunk header with a non-zero added-line count.
func parseHunkHeader(line string) (LineRange, bool) {
	m := hunkHeaderRe.FindStringSubmatch(line)
	if m == nil {
		return LineRange{}, false
	}
	start, _ := strconv.Atoi(m[1])
	count := 1
	if m[2] != "" {
		count, _ = strconv.Atoi(m[2])
	}
	if count == 0 {
		return LineRange{}, false
	}
	return LineRange{Start: start, End: start + count - 1}, true
}

// flushPending creates a FileChange from a pending path when no
// explicit `+++ b/` header has done so yet.
func flushPending(current **FileChange, pendingPath *string) {
	if *pendingPath != "" && *current == nil {
		*current = &FileChange{Path: *pendingPath}
		*pendingPath = ""
	}
}

// flushCurrent appends the current FileChange, if any, to changes and
// clears it.
func flushCurrent(changes *[]FileChange, current **FileChange) {
	if *current != nil {
		*changes = append(*changes, **current)
		*current = nil
	}
}
