package report

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/jflowers/gaze/internal/taxonomy"
)

// WriteText writes analysis results as human-readable text to the
// writer. Output is formatted as a table suitable for terminal
// display within 80 columns.
func WriteText(w io.Writer, results []taxonomy.AnalysisResult) error {
	for i, result := range results {
		if i > 0 {
			fmt.Fprintln(w)
		}
		if err := writeOneResult(w, result); err != nil {
			return err
		}
	}

	// Summary line.
	total := 0
	for _, r := range results {
		total += len(r.SideEffects)
	}
	fmt.Fprintf(w, "\n%d function(s) analyzed, %d side effect(s) detected\n",
		len(results), total)

	return nil
}

func writeOneResult(w io.Writer, result taxonomy.AnalysisResult) error {
	// Header.
	name := result.Target.QualifiedName()
	fmt.Fprintf(w, "=== %s ===\n", name)
	fmt.Fprintf(w, "    %s\n", result.Target.Signature)
	fmt.Fprintf(w, "    %s\n", result.Target.Location)

	if len(result.SideEffects) == 0 {
		fmt.Fprintln(w, "    No side effects detected.")
		return nil
	}

	fmt.Fprintln(w)

	// Side effects table.
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "    TIER\tTYPE\tDESCRIPTION\n")
	fmt.Fprintf(tw, "    ----\t----\t-----------\n")

	for _, e := range result.SideEffects {
		desc := e.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		fmt.Fprintf(tw, "    %s\t%s\t%s\n",
			e.Tier,
			e.Type,
			desc)
	}
	tw.Flush()

	// Tier summary.
	tierCounts := make(map[taxonomy.Tier]int)
	for _, e := range result.SideEffects {
		tierCounts[e.Tier]++
	}

	var parts []string
	for _, tier := range []taxonomy.Tier{
		taxonomy.TierP0, taxonomy.TierP1,
		taxonomy.TierP2, taxonomy.TierP3, taxonomy.TierP4,
	} {
		if c, ok := tierCounts[tier]; ok {
			parts = append(parts, fmt.Sprintf("%s: %d", tier, c))
		}
	}
	fmt.Fprintf(w, "\n    Summary: %s\n", strings.Join(parts, ", "))

	return nil
}
