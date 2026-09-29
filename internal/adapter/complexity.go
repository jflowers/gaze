// Package adapter implements external analyzer provider adapters
// that translate JSON-RPC protocol responses into the Phase 1
// provider interface types (crap.ComplexityProvider,
// crap.LineCoverageProvider, crap.SideEffectAnalyzer,
// crap.ContractCoverageProvider).
//
// Each adapter holds a reference to a protocol.Client and calls the
// appropriate protocol method in its interface method. The adapters
// are constructed by Session after the initialize handshake.
//
// Design decision D4: External adapters implement Phase 1 interfaces.
package adapter

import (
	"context"
	"path/filepath"

	"github.com/unbound-force/gaze/v2/internal/crap"
	"github.com/unbound-force/gaze/v2/internal/protocol"
)

// ExternalComplexityProvider implements crap.ComplexityProvider by
// calling the "complexity" protocol method on an external analyzer.
type ExternalComplexityProvider struct {
	client    *protocol.Client
	testFiles map[string]bool
}

// NewExternalComplexityProvider creates a complexity provider that delegates
// to the given protocol client and optionally filters test files.
func NewExternalComplexityProvider(client *protocol.Client, testFiles map[string]bool) *ExternalComplexityProvider {
	return &ExternalComplexityProvider{client: client, testFiles: testFiles}
}

// Analyze calls the "complexity" protocol method and converts the
// response to []crap.FunctionComplexity.
func (p *ExternalComplexityProvider) Analyze(patterns []string, rootDir string) ([]crap.FunctionComplexity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), protocol.AnalysisTimeout)
	defer cancel()

	result, err := callAndUnmarshal[protocol.ComplexityResult](ctx, p.client, protocol.MethodComplexity, protocol.ComplexityParams{
		RootPath: rootDir,
		Patterns: patterns,
	})
	if err != nil {
		return nil, err
	}

	funcs := convertComplexity(result.Functions)
	return filterTestFiles(funcs, p.testFiles), nil
}

// isTestFile reports whether path is in the test-file set after cleaning
// both sides. A nil or empty set returns false (D3).
func isTestFile(path string, testFiles map[string]bool) bool {
	return testFiles[filepath.Clean(path)]
}

// filterTestFiles removes functions whose cleaned file path is in
// testFiles. A nil or empty testFiles set disables filtering (D1).
func filterTestFiles(funcs []crap.FunctionComplexity, testFiles map[string]bool) []crap.FunctionComplexity {
	if len(testFiles) == 0 {
		return funcs
	}
	filtered := funcs[:0:0]
	for _, f := range funcs {
		if isTestFile(f.File, testFiles) {
			continue
		}
		filtered = append(filtered, f)
	}
	return filtered
}

// convertComplexity maps protocol FunctionComplexityData to
// crap.FunctionComplexity.
func convertComplexity(funcs []protocol.FunctionComplexityData) []crap.FunctionComplexity {
	out := make([]crap.FunctionComplexity, len(funcs))
	for i, f := range funcs {
		out[i] = crap.FunctionComplexity{
			Package:    f.Package,
			Function:   f.Name,
			File:       f.File,
			Line:       f.Line,
			Complexity: f.Complexity,
		}
	}
	return out
}
