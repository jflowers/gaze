package main

import (
	"fmt"
	"os"

	"github.com/jflowers/gaze/internal/analysis"
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

			// Validate format flag.
			if format != "text" && format != "json" {
				return fmt.Errorf("invalid format %q: must be 'text' or 'json'", format)
			}

			// Run analysis.
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

			// Output results.
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
