// Command fake_gemini is a test double for the gemini CLI used in
// GeminiAdapter integration tests.
//
// It accepts a subset of gemini's flags used by the adapter:
//   - -p (headless trigger; value ignored)
//   - --output-format json (required by adapter; value checked)
//   - -m / --model <name> (optional; echoed in response)
//   - --exit-error (exits non-zero to simulate subprocess failure)
//   - --empty-output (writes {"response": ""} to test FR-016)
//
// It also checks that GEMINI.md exists in its working directory (the
// adapter sets cmd.Dir to a temp dir containing the file).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

type geminiResponse struct {
	Response string `json:"response"`
}

func main() {
	var (
		_         = flag.String("p", "", "headless prompt (ignored)")
		_         = flag.String("output-format", "", "output format (expected: json)")
		model     = flag.String("model", "", "model name")
		mShort    = flag.String("m", "", "model name (short)")
		exitError = flag.Bool("exit-error", false, "exit non-zero")
		emptyOut  = flag.Bool("empty-output", false, "write empty response")
	)
	flag.Parse()

	if *exitError {
		fmt.Fprintln(os.Stderr, "fake_gemini: simulated error")
		os.Exit(1)
	}

	// Verify GEMINI.md exists in current working directory.
	if _, err := os.Stat("GEMINI.md"); err != nil {
		fmt.Fprintf(os.Stderr, "fake_gemini: GEMINI.md not found in working directory: %v\n", err)
		os.Exit(1)
	}

	// Read stdin (the JSON payload).
	_, _ = io.ReadAll(os.Stdin)

	if *emptyOut {
		enc := json.NewEncoder(os.Stdout)
		_ = enc.Encode(geminiResponse{Response: ""})
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

	resp := geminiResponse{
		Response: fmt.Sprintf("# Fake Gemini Report%s\n\n🔍 CRAP Analysis\n\n📊 Quality\n\n🧪 Classification\n\n🏥 Health\n", modelStr),
	}
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(resp)
}
