package aireport

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/unbound-force/gaze/internal/analysis"
	"github.com/unbound-force/gaze/internal/classify"
	"github.com/unbound-force/gaze/internal/config"
	"github.com/unbound-force/gaze/internal/crap"
	"github.com/unbound-force/gaze/internal/docscan"
	"github.com/unbound-force/gaze/internal/loader"
	"github.com/unbound-force/gaze/internal/quality"
	"github.com/unbound-force/gaze/internal/report"
	"github.com/unbound-force/gaze/internal/taxonomy"
	"golang.org/x/tools/go/packages"
)

// crapStepResult holds the outputs of runCRAPStep.
type crapStepResult struct {
	JSON         json.RawMessage
	CRAPload     int
	GazeCRAPload int
}

// runCRAPStep runs the CRAP analysis pipeline and returns the JSON output
// alongside the typed CRAPload and GazeCRAPload values for threshold
// evaluation (avoiding a second JSON unmarshal in EvaluateThresholds).
func runCRAPStep(patterns []string, moduleDir string, stderr io.Writer) (*crapStepResult, error) {
	opts := crap.DefaultOptions()
	opts.Stderr = stderr

	rpt, err := crap.Analyze(patterns, moduleDir, opts)
	if err != nil {
		return nil, fmt.Errorf("CRAP analysis: %w", err)
	}

	raw, err := captureJSON(func(w io.Writer) error {
		return crap.WriteJSON(w, rpt)
	})
	if err != nil {
		return nil, err
	}

	res := &crapStepResult{
		JSON:     raw,
		CRAPload: rpt.Summary.CRAPload,
	}
	if rpt.Summary.GazeCRAPload != nil {
		res.GazeCRAPload = *rpt.Summary.GazeCRAPload
	}
	return res, nil
}

// qualityStepResult holds the outputs of runQualityStep.
type qualityStepResult struct {
	JSON                json.RawMessage
	AvgContractCoverage int
}

// runQualityStep runs the quality pipeline across all matched packages and
// returns the aggregated JSON output alongside the typed AvgContractCoverage
// value for threshold evaluation.
func runQualityStep(patterns []string, moduleDir string, stderr io.Writer) (*qualityStepResult, error) {
	pkgPaths, err := resolvePackagePaths(patterns, moduleDir)
	if err != nil {
		return nil, fmt.Errorf("resolving packages for quality: %w", err)
	}
	if len(pkgPaths) == 0 {
		return nil, fmt.Errorf("no packages matched patterns %v", patterns)
	}

	gazeConfig := loadGazeConfigBestEffort()

	var allReports []taxonomy.QualityReport
	for _, pkgPath := range pkgPaths {
		reports := runQualityForPackage(pkgPath, gazeConfig, stderr)
		allReports = append(allReports, reports...)
	}

	summary := quality.BuildPackageSummary(allReports)
	raw, err := captureJSON(func(w io.Writer) error {
		return quality.WriteJSON(w, allReports, summary)
	})
	if err != nil {
		return nil, err
	}

	avgCov := 0
	if summary != nil {
		avgCov = int(summary.AverageContractCoverage)
	}
	return &qualityStepResult{
		JSON:                raw,
		AvgContractCoverage: avgCov,
	}, nil
}

// runQualityForPackage runs the quality pipeline on a single package.
// Returns nil (not an error) if the package has no tests or analysis fails.
func runQualityForPackage(
	pkgPath string,
	gazeConfig *config.GazeConfig,
	stderr io.Writer,
) []taxonomy.QualityReport {
	analysisOpts := analysis.Options{IncludeUnexported: false}
	results, err := analysis.LoadAndAnalyze(pkgPath, analysisOpts)
	if err != nil || len(results) == 0 {
		return nil
	}

	cfg := gazeConfig
	classified, err := runClassifyResults(results, pkgPath, cfg)
	if err != nil || len(classified) == 0 {
		return nil
	}

	testPkg, err := loadTestPackageForQuality(pkgPath)
	if err != nil {
		return nil
	}

	qualOpts := quality.Options{Stderr: stderr}
	reports, _, err := quality.Assess(classified, testPkg, qualOpts)
	if err != nil {
		return nil
	}
	return reports
}

