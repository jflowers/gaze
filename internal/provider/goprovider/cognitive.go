package goprovider

import (
	"fmt"

	"golang.org/x/tools/go/packages"

	"github.com/unbound-force/gaze/v2/internal/cognitive"
	"github.com/unbound-force/gaze/v2/internal/crap"
)

// GoCognitiveComplexityProvider implements
// crap.CognitiveComplexityProvider by loading Go packages and
// computing cognitive complexity via the cognitive package.
type GoCognitiveComplexityProvider struct{}

// NewCognitiveComplexityProvider creates a new
// GoCognitiveComplexityProvider.
func NewCognitiveComplexityProvider() *GoCognitiveComplexityProvider {
	return &GoCognitiveComplexityProvider{}
}

// Analyze computes cognitive complexity for all functions in the
// packages matched by patterns, rooted at rootDir. Test files are
// skipped; the per-file enumeration is delegated to
// cognitive.AnalyzeFile.
//
// Unlike the cyclomatic-complexity provider (which delegates to
// gocyclo's recursive directory walk), this provider passes the
// original package patterns (including ./...) directly to
// packages.Load with Dir set to rootDir so that wildcard patterns
// enumerate every matched package, not just the module-root package.
func (p *GoCognitiveComplexityProvider) Analyze(patterns []string, rootDir string) ([]crap.FunctionCognitiveComplexity, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
		Dir:   rootDir,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("loading packages for cognitive complexity: %w", err)
	}

	var results []crap.FunctionCognitiveComplexity
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			if testFileRegexp.MatchString(pkg.Fset.Position(file.Pos()).Filename) {
				continue
			}
			for _, fr := range cognitive.AnalyzeFile(pkg.Fset, file) {
				results = append(results, crap.FunctionCognitiveComplexity{
					Package:             fr.Package,
					Function:            fr.Function,
					File:                fr.File,
					Line:                fr.Line,
					CognitiveComplexity: fr.CognitiveComplexity,
				})
			}
		}
	}

	return results, nil
}
