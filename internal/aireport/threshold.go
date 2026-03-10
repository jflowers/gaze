package aireport

import (
	"encoding/json"
	"math"
)

// crapSummaryJSON is used to extract summary fields from the CRAP payload.
type crapSummaryJSON struct {
	Summary struct {
		CRAPload     int  `json:"crapload"`
		GazeCRAPload *int `json:"gaze_crapload"`
	} `json:"summary"`
}

// qualitySummaryJSON is used to extract summary fields from the quality payload.
type qualitySummaryJSON struct {
	QualitySummary *struct {
		AverageContractCoverage float64 `json:"average_contract_coverage"`
	} `json:"quality_summary"`
}

// EvaluateThresholds checks ThresholdConfig against ReportPayload summary
// data. It returns a slice of ThresholdResult values (one per non-nil
// threshold) and a boolean indicating whether all thresholds passed.
// Thresholds that cannot be evaluated (e.g., because the corresponding
// payload step failed) are treated as passed with a note that the actual
// value is 0.
func EvaluateThresholds(cfg ThresholdConfig, payload *ReportPayload) ([]ThresholdResult, bool) {
	if cfg.MaxCrapload == nil && cfg.MaxGazeCrapload == nil && cfg.MinContractCoverage == nil {
		return nil, true
	}

	var results []ThresholdResult
	allPassed := true

	// Extract CRAP summary if payload has CRAP data.
	var crapActual int
	var gazeCRAPActual int
	var hasGazeCRAP bool
	if payload != nil && len(payload.CRAP) > 0 {
		var cs crapSummaryJSON
		if err := json.Unmarshal(payload.CRAP, &cs); err == nil {
			crapActual = cs.Summary.CRAPload
			if cs.Summary.GazeCRAPload != nil {
				gazeCRAPActual = *cs.Summary.GazeCRAPload
				hasGazeCRAP = true
			}
		}
	}

	// Extract quality summary if payload has quality data.
	var avgCovActual int
	if payload != nil && len(payload.Quality) > 0 {
		var qs qualitySummaryJSON
		if err := json.Unmarshal(payload.Quality, &qs); err == nil && qs.QualitySummary != nil {
			avgCovActual = int(math.Round(qs.QualitySummary.AverageContractCoverage))
		}
	}

	if cfg.MaxCrapload != nil {
		limit := *cfg.MaxCrapload
		passed := crapActual <= limit
		if !passed {
			allPassed = false
		}
		results = append(results, ThresholdResult{
			Name:   "CRAPload",
			Actual: crapActual,
			Limit:  limit,
			Passed: passed,
		})
	}

	if cfg.MaxGazeCrapload != nil {
		limit := *cfg.MaxGazeCrapload
		actual := 0
		if hasGazeCRAP {
			actual = gazeCRAPActual
		}
		passed := actual <= limit
		if !passed {
			allPassed = false
		}
		results = append(results, ThresholdResult{
			Name:   "GazeCRAPload",
			Actual: actual,
			Limit:  limit,
			Passed: passed,
		})
	}

	if cfg.MinContractCoverage != nil {
		limit := *cfg.MinContractCoverage
		passed := avgCovActual >= limit
		if !passed {
			allPassed = false
		}
		results = append(results, ThresholdResult{
			Name:   "AvgContractCoverage",
			Actual: avgCovActual,
			Limit:  limit,
			Passed: passed,
		})
	}

	return results, allPassed
}
