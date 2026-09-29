package docsignal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/unbound-force/gaze/v2/internal/taxonomy"
)

// sidecarFile is the top-level structure of a sidecar contracts file
// (YAML or JSON). The two formats are structurally equivalent.
type sidecarFile struct {
	Version   int            `yaml:"version" json:"version"`
	Contracts []sidecarEntry `yaml:"contracts" json:"contracts"`
}

// sidecarEntry is a single (package, function, side_effect_type) → label
// declaration. The function key is the analyzer's verbatim reported
// function name (bare name, NOT receiver-qualified).
type sidecarEntry struct {
	Package        string `yaml:"package" json:"package"`
	Function       string `yaml:"function" json:"function"`
	SideEffectType string `yaml:"side_effect_type" json:"side_effect_type"`
	Label          string `yaml:"label" json:"label"`
	Reasoning      string `yaml:"reasoning" json:"reasoning"`
}

// LoadSidecar loads the sidecar contracts file at
// .uf/gaze/contracts.yaml (primary) or .uf/gaze/contracts.json
// (alternative), resolved relative to moduleRoot.
//
// Graceful degradation (design D5):
//   - When neither file exists, it returns (nil, nil) — no error, no
//     warning.
//   - When a file exists but is malformed, it writes a warning to
//     stderr and returns (nil, err); callers degrade to grammar-derived
//     signals only.
//   - Semantically invalid entries (unknown label, empty package or
//     function, or unknown side_effect_type) are skipped with a warning
//     to stderr; remaining valid entries are still loaded.
func LoadSidecar(moduleRoot string, stderr io.Writer) ([]DerivedSignal, error) {
	// The source file is recorded relative to the module root, matching
	// the docscan convention of storing paths relative to the repo root.
	relYAML := filepath.Join(sidecarDir, sidecarFileName+".yaml")
	relJSON := filepath.Join(sidecarDir, sidecarFileName+".json")

	yamlPath := filepath.Join(moduleRoot, relYAML)
	jsonPath := filepath.Join(moduleRoot, relJSON)

	if data, err := os.ReadFile(yamlPath); err == nil {
		return parseSidecar(data, relYAML, true, stderr)
	} else if !errors.Is(err, fs.ErrNotExist) {
		warnf(stderr, "warning: reading sidecar %s: %v\n", yamlPath, err)
		return nil, err
	}

	if data, err := os.ReadFile(jsonPath); err == nil {
		return parseSidecar(data, relJSON, false, stderr)
	} else if !errors.Is(err, fs.ErrNotExist) {
		warnf(stderr, "warning: reading sidecar %s: %v\n", jsonPath, err)
		return nil, err
	}

	// Neither file exists — proceed with grammar-derived signals only.
	return nil, nil
}

// parseSidecar decodes the raw sidecar file content (YAML when isYAML
// is true, JSON otherwise) and builds DerivedSignals with semantic
// validation. The sourceFile is the module-root-relative path (e.g.
// ".uf/gaze/contracts.yaml").
func parseSidecar(data []byte, sourceFile string, isYAML bool, stderr io.Writer) ([]DerivedSignal, error) {
	var file sidecarFile
	var err error
	if isYAML {
		err = yaml.Unmarshal(data, &file)
	} else {
		err = json.Unmarshal(data, &file)
	}
	if err != nil {
		warnf(stderr, "warning: parsing sidecar %s: %v\n", sourceFile, err)
		return nil, fmt.Errorf("parsing sidecar %s: %w", sourceFile, err)
	}

	// Validate the schema version. Version 0 (field omitted) and 1 are
	// accepted; a higher value signals a future schema this binary cannot
	// parse safely. Degrade gracefully rather than misinterpreting fields.
	if file.Version > 1 {
		warnf(stderr, "warning: unsupported sidecar version %d in %s, ignoring\n", file.Version, sourceFile)
		return nil, nil
	}

	return buildSidecarSignals(file.Contracts, filepath.ToSlash(sourceFile), stderr), nil
}

// buildSidecarSignals converts sidecar entries into DerivedSignals,
// skipping semantically invalid entries with a warning to stderr.
func buildSidecarSignals(entries []sidecarEntry, sourceFile string, stderr io.Writer) []DerivedSignal {
	signals := make([]DerivedSignal, 0, len(entries))
	for _, e := range entries {
		label := strings.ToLower(e.Label)
		if label != string(taxonomy.Contractual) && label != string(taxonomy.Incidental) {
			warnf(stderr, "warning: skipping sidecar entry: unknown label %q\n", e.Label)
			continue
		}
		if e.Package == "" || e.Function == "" {
			warnf(stderr, "warning: skipping sidecar entry: empty package or function\n")
			continue
		}
		if !taxonomy.IsKnownType(taxonomy.SideEffectType(e.SideEffectType)) {
			warnf(stderr, "warning: skipping sidecar entry: unknown side_effect_type %q\n", e.SideEffectType)
			continue
		}

		weight := SidecarWeight
		if label == string(taxonomy.Incidental) {
			weight = -SidecarWeight
		}

		reasoning := e.Reasoning
		if reasoning == "" {
			reasoning = fmt.Sprintf("sidecar declares %s as %s", e.SideEffectType, label)
		}

		signals = append(signals, DerivedSignal{
			Package:        e.Package,
			Function:       e.Function,
			SideEffectType: e.SideEffectType,
			Signal: taxonomy.Signal{
				Source:     SourceSidecar,
				SourceFile: sourceFile,
				Weight:     weight,
				Reasoning:  reasoning,
			},
		})
	}
	return signals
}

// warnf writes a warning to stderr, no-op when stderr is nil.
func warnf(stderr io.Writer, format string, args ...any) {
	if stderr != nil {
		_, _ = fmt.Fprintf(stderr, format, args...)
	}
}
