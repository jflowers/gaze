package crap

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

// WriteJSON writes the CRAP report as formatted JSON.
func WriteJSON(w io.Writer, report *Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// WriteText writes the CRAP report as human-readable text.
func WriteText(w io.Writer, report *Report) error {
	if len(report.Scores) == 0 {
		fmt.Fprintln(w, "No functions analyzed.")
		return nil
	}

	// Sort by CRAP score descending for display.
	sorted := make([]Score, len(report.Scores))
	copy(sorted, report.Scores)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CRAP > sorted[j].CRAP
	})

	// Table header.
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "CRAP\tCOMPLEXITY\tCOVERAGE\tFUNCTION\tFILE\n")
	fmt.Fprintf(tw, "----\t----------\t--------\t--------\t----\n")

	for _, s := range sorted {
		marker := ""
		if s.CRAP >= report.Summary.CRAPThreshold {
			marker = " *"
		}

		// Shorten file path for display.
		file := shortenPath(s.File)

		fmt.Fprintf(tw, "%.1f%s\t%d\t%.1f%%\t%s\t%s:%d\n",
			s.CRAP, marker,
			s.Complexity,
			s.LineCoverage,
			s.Function,
			file, s.Line)
	}
	tw.Flush()

	// Summary.
	fmt.Fprintln(w)
	fmt.Fprintf(w, "--- Summary ---\n")
	fmt.Fprintf(w, "Functions analyzed:  %d\n", report.Summary.TotalFunctions)
	fmt.Fprintf(w, "Avg complexity:     %.1f\n", report.Summary.AvgComplexity)
	fmt.Fprintf(w, "Avg line coverage:  %.1f%%\n", report.Summary.AvgLineCoverage)
	fmt.Fprintf(w, "Avg CRAP score:     %.1f\n", report.Summary.AvgCRAP)
	fmt.Fprintf(w, "CRAP threshold:     %.0f\n", report.Summary.CRAPThreshold)
	fmt.Fprintf(w, "CRAPload:           %d", report.Summary.CRAPload)
	if report.Summary.CRAPload > 0 {
		fmt.Fprintf(w, " (functions at or above threshold)")
	}
	fmt.Fprintln(w)

	// GazeCRAP and quadrant stats (when available).
	if report.Summary.GazeCRAPload != nil {
		fmt.Fprintf(w, "GazeCRAP threshold: %.0f\n", *report.Summary.GazeCRAPThreshold)
		fmt.Fprintf(w, "GazeCRAPload:       %d", *report.Summary.GazeCRAPload)
		if *report.Summary.GazeCRAPload > 0 {
			fmt.Fprintf(w, " (functions at or above threshold)")
		}
		fmt.Fprintln(w)
	}

	// Quadrant breakdown.
	if len(report.Summary.QuadrantCounts) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "--- Quadrant Breakdown ---\n")
		for _, q := range []Quadrant{Q1Safe, Q2ComplexButTested, Q3SimpleButUnderspecified, Q4Dangerous} {
			count := report.Summary.QuadrantCounts[q]
			fmt.Fprintf(w, "  %-30s  %d\n", string(q), count)
		}
	}

	// Worst offenders.
	if len(report.Summary.WorstCRAP) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "--- Worst Offenders (top %d by CRAP) ---\n",
			len(report.Summary.WorstCRAP))
		for i, s := range report.Summary.WorstCRAP {
			fmt.Fprintf(w, "  %d. %.1f  %s  (%s:%d)\n",
				i+1, s.CRAP, s.Function,
				shortenPath(s.File), s.Line)
		}
	}

	return nil
}

// shortenPath removes common Go module path prefixes and returns
// a shorter relative-looking path.
func shortenPath(path string) string {
	// Find the last occurrence of a known directory marker.
	markers := []string{"/internal/", "/cmd/", "/pkg/"}
	for _, m := range markers {
		if idx := strings.LastIndex(path, m); idx >= 0 {
			return path[idx+1:]
		}
	}

	// Fall back to just the filename.
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return path
}
