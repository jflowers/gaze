package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jflowers/gaze/internal/analysis"
	"github.com/jflowers/gaze/internal/crap"
	"github.com/jflowers/gaze/internal/report"
	"github.com/spf13/cobra"
)

// Set by build flags.
var version = "dev"

func main() {
	root := &cobra.Command{
		Use:   "gaze",
		Short: "Gaze — test quality analysis via side effect detection",
		Long: `Gaze analyzes Go functions to detect observable side effects
and measures whether unit tests assert on all contractual changes
produced by their test targets.`,
		Version: version,
	}

	root.AddCommand(newAnalyzeCmd())
	root.AddCommand(newCrapCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newAnalyzeCmd() *cobra.Command {
	var (
		function          string
		format            string
		includeUnexported bool
	)

	cmd := &cobra.Command{
		Use:   "analyze [package]",
		Short: "Analyze side effects of Go functions",
		Long: `Analyze a Go package (or specific function) and report all
observable side effects each function produces.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkgPath := args[0]

			if format != "text" && format != "json" {
				return fmt.Errorf("invalid format %q: must be 'text' or 'json'", format)
			}

			opts := analysis.Options{
				IncludeUnexported: includeUnexported,
				FunctionFilter:    function,
			}

			results, err := analysis.LoadAndAnalyze(pkgPath, opts)
			if err != nil {
				return err
			}

			if len(results) == 0 {
				if function != "" {
					return fmt.Errorf("function %q not found in package %q", function, pkgPath)
				}
				fmt.Fprintln(os.Stderr, "no functions found to analyze")
				return nil
			}

			switch format {
			case "json":
				return report.WriteJSON(os.Stdout, results)
			default:
				return report.WriteText(os.Stdout, results)
			}
		},
	}

	cmd.Flags().StringVarP(&function, "function", "f", "",
		"analyze a specific function (default: all exported)")
	cmd.Flags().StringVar(&format, "format", "text",
		"output format: text or json")
	cmd.Flags().BoolVar(&includeUnexported, "include-unexported", false,
		"include unexported functions")

	return cmd
}

func newCrapCmd() *cobra.Command {
	var (
		format            string
		coverProfile      string
		crapThreshold     float64
		gazeCrapThreshold float64
		maxCrapload       int
		maxGazeCrapload   int
	)

	cmd := &cobra.Command{
		Use:   "crap [packages...]",
		Short: "Compute CRAP scores for Go functions",
		Long: `Compute CRAP (Change Risk Anti-Patterns) scores by combining
cyclomatic complexity with test coverage. Reports per-function
CRAP scores and the project's CRAPload (count of functions above
the threshold).

If no coverage profile is provided, runs 'go test -coverprofile'
automatically.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if format != "text" && format != "json" {
				return fmt.Errorf("invalid format %q: must be 'text' or 'json'", format)
			}

			moduleDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			opts := crap.Options{
				CoverProfile:      coverProfile,
				CRAPThreshold:     crapThreshold,
				GazeCRAPThreshold: gazeCrapThreshold,
				MaxCRAPload:       maxCrapload,
				MaxGazeCRAPload:   maxGazeCrapload,
				IgnoreGenerated:   true,
			}

			rpt, err := crap.Analyze(args, moduleDir, opts)
			if err != nil {
				return err
			}

			switch format {
			case "json":
				if err := crap.WriteJSON(os.Stdout, rpt); err != nil {
					return err
				}
			default:
				if err := crap.WriteText(os.Stdout, rpt); err != nil {
					return err
				}
			}

			// CI summary line (when thresholds are set).
			if maxCrapload > 0 || maxGazeCrapload > 0 {
				var parts []string
				if maxCrapload > 0 {
					status := "PASS"
					if rpt.Summary.CRAPload > maxCrapload {
						status = "FAIL"
					}
					parts = append(parts, fmt.Sprintf("CRAPload: %d/%d (%s)",
						rpt.Summary.CRAPload, maxCrapload, status))
				}
				if maxGazeCrapload > 0 && rpt.Summary.GazeCRAPload != nil {
					status := "PASS"
					if *rpt.Summary.GazeCRAPload > maxGazeCrapload {
						status = "FAIL"
					}
					parts = append(parts, fmt.Sprintf("GazeCRAPload: %d/%d (%s)",
						*rpt.Summary.GazeCRAPload, maxGazeCrapload, status))
				}
				fmt.Fprintln(os.Stderr, strings.Join(parts, " | "))
			}

			// CI enforcement.
			if maxCrapload > 0 && rpt.Summary.CRAPload > maxCrapload {
				return fmt.Errorf("CRAPload %d exceeds maximum %d",
					rpt.Summary.CRAPload, maxCrapload)
			}
			if maxGazeCrapload > 0 && rpt.Summary.GazeCRAPload != nil &&
				*rpt.Summary.GazeCRAPload > maxGazeCrapload {
				return fmt.Errorf("GazeCRAPload %d exceeds maximum %d",
					*rpt.Summary.GazeCRAPload, maxGazeCrapload)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text",
		"output format: text or json")
	cmd.Flags().StringVar(&coverProfile, "coverprofile", "",
		"path to coverage profile (default: generate via go test)")
	cmd.Flags().Float64Var(&crapThreshold, "crap-threshold", 15,
		"CRAP score threshold for flagging functions")
	cmd.Flags().Float64Var(&gazeCrapThreshold, "gaze-crap-threshold", 15,
		"GazeCRAP score threshold (used when contract coverage available)")
	cmd.Flags().IntVar(&maxCrapload, "max-crapload", 0,
		"fail if CRAPload exceeds this (0 = no limit)")
	cmd.Flags().IntVar(&maxGazeCrapload, "max-gaze-crapload", 0,
		"fail if GazeCRAPload exceeds this (0 = no limit)")

	return cmd
}
