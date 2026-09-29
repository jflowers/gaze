package docsignal

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/unbound-force/gaze/v2/internal/docscan"
)

// typeDecl is an unexported intermediate type representing a parsed
// grammar annotation before fan-out to per-function tuples.
type typeDecl struct {
	typeName   string
	label      string // "contractual" or "incidental"
	source     string // SourceArchitectureDoc or SourceReadme
	sourceFile string
	weight     int // signed: positive for contractual, negative for incidental
	priority   docscan.Priority
	excerpt    string
	reasoning  string
}

// parseAnnotations extracts type-level grammar annotations from a
// single docscan.DocumentFile. Returns nil if no annotations are found.
// Malformed annotations are silently skipped; valid annotations in the
// same document are still extracted.
func parseAnnotations(doc docscan.DocumentFile) []typeDecl {
	baseName := filepath.Base(doc.Path)
	isReadme := strings.HasPrefix(strings.ToUpper(baseName), "README")

	source := SourceArchitectureDoc
	absWeight := ArchitectureDocWeight
	if isReadme {
		source = SourceReadme
		absWeight = ReadmeWeight
	}

	var decls []typeDecl

	// Scan line by line for both annotation forms.
	lines := strings.Split(doc.Content, "\n")
	inFencedBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Fenced block callout start.
		if fencedBlockStartPattern.MatchString(trimmed) {
			inFencedBlock = true
			continue
		}

		if inFencedBlock {
			// A declaration line inside a fenced block.
			if m := fencedBlockEntryPattern.FindStringSubmatch(line); m != nil {
				typeName := m[1]
				label := strings.ToLower(m[2])
				decls = append(decls, newTypeDecl(typeName, label, source, absWeight, doc, strings.TrimSpace(line)))
				continue
			}
			// A non-entry line still prefixed with ">" (e.g. a wrapped
			// continuation) keeps the block open; anything else ends it.
			if strings.HasPrefix(strings.TrimSpace(line), ">") {
				continue
			}
			inFencedBlock = false
			continue
		}

		// HTML comment form.
		if m := htmlCommentPattern.FindStringSubmatch(trimmed); m != nil {
			label := strings.ToLower(m[1])
			typeName := m[2]
			decls = append(decls, newTypeDecl(typeName, label, source, absWeight, doc, trimmed))
		}
	}

	return decls
}

// newTypeDecl constructs a typeDecl, applying the label sign to the
// absolute weight and building the reasoning string.
func newTypeDecl(typeName, label, source string, absWeight int, doc docscan.DocumentFile, excerpt string) typeDecl {
	weight := absWeight
	if label == "incidental" {
		weight = -absWeight
	}
	return typeDecl{
		typeName:   typeName,
		label:      label,
		source:     source,
		sourceFile: doc.Path,
		weight:     weight,
		priority:   doc.Priority,
		excerpt:    excerpt,
		reasoning:  makeReasoning(source, typeName, label),
	}
}

// resolveConflicts applies priority-based conflict precedence:
// PriorityModuleRoot wins over PriorityOther for the same type name.
// On a tie, the first declaration is kept. The result is sorted by
// type name for deterministic output.
func resolveConflicts(decls []typeDecl) []typeDecl {
	// Group by type name, keeping the highest-priority declaration.
	best := make(map[string]typeDecl)
	for _, d := range decls {
		key := d.typeName
		if existing, ok := best[key]; ok {
			// Lower priority value = higher priority.
			if d.priority < existing.priority {
				best[key] = d
			}
		} else {
			best[key] = d
		}
	}

	result := make([]typeDecl, 0, len(best))
	for _, d := range best {
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].typeName < result[j].typeName
	})
	return result
}

// makeReasoning constructs a human-readable reasoning string for a
// grammar-derived signal.
func makeReasoning(source, typeName, label string) string {
	switch source {
	case SourceArchitectureDoc:
		return fmt.Sprintf("design doc declares %s as %s", typeName, label)
	case SourceReadme:
		return fmt.Sprintf("readme declares %s as %s", typeName, label)
	default:
		return fmt.Sprintf("doc declares %s as %s", typeName, label)
	}
}
