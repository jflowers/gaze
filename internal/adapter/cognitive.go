package adapter

import (
	"context"

	"github.com/unbound-force/gaze/v2/internal/crap"
	"github.com/unbound-force/gaze/v2/internal/protocol"
)

// ExternalCognitiveComplexityProvider implements
// crap.CognitiveComplexityProvider by calling the
// "cognitive_complexity" protocol method on an external analyzer.
type ExternalCognitiveComplexityProvider struct {
	client *protocol.Client
}

// NewExternalCognitiveComplexityProvider creates a cognitive
// complexity provider that delegates to the given protocol client.
func NewExternalCognitiveComplexityProvider(client *protocol.Client) *ExternalCognitiveComplexityProvider {
	return &ExternalCognitiveComplexityProvider{client: client}
}

// Analyze calls the "cognitive_complexity" protocol method and
// converts the response to []crap.FunctionCognitiveComplexity.
func (p *ExternalCognitiveComplexityProvider) Analyze(patterns []string, rootDir string) ([]crap.FunctionCognitiveComplexity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), protocol.AnalysisTimeout)
	defer cancel()

	result, err := callAndUnmarshal[protocol.CognitiveComplexityResult](ctx, p.client, protocol.MethodCognitiveComplexity, protocol.CognitiveComplexityParams{
		RootPath: rootDir,
		Patterns: patterns,
	})
	if err != nil {
		return nil, err
	}

	return convertCognitiveComplexity(result.Functions), nil
}

// convertCognitiveComplexity maps protocol
// FunctionCognitiveComplexityData to crap.FunctionCognitiveComplexity.
func convertCognitiveComplexity(funcs []protocol.FunctionCognitiveComplexityData) []crap.FunctionCognitiveComplexity {
	out := make([]crap.FunctionCognitiveComplexity, len(funcs))
	for i, f := range funcs {
		out[i] = crap.FunctionCognitiveComplexity{
			Package:             f.Package,
			Function:            f.Name,
			File:                f.File,
			Line:                f.Line,
			CognitiveComplexity: f.CognitiveComplexity,
		}
	}
	return out
}
