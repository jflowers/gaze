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

// Compile-time check that GeminiAdapter implements AdapterValidator.
var _ AdapterValidator = &GeminiAdapter{}

// geminiOutput is the parsed JSON response from gemini --output-format json.
type geminiOutput struct {
	Response string `json:"response"`
}

// ValidateBinary implements AdapterValidator. It checks that the gemini binary
// is available on PATH (FR-012). Call this before running the analysis pipeline
// to give users an immediate error rather than failing after minutes of work.
func (a *GeminiAdapter) ValidateBinary() error {
	if _, err := exec.LookPath("gemini"); err != nil {
		return fmt.Errorf("gemini not found on PATH (FR-012): %w", err)
	}
	return nil
}

// Format implements AIAdapter. It creates a temp directory with GEMINI.md,
// invokes gemini with the payload on stdin, parses the JSON response, and
// returns the formatted report string.
//
// Arguments are passed as distinct Go strings — no shell interpolation.
// The temp directory is removed in a deferred cleanup after the subprocess exits.
// Returns an error when:
//   - gemini is not found on PATH (detected by ValidateBinary; also checked here as defense).
//   - The subprocess exits non-zero.
//   - The response field is empty or whitespace (FR-016).
func (a *GeminiAdapter) Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error) {
	// Create temp directory and write system prompt as GEMINI.md.
	tmpDir, err := os.MkdirTemp("", "gaze-gemini-*")
	if err != nil {
		return "", fmt.Errorf("creating temp directory for GEMINI.md: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	geminiMD := filepath.Join(tmpDir, "GEMINI.md")
	if err := os.WriteFile(geminiMD, []byte(systemPrompt), 0600); err != nil {
		return "", fmt.Errorf("writing GEMINI.md: %w", err)
	}

	// Build args: -p "" (headless), --output-format json, [-m <model>]
	args := []string{"-p", "", "--output-format", "json"}
	if a.config.Model != "" {
		args = append(args, "-m", a.config.Model)
	}

	// Gemini reads GEMINI.md from the working directory, so cmdDir = tmpDir.
	outBytes, err := runSubprocess(ctx, "gemini", args, tmpDir, payload)
	if err != nil {
		return "", err
	}

	var resp geminiOutput
	if err := json.Unmarshal(outBytes, &resp); err != nil {
		return "", fmt.Errorf("parsing gemini JSON response: %w", err)
	}

	if strings.TrimSpace(resp.Response) == "" {
		return "", fmt.Errorf("gemini returned empty response field (FR-016): ensure the gemini CLI is working correctly")
	}
	return resp.Response, nil
}
