package aireport

import (
	"context"
	"io"
)

// GeminiAdapter invokes the gemini CLI to format the analysis payload.
// The system prompt is written as GEMINI.md in a temporary directory
// (cmd.Dir is set to that directory) because the gemini CLI reads its
// system prompt from this file rather than a command-line flag.
type GeminiAdapter struct {
	config AdapterConfig
}

// Format implements AIAdapter. It creates a temp directory with GEMINI.md,
// invokes gemini with the payload on stdin, parses the JSON response, and
// returns the formatted report string.
func (a *GeminiAdapter) Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error) {
	// TODO(Phase 5): implement GeminiAdapter.
	return "", nil
}
