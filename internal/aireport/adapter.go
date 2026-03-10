package aireport

import (
	"context"
	"fmt"
	"io"
	"time"
)

// AIAdapter formats an analysis payload using an external AI CLI or API.
// Implementations must be safe to call with a context that may be cancelled.
type AIAdapter interface {
	// Format invokes the AI integration with the given system prompt and
	// JSON payload (from payload io.Reader), returning the formatted
	// markdown report or an error.
	Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error)
}

// AdapterConfig holds the user-specified AI adapter configuration.
type AdapterConfig struct {
	// Name is the adapter identifier: "claude", "gemini", or "ollama".
	Name string

	// Model is the model name to use. Required for ollama; optional for
	// claude and gemini (uses each CLI's default when empty).
	Model string

	// Timeout is the maximum duration to wait for the AI adapter to respond.
	// Applied to the subprocess or HTTP request context.
	// Default: 10 minutes.
	Timeout time.Duration

	// OllamaHost overrides the default ollama server URL.
	// Reads from OLLAMA_HOST env var when empty.
	OllamaHost string
}

// validAdapters is the exact allowlist of supported AI adapter names.
var validAdapters = map[string]bool{
	"claude": true,
	"gemini": true,
	"ollama": true,
}

// NewAdapter creates an AIAdapter for the given config. It returns an error
// if cfg.Name is not in the allowlist {"claude", "gemini", "ollama"}.
func NewAdapter(cfg AdapterConfig) (AIAdapter, error) {
	if !validAdapters[cfg.Name] {
		return nil, fmt.Errorf(
			"unknown AI adapter %q: must be one of \"claude\", \"gemini\", or \"ollama\"",
			cfg.Name,
		)
	}
	switch cfg.Name {
	case "claude":
		return &ClaudeAdapter{config: cfg}, nil
	case "gemini":
		return &GeminiAdapter{config: cfg}, nil
	case "ollama":
		return &OllamaAdapter{config: cfg}, nil
	}
	// Unreachable — validAdapters check above covers all cases.
	panic("aireport: unreachable adapter case")
}

// FakeAdapterCall records one invocation of FakeAdapter.Format.
type FakeAdapterCall struct {
	// SystemPrompt is the system prompt passed to Format.
	SystemPrompt string
	// Payload is the full payload bytes read from the io.Reader argument.
	Payload []byte
}

// FakeAdapter is an AIAdapter for use in tests.
type FakeAdapter struct {
	// Response is the string returned by Format.
	Response string
	// Err is the error returned by Format. When non-nil, Response is ignored.
	Err error
	// Calls records each invocation of Format for assertion in tests.
	Calls []FakeAdapterCall
}

// Format implements AIAdapter. It records the call and returns the configured
// Response or Err.
func (f *FakeAdapter) Format(_ context.Context, systemPrompt string, payload io.Reader) (string, error) {
	var payloadBytes []byte
	if payload != nil {
		payloadBytes, _ = io.ReadAll(payload)
	}
	f.Calls = append(f.Calls, FakeAdapterCall{
		SystemPrompt: systemPrompt,
		Payload:      payloadBytes,
	})
	if f.Err != nil {
		return "", f.Err
	}
	return f.Response, nil
}
