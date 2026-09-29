package docsignal

import (
	"io"
	"sort"

	"github.com/unbound-force/gaze/v2/internal/docscan"
	"github.com/unbound-force/gaze/v2/internal/taxonomy"
)

// tupleKey identifies a (package, function, side_effect_type) tuple.
type tupleKey struct {
	pkg      string
	function string
	seType   string
}

// DeriveSignals orchestrates the full doc-signal derivation pipeline:
//
//  1. Parse grammar annotations from each document (type-level
//     declarations with filename-based source/weight and priority-based
//     conflict precedence).
//  2. Fan out each type-level annotation to every (package, function)
//     tuple in cached whose effect set contains the annotated type.
//  3. Load the sidecar file (if any) and merge with sidecar precedence:
//     a sidecar entry for the same tuple discards the grammar-derived
//     signal for that tuple.
//
// Every step degrades gracefully (design D5): no docs, no annotations,
// no cached effects, or no sidecar file simply produce an empty or
// partial result. The returned slice is sorted by
// (Package, Function, SideEffectType, Source) for deterministic output.
//
// stderr receives sidecar warnings (malformed file, skipped entries). It
// may be nil to suppress warnings. It is threaded through to LoadSidecar so
// callers can capture diagnostics through the injected writer rather than
// the process's os.Stderr.
func DeriveSignals(
	docs []docscan.DocumentFile,
	cached []taxonomy.AnalysisResult,
	moduleRoot string,
	stderr io.Writer,
) []DerivedSignal {
	// 1. Parse grammar annotations across all documents.
	var allDecls []typeDecl
	for _, doc := range docs {
		allDecls = append(allDecls, parseAnnotations(doc)...)
	}

	// 2. Resolve priority conflicts (PriorityModuleRoot > PriorityOther).
	decls := resolveConflicts(allDecls)

	// 3. Fan out type-level annotations to per-function tuples.
	grammarSignals := fanOut(decls, cached)

	// 4. Load the sidecar file (warnings go to stderr; errors degrade
	// gracefully to grammar-only signals).
	sidecarSignals, _ := LoadSidecar(moduleRoot, stderr)

	// 5. Merge with sidecar precedence.
	return mergeWithSidecar(grammarSignals, sidecarSignals)
}

// fanOut expands type-level declarations into per-function DerivedSignals
// using the cached analysis results. Duplicate tuples (e.g. a function
// with multiple effects of the same type) are emitted once.
func fanOut(decls []typeDecl, cached []taxonomy.AnalysisResult) []DerivedSignal {
	// Map type name → declaration for O(1) lookup during fan-out.
	byType := make(map[string]typeDecl, len(decls))
	for _, d := range decls {
		byType[d.typeName] = d
	}

	seen := make(map[tupleKey]bool)
	var signals []DerivedSignal
	for _, r := range cached {
		for _, se := range r.SideEffects {
			decl, ok := byType[string(se.Type)]
			if !ok {
				continue
			}
			key := tupleKey{
				pkg:      r.Target.Package,
				function: r.Target.Function,
				seType:   string(se.Type),
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			signals = append(signals, DerivedSignal{
				Package:        r.Target.Package,
				Function:       r.Target.Function,
				SideEffectType: string(se.Type),
				Signal: taxonomy.Signal{
					Source:     decl.source,
					SourceFile: decl.sourceFile,
					Weight:     decl.weight,
					Excerpt:    decl.excerpt,
					Reasoning:  decl.reasoning,
				},
			})
		}
	}
	return signals
}

// mergeWithSidecar combines grammar-derived and sidecar signals, giving
// sidecar entries precedence over grammar-derived signals for the same
// tuple. The result is sorted deterministically.
func mergeWithSidecar(grammar, sidecar []DerivedSignal) []DerivedSignal {
	sidecarKeys := make(map[tupleKey]bool, len(sidecar))
	result := make([]DerivedSignal, 0, len(grammar)+len(sidecar))

	for _, s := range sidecar {
		sidecarKeys[tupleKey{s.Package, s.Function, s.SideEffectType}] = true
		result = append(result, s)
	}
	for _, g := range grammar {
		if sidecarKeys[tupleKey{g.Package, g.Function, g.SideEffectType}] {
			continue // sidecar takes precedence
		}
		result = append(result, g)
	}

	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.Package != b.Package {
			return a.Package < b.Package
		}
		if a.Function != b.Function {
			return a.Function < b.Function
		}
		if a.SideEffectType != b.SideEffectType {
			return a.SideEffectType < b.SideEffectType
		}
		return a.Signal.Source < b.Signal.Source
	})

	return result
}
