// Command fake_opencode is a test double for the opencode CLI used in
// OpenCodeAdapter integration tests.
//
// It accepts a subset of opencode's flags used by the adapter:
//   - run (subcommand; consumed before flag parsing)
//   - --dir <path> (directory containing .opencode/agents/gaze-reporter.md)
//   - --agent <name> (agent name; ignored after dir verification)
//   - --format <value> (output format; ignored)
//   - -m / --model <name> (optional; value echoed in output)
//   - --exit-error (exits non-zero to simulate subprocess failure)
//   - --empty-output (writes nothing to stdout to test FR-009)
//
// On success it verifies that:
//  1. "run" is the first argument (before flags).
//  2. --dir points to a directory containing .opencode/agents/gaze-reporter.md.
//  3. That agent file starts with "---" (frontmatter present).
//
// Then it reads stdin (the JSON payload) and writes a fixed markdown response.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]

	// Consume the "run" subcommand before flag parsing so that the flag
	// package sees only flags and the trailing positional empty-string arg.
	if len(args) == 0 || args[0] != "run" {
		fmt.Fprintln(os.Stderr, "fake_opencode: expected 'run' as first argument")
		os.Exit(1)
	}
	args = args[1:] // strip "run"

	// Re-parse the remaining args with the flag package.
	fs := flag.NewFlagSet("opencode", flag.ContinueOnError)
	var (
		dir       = fs.String("dir", "", "directory containing .opencode/agents/")
		_         = fs.String("agent", "", "agent name (ignored)")
		_         = fs.String("format", "", "output format (ignored)")
		model     = fs.String("model", "", "model name")
		mShort    = fs.String("m", "", "model name (short)")
		exitError = fs.Bool("exit-error", false, "exit non-zero")
		emptyOut  = fs.Bool("empty-output", false, "write nothing to stdout")
	)
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "fake_opencode: flag parse error: %v\n", err)
		os.Exit(1)
	}

	if *exitError {
		fmt.Fprintln(os.Stderr, "fake_opencode: simulated error")
		os.Exit(1)
	}

	// Verify --dir is provided and agent file exists under it.
	if *dir == "" {
		fmt.Fprintln(os.Stderr, "fake_opencode: --dir is required")
		os.Exit(1)
	}
	agentFile := filepath.Join(*dir, ".opencode", "agents", "gaze-reporter.md")
	content, err := os.ReadFile(agentFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fake_opencode: agent file not found at %s: %v\n", agentFile, err)
		os.Exit(1)
	}
	// Verify frontmatter is present (file starts with "---").
	if !strings.HasPrefix(string(content), "---") {
		fmt.Fprintf(os.Stderr, "fake_opencode: agent file missing frontmatter (does not start with ---)\n")
		os.Exit(1)
	}

	// Read stdin (the JSON payload) — discard but consume to avoid broken pipe.
	_, _ = io.ReadAll(os.Stdin)

	if *emptyOut {
		// Write nothing — simulates FR-009 empty output scenario.
		return
	}

	modelName := *model
	if modelName == "" {
		modelName = *mShort
	}
	modelStr := ""
	if modelName != "" {
		modelStr = fmt.Sprintf(" (model: %s)", modelName)
	}

	fmt.Fprintf(os.Stdout,
		"# Fake OpenCode Report%s\n\n🔍 CRAP Analysis\n\n📊 Quality\n\n🧪 Classification\n\n🏥 Health\n",
		modelStr,
	)
}
