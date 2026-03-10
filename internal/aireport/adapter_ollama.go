package aireport

import (
	"context"
	"io"
	"net/http"
)

// OllamaAdapter invokes the ollama HTTP API to format the analysis payload.
// Unlike the claude and gemini adapters, this uses net/http rather than
// exec.Command because the ollama CLI has no system prompt flag.
type OllamaAdapter struct {
	config AdapterConfig

	// httpClient is the HTTP client to use for requests. When nil,
	// http.DefaultClient is used. Set this field in tests to inject a
	// custom client (e.g., one pointing at an httptest.Server).
	httpClient *http.Client
}

// Format implements AIAdapter. It POSTs the system prompt and payload to
// the ollama /api/generate endpoint and returns the response field.
func (a *OllamaAdapter) Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error) {
	// TODO(Phase 6): implement OllamaAdapter.
	return "", nil
}