// runClassifyStep runs classification on all matched packages and returns the JSON output.
func runClassifyStep(patterns []string, moduleDir string) (json.RawMessage, error) {
	// Use the first resolved package path for analysis + classify.
	pkgPaths, err := resolvePackagePaths(patterns, moduleDir)
	if err != nil {
		return nil, fmt.Errorf("resolving packages for classification: %w", err)
	}
	if len(pkgPaths) == 0 {
		return nil, fmt.Errorf("no packages matched patterns %v", patterns)
	}

	gazeConfig := loadGazeConfigBestEffort()
	var allResults []taxonomy.AnalysisResult

	for _, pkgPath := range pkgPaths {
		analysisOpts := analysis.Options{IncludeUnexported: false}
		results, err := analysis.LoadAndAnalyze(pkgPath, analysisOpts)
		if err != nil || len(results) == 0 {
			continue
		}
		classified, err := runClassifyResults(results, pkgPath, gazeConfig)
		if err != nil {
			continue
		}
		allResults = append(allResults, classified...)
	}

	return captureJSON(func(w io.Writer) error {
		return report.WriteJSON(w, allResults, "")
	})
}

// runDocscanStep runs the documentation scanner and returns the JSON output.
func runDocscanStep(moduleDir string) (json.RawMessage, error) {
	cfg := loadGazeConfigBestEffort()
	scanOpts := docscan.ScanOptions{Config: cfg}

	docs, err := docscan.Scan(moduleDir, scanOpts)
	if err != nil {
		return nil, fmt.Errorf("docscan: %w", err)
	}
	return captureJSON(func(w io.Writer) error {
		enc := json.NewEncoder(w)
		return enc.Encode(docs)
	})
}

// resolvePackagePaths resolves package patterns to individual package paths,
// filtering out test-variant packages. Returns deduplicated package paths.
func resolvePackagePaths(patterns []string, moduleDir string) ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName,
		Dir:  moduleDir,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("resolving package patterns: %w", err)
	}

	pkgPaths := make([]string, 0, len(pkgs))
	seen := make(map[string]bool)
	for _, pkg := range pkgs {
		if pkg.PkgPath == "" || seen[pkg.PkgPath] || strings.HasSuffix(pkg.PkgPath, "_test") {
			continue
		}
		seen[pkg.PkgPath] = true
		pkgPaths = append(pkgPaths, pkg.PkgPath)
	}
	return pkgPaths, nil
}

// runClassifyResults runs the mechanical classification pipeline.
func runClassifyResults(
	results []taxonomy.AnalysisResult,
	pkgPath string,
	cfg *config.GazeConfig,
) ([]taxonomy.AnalysisResult, error) {
	targetResult, err := loader.Load(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("loading target package for classification: %w", err)
	}

	cwd, err := os.Getwd()
	var modPkgs []*packages.Package
	if err == nil {
		modResult, modErr := loader.LoadModule(cwd)
		if modErr == nil {
			modPkgs = modResult.Packages
		}
	}

	clOpts := classify.Options{
		Config:         cfg,
		ModulePackages: modPkgs,
		TargetPkg:      targetResult.Pkg,
	}
	return classify.Classify(results, clOpts), nil
}

// loadGazeConfigBestEffort loads the GazeConfig from cwd, falling back to
// the default config on any error.
func loadGazeConfigBestEffort() *config.GazeConfig {
	cwd, err := os.Getwd()
	if err != nil {
		return config.DefaultConfig()
	}
	cfgPath := filepath.Join(cwd, ".gaze.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return config.DefaultConfig()
	}
	return cfg
}

// loadTestPackageForQuality loads a Go package with test files included.
func loadTestPackageForQuality(pkgPath string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
		Tests: true,
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return nil, fmt.Errorf("loading test package: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found for %q", pkgPath)
	}
	for _, pkg := range pkgs {
		if quality.HasTestSyntax(pkg) {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("no test package found for %q", pkgPath)
}
