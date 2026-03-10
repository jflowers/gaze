// Command fake_claude is a test double for the claude CLI used in
// ClaudeAdapter integration tests.
//
// It accepts a subset of claude's flags used by the adapter:
//   - -p (headless mode trigger; value ignored)
//   - --system-prompt-file <path> (system prompt file; contents verified to exist)
//   - --model <name> (optional; value echoed in output)
//   - --exit-error (exits non-zero to simulate subprocess failure)
//   - --empty-output (writes nothing to stdout to test FR-016)
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	var (
		_          = flag.String("p", "", "headless prompt (ignored)")
		promptFile = flag.String("system-prompt-file", "", "system prompt file path")
		model      = flag.String("model", "", "model name")
		exitError  = flag.Bool("exit-error", false, "exit non-zero")
		emptyOut   = flag.Bool("empty-output", false, "write nothing to stdout")
	)
	flag.Parse()

	if *exitError {
		fmt.Fprintln(os.Stderr, "fake_claude: simulated error")
		os.Exit(1)
	}

	if *promptFile != "" {
		if _, err := os.Stat(*promptFile); err != nil {
			fmt.Fprintf(os.Stderr, "fake_claude: system-prompt-file not found: %v\n", err)
			os.Exit(1)
		}
	}

	// Read stdin (the JSON payload).
	_, _ = io.ReadAll(os.Stdin)

	if *emptyOut {
		// Write nothing — simulates FR-016 empty output scenario.
		return
	}

	modelStr := ""
	if *model != "" {
		modelStr = fmt.Sprintf(" (model: %s)", *model)
	}
	fmt.Fprintf(os.Stdout, "# Fake Claude Report%s\n\n🔍 CRAP Analysis\n\n📊 Quality\n\n🧪 Classification\n\n🏥 Health\n", modelStr)
}
