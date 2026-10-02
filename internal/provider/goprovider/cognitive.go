package goprovider

import (
	"github.com/unbound-force/gaze/v2/internal/cognitive"
	"github.com/unbound-force/gaze/v2/internal/crap"
	"golang.org/x/tools/go/packages"
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
func (p *GoCognitiveComplexityProvider) Analyze(patterns []string, rootDir string) ([]crap.FunctionCognitiveComplexity, error) {
	absPaths, err := crap.ResolvePatterns(patterns, rootDir)
	if err != nil {
		return nil, err
	}

	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, absPaths...)
	if err != nil {
		return nil, err
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
