package aireport

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ClaudeAdapter invokes the claude CLI to format the analysis payload.
// The system prompt is written to a temporary file and passed via
// --system-prompt-file to avoid OS argument length limits (~13 KB prompt).
type ClaudeAdapter struct {
	config AdapterConfig
}

// Compile-time check that ClaudeAdapter implements AdapterValidator.
var _ AdapterValidator = &ClaudeAdapter{}

// ValidateBinary implements AdapterValidator. It checks that the claude binary
// is available on PATH (FR-012). Call this before running the analysis pipeline
// to give users an immediate error rather than failing after minutes of work.
func (a *ClaudeAdapter) ValidateBinary() error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude not found on PATH (FR-012): %w", err)
	}
	return nil
}

// Format implements AIAdapter. It writes the system prompt to a temp file,
// invokes claude with the payload on stdin, and returns the formatted report.
//
// Arguments are passed as distinct Go strings — no shell interpolation.
// The temp file is removed in a deferred cleanup after the subprocess exits.
// Returns an error when:
//   - claude is not found on PATH (detected by ValidateBinary; also checked here as defense).
//   - The subprocess exits non-zero.
//   - The output is empty or whitespace (FR-016).
func (a *ClaudeAdapter) Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error) {
	// Write system prompt to a temporary directory with explicit 0600 permissions.
	tmpDir, err := os.MkdirTemp("", "gaze-claude-prompt-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir for system prompt: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpPath := filepath.Join(tmpDir, "prompt.md")
	if err := os.WriteFile(tmpPath, []byte(systemPrompt), 0600); err != nil {
		return "", fmt.Errorf("writing system prompt to temp file: %w", err)
	}

	// Build args: -p "" (headless), --system-prompt-file <path>, [--model <name>]
	args := []string{"-p", "", "--system-prompt-file", tmpPath}
	if a.config.Model != "" {
		args = append(args, "--model", a.config.Model)
	}

	outBytes, err := runSubprocess(ctx, "claude", args, "", payload)
	if err != nil {
		return "", err
	}

	result := string(outBytes)
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("claude returned empty output (FR-016): ensure the claude CLI is working correctly")
	}
	return result, nil
}
