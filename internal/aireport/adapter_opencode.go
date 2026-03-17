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

// OpenCodeAdapter invokes the opencode CLI to format the analysis payload.
// The system prompt is written as .opencode/agents/gaze-reporter.md in a
// temporary directory passed via --dir, because opencode reads its agent
// system prompt from that location when --agent is specified.
//
// This mirrors the Gemini adapter's temp-dir pattern adapted for opencode's
// agent file convention: the temp dir is created, the agent file is written,
// and --dir <tmpDir> --agent gaze-reporter are passed as CLI flags.
// The temporary directory is removed after the subprocess exits.
type OpenCodeAdapter struct {
	config AdapterConfig
}

// Compile-time check that OpenCodeAdapter implements AdapterValidator.
var _ AdapterValidator = &OpenCodeAdapter{}

// ValidateBinary implements AdapterValidator. It checks that the opencode
// binary is available on PATH (FR-007). Call this before running the analysis
// pipeline to give users an immediate error rather than failing after minutes
// of work.
func (a *OpenCodeAdapter) ValidateBinary() error {
	if _, err := exec.LookPath("opencode"); err != nil {
		return fmt.Errorf("opencode not found on PATH (FR-007): %w", err)
	}
	return nil
}

// Format implements AIAdapter. It creates a temp directory containing
// .opencode/agents/gaze-reporter.md (with empty YAML frontmatter prepended),
// invokes opencode run with the payload on stdin, and returns the formatted
// report string.
//
// Arguments are passed as distinct Go strings — no shell interpolation.
// The temp directory is removed in a deferred cleanup after the subprocess exits.
// Returns an error when:
//   - opencode is not found on PATH (detected by ValidateBinary; also checked here as defense).
//   - The temp directory or agent file cannot be created.
//   - The subprocess exits non-zero.
//   - The output is empty or whitespace (FR-009).
func (a *OpenCodeAdapter) Format(ctx context.Context, systemPrompt string, payload io.Reader) (string, error) {
	// Create temp directory.
	tmpDir, err := os.MkdirTemp("", "gaze-opencode-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir for opencode agent: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create the .opencode/agents/ subdirectory inside the temp dir.
	agentsDir := filepath.Join(tmpDir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0700); err != nil {
		return "", fmt.Errorf("creating .opencode/agents dir: %w", err)
	}

	// Write the agent definition file with empty YAML frontmatter prepended.
	agentFile := filepath.Join(agentsDir, "gaze-reporter.md")
	content := "---\n---\n" + systemPrompt
	if err := os.WriteFile(agentFile, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("writing agent file: %w", err)
	}

	// Build args: run (subcommand), --dir <tmpDir>, --agent gaze-reporter,
	// --format default (plain-text stdout), [-m <model>], "" (headless trigger).
	args := []string{"run", "--dir", tmpDir, "--agent", "gaze-reporter", "--format", "default"}
	if a.config.Model != "" {
		args = append(args, "-m", a.config.Model)
	}
	args = append(args, "")

	outBytes, stderrBytes, err := runSubprocess(ctx, "opencode", args, "", payload)
	if err != nil {
		return "", err
	}

	result := string(outBytes)
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("opencode returned empty output (FR-009): ensure the opencode CLI is working correctly%s", formatStderrSuffix(stderrBytes))
	}
	return result, nil
}
