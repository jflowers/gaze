// Package aireport implements the AI-powered CI quality report pipeline
// for the gaze report subcommand. It orchestrates gaze's four analysis
// operations, assembles a combined JSON payload, and delegates formatting
// to an external AI adapter (claude, gemini, or ollama).
package aireport

import "encoding/json"

// ReportPayload is the combined analysis data passed to the AI adapter
// and written to stdout in --format=json mode.
type ReportPayload struct {
	// CRAP holds the raw JSON from gaze crap --format=json.
	// Nil when the CRAP analysis step failed.
	CRAP json.RawMessage `json:"crap"`

	// Quality holds the raw JSON from gaze quality --format=json.
	// Nil when the quality analysis step failed.
	Quality json.RawMessage `json:"quality"`

	// Classify holds the raw JSON from gaze analyze --classify --format=json.
	// Nil when the classification step failed.
	Classify json.RawMessage `json:"classify"`

	// Docscan holds the raw JSON from gaze docscan ([]docscan.DocumentFile).
	// Nil when the docscan step failed.
	Docscan json.RawMessage `json:"docscan"`

	// Errors records step-level failures. A nil pointer value means the
	// step succeeded. A non-nil pointer is the error message string.
	Errors PayloadErrors `json:"errors"`
}

// PayloadErrors records per-step failure messages for ReportPayload.
// A nil field means the step succeeded; a non-nil field is the error
// message string.
type PayloadErrors struct {
	CRAP     *string `json:"crap"`
	Quality  *string `json:"quality"`
	Classify *string `json:"classify"`
	Docscan  *string `json:"docscan"`
}

// ThresholdConfig holds the CI gate thresholds for gaze report.
// A nil field means the threshold was not provided on the command line
// and is disabled. A non-nil field with value 0 means "fail if any
// function exceeds this" — zero is a valid live threshold (not "disabled").
type ThresholdConfig struct {
	// MaxCrapload fails the command if the project CRAPload exceeds this value.
	MaxCrapload *int

	// MaxGazeCrapload fails the command if the GazeCRAPload exceeds this value.
	MaxGazeCrapload *int

	// MinContractCoverage fails the command if average contract coverage
	// (as a percentage integer) is below this value.
	MinContractCoverage *int
}

// ThresholdResult records the evaluation outcome for one threshold.
type ThresholdResult struct {
	// Name is the human-readable threshold label (e.g., "CRAPload").
	Name string

	// Actual is the measured value extracted from the ReportPayload.
	Actual int

	// Limit is the configured threshold value.
	Limit int

	// Passed is true when Actual is within the acceptable range.
	Passed bool
}
