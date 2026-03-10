package aireport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GeminiAdapter invokes the gemini CLI to format the analysis payload.
// The system prompt is written as GEMINI.md in a temporary directory
// (cmd.Dir is set to that directory) because the gemini CLI reads its
// system prompt from this file rather than a command-line flag.
type GeminiAdapter struct {
	config AdapterConfig
}

// geminiOutput is the parsed JSON response from gemini --output-format json.
type geminiOutput struct {
	Response string `json:"response"`
}

// Format implements AIAdapter. It creates a temp directory with GEMINI.md,
// invokes gemini with the payload on stdin, parses the JSON response, and
// returns the formatted report string.
//
// Arguments are passed as distinct Go strings — no shell interpolation.
// The temp directory is removed in a deferred cleanup after the subprocess exits.
// Returns an error when:
//   - gemini is not found on PATH (FR-012).
//   - The subprocess exits non-zero.
//   - The response field is empty or whitespace (FR-016).
func (a *GeminiAdapter) Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error) {
	// Verify gemini is available (FR-012).
	geminiPath, err := exec.LookPath("gemini")
	if err != nil {
		return "", fmt.Errorf("gemini not found on PATH (FR-012): %w", err)
	}

	// Create temp directory and write system prompt as GEMINI.md.
	tmpDir, err := os.MkdirTemp("", "gaze-gemini-*")
	if err != nil {
		return "", fmt.Errorf("creating temp directory for GEMINI.md: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	geminiMD := filepath.Join(tmpDir, "GEMINI.md")
	if err := os.WriteFile(geminiMD, []byte(systemPrompt), 0644); err != nil {
		return "", fmt.Errorf("writing GEMINI.md: %w", err)
	}

	// Build args: -p "" (headless), --output-format json, [-m <model>]
	args := []string{"-p", "", "--output-format", "json"}
	if a.config.Model != "" {
		args = append(args, "-m", a.config.Model)
	}

	cmd := exec.CommandContext(ctx, geminiPath, args...)
	cmd.Dir = tmpDir
	cmd.Stdin = payload

	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		return "", fmt.Errorf("gemini exited with error: %w\nstderr: %s", err, stderr)
	}

	var resp geminiOutput
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("parsing gemini JSON response: %w", err)
	}

	if strings.TrimSpace(resp.Response) == "" {
		return "", fmt.Errorf("gemini returned empty response field (FR-016): ensure the gemini CLI is working correctly")
	}
	return resp.Response, nil
}
