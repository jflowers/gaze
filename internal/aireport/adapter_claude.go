package aireport

import (
	"context"
	"io"
)

// ClaudeAdapter invokes the claude CLI to format the analysis payload.
// The system prompt is written to a temporary file and passed via
// --system-prompt-file to avoid OS argument length limits.
type ClaudeAdapter struct {
	config AdapterConfig
}

// Format implements AIAdapter. It writes the system prompt to a temp file,
// invokes claude with the payload on stdin, and returns the formatted report.
func (a *ClaudeAdapter) Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error) {
	// TODO(Phase 4): implement ClaudeAdapter.
	return "", nil
}
