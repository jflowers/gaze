// Package docsignal derives classification signals from Markdown
// documentation files already discovered by docscan.Scan, and from
// an optional sidecar file (.uf/gaze/contracts.yaml or .json).
//
// Doc-derived signals supplement protocol-derived signals for external
// analyzers. Go-native projects use internal/classify/godoc.go and
// are not affected by this package.
package docsignal

import (
	"regexp"

	"github.com/unbound-force/gaze/v2/internal/taxonomy"
)

// DerivedSignal binds a taxonomy.Signal to the (package, function,
// side_effect_type) tuple it applies to. At the mergeClassifications
// boundary, each DerivedSignal is reduced to its Signal field and
// grouped alongside protocol-derived signals.
type DerivedSignal struct {
	Package        string
	Function       string
	SideEffectType string
	Signal         taxonomy.Signal
}

// Grammar annotation patterns for the two supported syntax forms.
//
// HTML comment form: <!-- gaze:contractual|incidental TypeName -->
// The gaze: prefix and label tokens are matched case-insensitively.
// The type name is captured verbatim (analyzer-emitted name).
var htmlCommentPattern = regexp.MustCompile(
	`(?i)<!--\s*gaze:(contractual|incidental)\s+(\S+)\s*-->`,
)

// Fenced block callout start: > [!gaze-contract]
// Matched case-insensitively.
var fencedBlockStartPattern = regexp.MustCompile(
	`(?i)^>\s*\[!gaze-contract\]`,
)

// Fenced block entry: > - `TypeName` — contractual|incidental
// The type name is captured from backticks; the label is captured
// after the em dash. The human comment after the label is ignored.
// Label tokens are matched case-insensitively.
var fencedBlockEntryPattern = regexp.MustCompile(
	"(?i)> - `([^`]+)` — (contractual|incidental)",
)

// Filename-based weight constants. The source token and weight are
// determined by the document's basename, not by its priority tier.
//
// Design decision D2: non-README docs carry higher weight than READMEs
// because design docs are typically more authoritative about intent.
// Sidecar weight is highest to reflect explicit, intentional declarations.
const (
	// ArchitectureDocWeight is the contractual/incidental weight for
	// annotations in non-README documents (design docs, ADRs, etc.).
	ArchitectureDocWeight = 25

	// ReadmeWeight is the contractual/incidental weight for annotations
	// in documents whose basename starts with "README" (case-insensitive).
	ReadmeWeight = 15

	// SidecarWeight is the contractual/incidental weight for entries
	// in the .uf/gaze/contracts.yaml or .json sidecar file.
	SidecarWeight = 30
)

// Source tokens identify the origin of a derived signal.
// Signal.Source carries the bare token; the document path is in
// Signal.SourceFile.
const (
	// SourceArchitectureDoc identifies signals derived from non-README
	// Markdown documents (design docs, ADRs, etc.).
	SourceArchitectureDoc = "architecture_doc"

	// SourceReadme identifies signals derived from README documents.
	SourceReadme = "readme"

	// SourceSidecar identifies signals derived from the sidecar file
	// (.uf/gaze/contracts.yaml or .json).
	SourceSidecar = "sidecar"
)

// sidecarFileName is the base filename (without extension) of the
// sidecar contracts file.
const sidecarFileName = "contracts"

// sidecarDir is the directory path (relative to module root) where
// the sidecar file resides.
const sidecarDir = ".uf/gaze"
