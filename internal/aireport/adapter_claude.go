package aireport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	// Resolve claude path (defense-in-depth: ValidateBinary should have run first).
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude not found on PATH (FR-012): %w", err)
	}

	// Write system prompt to a temporary file.
	tmpFile, err := os.CreateTemp("", "gaze-claude-prompt-*.md")
	if err != nil {
		return "", fmt.Errorf("creating temp file for system prompt: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(systemPrompt); err != nil {
		_ = tmpFile.Close()
		return "", fmt.Errorf("writing system prompt to temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("closing system prompt temp file: %w", err)
	}

	// Build args: -p "" (headless), --system-prompt-file <path>, [--model <name>]
	args := []string{"-p", "", "--system-prompt-file", tmpFile.Name()}
	if a.config.Model != "" {
		args = append(args, "--model", a.config.Model)
	}

	cmd := exec.CommandContext(ctx, claudePath, args...)
	cmd.Stdin = payload

	// Capture stdout with a bounded pipe to prevent OOM on large outputs.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("creating stdout pipe for claude: %w", err)
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("starting claude: %w", err)
	}

	outBytes, readErr := io.ReadAll(io.LimitReader(stdoutPipe, maxAdapterOutputBytes))
	waitErr := cmd.Wait()

	if waitErr != nil {
		// Truncate stderr to avoid leaking secrets.
		stderrSnippet := stderrBuf.String()
		if len(stderrSnippet) > maxAdapterStderrBytes {
			stderrSnippet = stderrSnippet[:maxAdapterStderrBytes] + "... (truncated)"
		}
		return "", fmt.Errorf("claude exited with error: %w\nstderr: %s", waitErr, stderrSnippet)
	}
	if readErr != nil {
		return "", fmt.Errorf("reading claude output: %w", readErr)
	}

	result := string(outBytes)
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("claude returned empty output (FR-016): ensure the claude CLI is working correctly")
	}
	return result, nil
}
