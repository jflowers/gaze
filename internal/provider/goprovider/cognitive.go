package goprovider

import (
	"go/ast"

	"github.com/unbound-force/gaze/internal/cognitive"
	"github.com/unbound-force/gaze/internal/crap"
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
// packages matched by patterns, rooted at rootDir.
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
			fresults := analyzeCognitiveFile(pkg, file)
			results = append(results, fresults...)
		}
	}

	return results, nil
}

func analyzeCognitiveFile(pkg *packages.Package, file *ast.File) []crap.FunctionCognitiveComplexity {
	var results []crap.FunctionCognitiveComplexity
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}
		if testFileRegexp.MatchString(pkg.Fset.Position(file.Pos()).Filename) {
			continue
		}
		cc := cognitive.AnalyzeFunc(pkg.Fset, funcDecl)
		pos := pkg.Fset.Position(funcDecl.Pos())
		results = append(results, crap.FunctionCognitiveComplexity{
			Package:             pkg.Name,
			Function:            funcDecl.Name.Name,
			File:                pos.Filename,
			Line:                pos.Line,
			CognitiveComplexity: cc,
		})
	}
	return results
}
