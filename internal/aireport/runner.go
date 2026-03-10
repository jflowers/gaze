package aireport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// RunnerOptions configures the report pipeline runner.
type RunnerOptions struct {
	// Patterns is the package pattern(s) to analyze (e.g., "./...").
	Patterns []string

	// ModuleDir is the root of the Go module being analyzed.
	ModuleDir string

	// Adapter is the AI adapter to use for formatting.
	// Required when Format is "text"; ignored when Format is "json".
	Adapter AIAdapter

	// AdapterCfg holds adapter-specific configuration (timeout, model, etc.).
	AdapterCfg AdapterConfig

	// SystemPrompt is the formatting instructions for the AI adapter.
	// Loaded from .opencode/agents/gaze-reporter.md (stripped of YAML
	// frontmatter) or the embedded default prompt.
	SystemPrompt string

	// Format is "text" (default) or "json".
	Format string

	// Stdout receives the formatted report (text mode) or combined JSON
	// payload (json mode).
	Stdout io.Writer

	// Stderr receives progress signals, threshold summaries, and warnings.
	Stderr io.Writer

	// Thresholds holds the CI gate configuration. EvaluateThresholds is
	// called after the pipeline completes. Results are printed to Stderr.
	// When any threshold fails, Run returns a non-nil error.
	Thresholds ThresholdConfig

	// StepSummaryPath is the value of $GITHUB_STEP_SUMMARY, if set.
	// Empty means Step Summary output is disabled.
	StepSummaryPath string

	// AnalyzeFunc overrides the analysis pipeline for testing.
	// When nil, the production pipeline is called.
	AnalyzeFunc func(patterns []string, moduleDir string) (*ReportPayload, error)
}

// Run executes the report pipeline according to opts.
//
// In --format=json mode it assembles a ReportPayload from the four analysis
// steps (CRAP, Quality, Classification, Docscan) and writes the combined
// JSON to opts.Stdout. Each step's failure is recorded in PayloadErrors and
// the remaining steps still run (partial-failure mode).
//
// In --format=text mode (default) it additionally calls opts.Adapter.Format
// with the system prompt and JSON payload to produce a markdown report, then
// writes that to opts.Stdout and optionally to $GITHUB_STEP_SUMMARY.
//
// Returns an error when:
//   - The AI adapter binary is not on PATH (FR-012, text mode only; checked before analysis).
//   - No packages match the patterns (FR-013).
//   - The AI adapter returns empty output (FR-016, text mode only).
//   - The AI adapter invocation fails (text mode only).
func Run(opts RunnerOptions) error {
	if opts.Format != "text" && opts.Format != "json" {
		opts.Format = "text"
	}

	// In text mode, Adapter is required.
	if opts.Format == "text" && opts.Adapter == nil {
		return fmt.Errorf("text format requires a non-nil Adapter")
	}

	// Pre-flight binary check (FR-012): verify the AI CLI is on PATH in text
	// mode BEFORE running the analysis pipeline (which may take minutes).
	// ValidateAdapterBinary is a no-op for adapters that do not implement
	// AdapterValidator (e.g. FakeAdapter, OllamaAdapter), so tests using
	// FakeAdapter are unaffected.
	if opts.Format == "text" {
		if err := ValidateAdapterBinary(opts.Adapter); err != nil {
			return err
		}
	}

	// Step 1: Run the analysis pipeline to produce the payload.
	var payload *ReportPayload
	var err error

	analyzeFunc := opts.AnalyzeFunc
	if analyzeFunc == nil {
		analyzeFunc = func(patterns []string, moduleDir string) (*ReportPayload, error) {
			return runProductionPipeline(patterns, moduleDir, opts.Stderr)
		}
	}

	payload, err = analyzeFunc(opts.Patterns, opts.ModuleDir)
	if err != nil {
		return err
	}

	// Step 2: --format=json: write payload to stdout, evaluate thresholds, return.
	if opts.Format == "json" {
		enc := json.NewEncoder(opts.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			return err
		}
		return evaluateAndPrintThresholds(opts.Thresholds, payload, opts.Stderr)
	}

	// Step 3: --format=text: invoke AI adapter.
	_, _ = fmt.Fprintln(opts.Stderr, "Formatting report...")

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshalling payload for AI adapter: %w", err)
	}

	timeout := opts.AdapterCfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	formatted, err := opts.Adapter.Format(ctx, opts.SystemPrompt, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("AI formatting failed: %w", err)
	}
	if strings.TrimSpace(formatted) == "" {
		return fmt.Errorf("AI adapter returned empty output (FR-016): ensure the adapter is configured and working correctly")
	}

	// Step 4: Write formatted report to stdout.
	if _, err := fmt.Fprint(opts.Stdout, formatted); err != nil {
		return fmt.Errorf("writing report to stdout: %w", err)
	}

	// Step 5: Write to GITHUB_STEP_SUMMARY if set.
	if opts.StepSummaryPath != "" {
		_, _ = fmt.Fprintln(opts.Stderr, "Writing Step Summary...")
		WriteStepSummary(opts.StepSummaryPath, formatted, opts.Stderr)
	}

	return evaluateAndPrintThresholds(opts.Thresholds, payload, opts.Stderr)
}

