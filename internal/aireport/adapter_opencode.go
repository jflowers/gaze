package aireport

import (
	"bytes"
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
	// Resolve opencode path (defense-in-depth: ValidateBinary should have run first).
	opencodePath, err := exec.LookPath("opencode")
	if err != nil {
		return "", fmt.Errorf("opencode not found on PATH (FR-007): %w", err)
	}

	// Create temp directory. os.MkdirTemp sets mode 0700 on the directory.
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
	// The frontmatter ensures opencode recognizes the file as a valid agent
	// definition regardless of parser version. Use 0600 to restrict access
	// to the owner only — the system prompt may contain proprietary instructions.
	agentFile := filepath.Join(agentsDir, "gaze-reporter.md")
	content := "---\n---\n" + systemPrompt
	if err := os.WriteFile(agentFile, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("writing agent file: %w", err)
	}

	// Build args: run (subcommand), --dir <tmpDir>, --agent gaze-reporter,
	// --format default (plain-text stdout, not NDJSON event stream),
	// [-m <model>] (before the positional arg to ensure flag parsing works),
	// "" (empty positional message arg — headless trigger, mirrors -p "" for
	// claude/gemini).
	args := []string{"run", "--dir", tmpDir, "--agent", "gaze-reporter", "--format", "default"}
	if a.config.Model != "" {
		args = append(args, "-m", a.config.Model)
	}
	args = append(args, "")

	cmd := exec.CommandContext(ctx, opencodePath, args...)
	cmd.Stdin = payload

	// Capture stdout with a bounded pipe to prevent OOM on large outputs.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("creating stdout pipe for opencode: %w", err)
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("starting opencode: %w", err)
	}

	outBytes, readErr := io.ReadAll(io.LimitReader(stdoutPipe, maxAdapterOutputBytes))
	waitErr := cmd.Wait()

	if waitErr != nil {
		// Truncate stderr to avoid leaking secrets.
		stderrSnippet := stderrBuf.String()
		if len(stderrSnippet) > maxAdapterStderrBytes {
			stderrSnippet = stderrSnippet[:maxAdapterStderrBytes] + "... (truncated)"
		}
		return "", fmt.Errorf("opencode exited with error: %w\nstderr: %s", waitErr, stderrSnippet)
	}
	if readErr != nil {
		return "", fmt.Errorf("reading opencode output: %w", readErr)
	}

	result := string(outBytes)
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("opencode returned empty output (FR-009): ensure the opencode CLI is working correctly")
	}
	return result, nil
}
