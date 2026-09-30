// Package goprovider provides Go-specific implementations of the
// provider interfaces defined in internal/crap/provider.go. These
// adapters wrap existing Go analysis tooling (gocyclo, go test
// -coverprofile, analysis.LoadAndAnalyze, classify.Classify) behind
// language-neutral interfaces, enabling the universal scoring core
// to operate without direct Go-specific imports.
package goprovider

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/fzipp/gocyclo"
	"github.com/unbound-force/gaze/v2/internal/crap"
)

// testFileRegexp matches Go test files by suffix. Moved from
// crap/analyze.go per design decision D9 — it is specific to Go
// complexity analysis and not needed by the universal scoring core.
var testFileRegexp = regexp.MustCompile(`_test\.go$`)

// GoComplexityProvider implements crap.ComplexityProvider by wrapping
// gocyclo.Analyze(). It converts []gocyclo.Stat to
// []crap.FunctionComplexity, removing the go/token dependency from
// the scoring pipeline.
type GoComplexityProvider struct{}

// NewComplexityProvider creates a new GoComplexityProvider.
func NewComplexityProvider() *GoComplexityProvider {
	return &GoComplexityProvider{}
}

// Analyze computes cyclomatic complexity for all functions in the
// packages matched by patterns, rooted at rootDir. Uses
// crap.ResolvePatterns to convert Go package patterns to filesystem
// paths, then delegates to gocyclo.Analyze.
func (p *GoComplexityProvider) Analyze(patterns []string, rootDir string) ([]crap.FunctionComplexity, error) {
	absPaths, err := crap.ResolvePatterns(patterns, rootDir)
	if err != nil {
		return nil, err
	}

	stats := gocyclo.Analyze(absPaths, testFileRegexp)

	endLines := buildEndLineMap(absPaths)

	result := make([]crap.FunctionComplexity, len(stats))
	for i, stat := range stats {
		result[i] = crap.FunctionComplexity{
			Package:    stat.PkgName,
			Function:   stat.FuncName,
			File:       stat.Pos.Filename,
			Line:       stat.Pos.Line,
			EndLine:    endLines[stat.Pos.Filename][stat.Pos.Line],
			Complexity: stat.Complexity,
		}
	}

	return result, nil
}

func buildEndLineMap(paths []string) map[string]map[int]int {
	fset := token.NewFileSet()
	result := make(map[string]map[int]int)

	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			log.Warn("skipping path: stat failed", "path", p, "err", err)
			continue
		}
		if info.IsDir() {
			_ = filepath.Walk(p, func(path string, fi os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !fi.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
					files = append(files, path)
				}
				return nil
			})
		} else {
			files = append(files, p)
		}
	}

	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			log.Warn("skipping file: parse failed", "file", file, "err", err)
			continue
		}
		filename := fset.Position(f.Pos()).Filename
		lineMap := make(map[int]int)
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			startLine := fset.Position(fn.Pos()).Line
			endLine := fset.Position(fn.End()).Line
			lineMap[startLine] = endLine
		}
		result[filename] = lineMap
	}

	return result
}