// evaluateAndPrintThresholds evaluates cfg thresholds against payload and
// prints results to stderr. Returns a non-nil error if any threshold failed.
// When cfg has no thresholds set, it is a no-op.
func evaluateAndPrintThresholds(cfg ThresholdConfig, payload *ReportPayload, stderr io.Writer) error {
	results, allPassed := EvaluateThresholds(cfg, payload)
	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		_, _ = fmt.Fprintf(stderr, "%s: %d/%d (%s)\n", r.Name, r.Actual, r.Limit, status)
	}
	if !allPassed {
		return fmt.Errorf("one or more quality thresholds failed")
	}
	return nil
}

// errString converts an error to a *string for PayloadErrors.
func errString(err error) *string {
	if err == nil {
		return nil
	}
	s := err.Error()
	return &s
}

// runProductionPipeline runs the four-step analysis pipeline and returns
// a ReportPayload. Each step failure is recorded as a non-nil PayloadErrors
// field; remaining steps still run.
func runProductionPipeline(patterns []string, moduleDir string, stderr io.Writer) (*ReportPayload, error) {
	payload := &ReportPayload{}

	// Validate patterns are non-empty.
	if len(patterns) == 0 {
		return nil, fmt.Errorf("no package patterns specified")
	}

	// Step 1: CRAP analysis.
	_, _ = fmt.Fprintln(stderr, "Analyzing packages... (CRAP)")
	if crapRes, err := runCRAPStep(patterns, moduleDir, stderr); err != nil {
		payload.Errors.CRAP = errString(err)
	} else {
		payload.CRAP = crapRes.JSON
		payload.Summary.CRAPload = crapRes.CRAPload
		payload.Summary.GazeCRAPload = crapRes.GazeCRAPload
	}

	// Step 2: Quality analysis.
	_, _ = fmt.Fprintln(stderr, "Analyzing packages... (Quality)")
	if qualRes, err := runQualityStep(patterns, moduleDir, stderr); err != nil {
		payload.Errors.Quality = errString(err)
	} else {
		payload.Quality = qualRes.JSON
		payload.Summary.AvgContractCoverage = qualRes.AvgContractCoverage
	}

	// Step 3: Classification analysis.
	_, _ = fmt.Fprintln(stderr, "Analyzing packages... (Classification)")
	if classifyJSON, err := runClassifyStep(patterns, moduleDir); err != nil {
		payload.Errors.Classify = errString(err)
	} else {
		payload.Classify = classifyJSON
	}

	// Step 4: Documentation scan.
	_, _ = fmt.Fprintln(stderr, "Scanning documentation...")
	if docscanJSON, err := runDocscanStep(moduleDir); err != nil {
		payload.Errors.Docscan = errString(err)
	} else {
		payload.Docscan = docscanJSON
	}

	return payload, nil
}

// captureJSON runs fn writing JSON to a buffer and returns the bytes.
func captureJSON(fn func(w io.Writer) error) (json.RawMessage, error) {
	var buf bytes.Buffer
	if err := fn(&buf); err != nil {
		return nil, err
	}
	return json.RawMessage(buf.Bytes()), nil
}
