package aireport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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

// ollamaRequest is the JSON body sent to /api/generate.
type ollamaRequest struct {
	Model  string `json:"model"`
	System string `json:"system"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaResponse is the parsed JSON response from /api/generate.
type ollamaResponse struct {
	Response string `json:"response"`
}

// Format implements AIAdapter. It POSTs the system prompt and payload to
// the ollama /api/generate endpoint and returns the response field.
//
// Returns an error when:
//   - --model is empty (FR-003: required for ollama).
//   - The HTTP request fails.
//   - The response status is non-2xx.
//   - The response field is empty or whitespace (FR-016).
func (a *OllamaAdapter) Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error) {
	if a.config.Model == "" {
		return "", fmt.Errorf("--model is required when using ollama (FR-003)")
	}

	// Limit payload read to maxAdapterOutputBytes to prevent OOM on oversized JSON.
	payloadBytes, err := io.ReadAll(io.LimitReader(payload, maxAdapterOutputBytes))
	if err != nil {
		return "", fmt.Errorf("reading payload: %w", err)
	}

	host := a.config.OllamaHost
	if host == "" {
		host = os.Getenv("OLLAMA_HOST")
	}
	if host == "" {
		host = "http://localhost:11434"
	}

	// Validate the host URL to prevent SSRF from a malformed OLLAMA_HOST value.
	// Restrict to http/https schemes to reject ftp://, file://, etc.
	baseURL, err := url.Parse(host)
	if err != nil || (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" {
		return "", fmt.Errorf("invalid ollama host URL %q: must be an absolute http or https URL (e.g. http://localhost:11434)", host)
	}
	generateURL := baseURL.JoinPath("/api/generate").String()

	reqBody := ollamaRequest{
		Model:  a.config.Model,
		System: systemPrompt,
		Prompt: string(payloadBytes),
		Stream: false,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshalling ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		generateURL, bytes.NewReader(reqBytes))
	if err != nil {
		return "", fmt.Errorf("building ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := a.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ollama HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Cap error body at maxAdapterStderrBytes — this is an error message, not a report.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxAdapterStderrBytes))
		return "", fmt.Errorf("ollama returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxAdapterOutputBytes)).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("parsing ollama response: %w", err)
	}

	if strings.TrimSpace(ollamaResp.Response) == "" {
		return "", fmt.Errorf("ollama returned empty response field (FR-016): ensure the model is loaded and working")
	}
	return ollamaResp.Response, nil
}
