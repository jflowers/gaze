// Package classify implements the contractual classification engine.
package classify

import (
	"fmt"

	"github.com/unbound-force/gaze/internal/config"
	"github.com/unbound-force/gaze/internal/taxonomy"
)

// baseConfidence is the neutral starting point for confidence
// scoring. Signals adjust from this base: positive pushes toward
// contractual, negative pushes toward incidental.
const baseConfidence = 50

// maxContradictionPenalty is the maximum penalty for contradicting
// signals (FR-007).
const maxContradictionPenalty = 20

// accumulateSignals sums signal weights from the base confidence,
// skipping zero-weight/empty-source signals. Returns the accumulated
// score and whether positive and negative signals were both present.
func accumulateSignals(signals []taxonomy.Signal) (score int, hasPositive, hasNegative bool) {
	score = baseConfidence
	for _, s := range signals {
		if s.Weight == 0 && s.Source == "" {
			continue
		}
		score += s.Weight
		if s.Weight > 0 {
			hasPositive = true
		}
		if s.Weight < 0 {
			hasNegative = true
		}
	}
	return score, hasPositive, hasNegative
}

// classifyLabel determines the classification label and reasoning
// string based on the score and configured thresholds.
func classifyLabel(score, contractualThreshold, incidentalThreshold int) (taxonomy.ClassificationLabel, string) {
	switch {
	case score >= contractualThreshold:
		return taxonomy.Contractual, fmt.Sprintf(
			"confidence %d >= %d (contractual threshold)",
			score, contractualThreshold,
		)
	case score < incidentalThreshold:
		return taxonomy.Incidental, fmt.Sprintf(
			"confidence %d < %d (incidental threshold)",
			score, incidentalThreshold,
		)
	default:
		return taxonomy.Ambiguous, fmt.Sprintf(
			"confidence %d in ambiguous range [%d, %d)",
			score, incidentalThreshold, contractualThreshold,
		)
	}
}

// ComputeScore computes the confidence score from a set of signals,
// applies contradiction detection and penalty, clamps to 0-100,
// and returns a Classification based on the configured thresholds.
func ComputeScore(signals []taxonomy.Signal, cfg *config.GazeConfig) taxonomy.Classification {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	score, hasPositive, hasNegative := accumulateSignals(signals)

	// Apply contradiction penalty if both positive and negative
	// signals exist.
	contradictionApplied := false
	if hasPositive && hasNegative {
		score -= maxContradictionPenalty
		contradictionApplied = true
		signals = append(signals, taxonomy.Signal{
			Source:    "contradiction",
			Weight:    -maxContradictionPenalty,
			Reasoning: "contradicting signals detected — positive and negative evidence both present",
		})
	}

	// Clamp to 0-100.
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	label, reasoning := classifyLabel(
		score,
		cfg.Classification.Thresholds.Contractual,
		cfg.Classification.Thresholds.Incidental,
	)

	if contradictionApplied {
		reasoning += "; contradiction penalty applied"
	}

	// Filter out empty signals from the result.
	filtered := make([]taxonomy.Signal, 0, len(signals))
	for _, s := range signals {
		if s.Source != "" {
			filtered = append(filtered, s)
		}
	}

	return taxonomy.Classification{
		Label:      label,
		Confidence: score,
		Signals:    filtered,
		Reasoning:  reasoning,
	}
}
