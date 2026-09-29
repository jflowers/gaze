package docsignal

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSidecarFile(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

const validYAML = `version: 1
contracts:
  - package: "mypackage"
    function: "process_items"
    side_effect_type: "ContainerMutation"
    label: contractual
  - package: "mypackage"
    function: "process_items"
    side_effect_type: "StreamOutput"
    label: incidental
`

func TestLoadSidecar_YAML(t *testing.T) {
	root := t.TempDir()
	writeSidecarFile(t, root, ".uf/gaze/contracts.yaml", validYAML)

	var stderr bytes.Buffer
	signals, err := LoadSidecar(root, &stderr)
	if err != nil {
		t.Fatalf("LoadSidecar: %v", err)
	}
	if len(signals) != 2 {
		t.Fatalf("got %d signals, want 2", len(signals))
	}

	// contractual ContainerMutation.
	s0 := signals[0]
	if s0.Package != "mypackage" || s0.Function != "process_items" || s0.SideEffectType != "ContainerMutation" {
		t.Errorf("signal[0] tuple = (%q, %q, %q), want (mypackage, process_items, ContainerMutation)",
			s0.Package, s0.Function, s0.SideEffectType)
	}
	if s0.Signal.Source != SourceSidecar {
		t.Errorf("signal[0].Source = %q, want %q", s0.Signal.Source, SourceSidecar)
	}
	if s0.Signal.SourceFile != ".uf/gaze/contracts.yaml" {
		t.Errorf("signal[0].SourceFile = %q, want %q", s0.Signal.SourceFile, ".uf/gaze/contracts.yaml")
	}
	if s0.Signal.Weight != 30 {
		t.Errorf("signal[0].Weight = %d, want 30", s0.Signal.Weight)
	}
	if s0.Signal.Reasoning != "sidecar declares ContainerMutation as contractual" {
		t.Errorf("signal[0].Reasoning = %q", s0.Signal.Reasoning)
	}

	// incidental StreamOutput.
	s1 := signals[1]
	if s1.SideEffectType != "StreamOutput" {
		t.Errorf("signal[1].SideEffectType = %q, want %q", s1.SideEffectType, "StreamOutput")
	}
	if s1.Signal.Weight != -30 {
		t.Errorf("signal[1].Weight = %d, want -30", s1.Signal.Weight)
	}
}

func TestLoadSidecar_JSON(t *testing.T) {
	root := t.TempDir()
	writeSidecarFile(t, root, ".uf/gaze/contracts.json", `{
  "version": 1,
  "contracts": [
    {"package": "mypackage", "function": "run", "side_effect_type": "ErrorReturn", "label": "contractual"}
  ]
}`)

	signals, err := LoadSidecar(root, nil)
	if err != nil {
		t.Fatalf("LoadSidecar: %v", err)
	}
	if len(signals) != 1 {
		t.Fatalf("got %d signals, want 1", len(signals))
	}
	s := signals[0]
	if s.Package != "mypackage" || s.Function != "run" || s.SideEffectType != "ErrorReturn" {
		t.Errorf("tuple = (%q, %q, %q), want (mypackage, run, ErrorReturn)", s.Package, s.Function, s.SideEffectType)
	}
	if s.Signal.SourceFile != ".uf/gaze/contracts.json" {
		t.Errorf("SourceFile = %q, want %q", s.Signal.SourceFile, ".uf/gaze/contracts.json")
	}
	if s.Signal.Weight != 30 {
		t.Errorf("Weight = %d, want 30", s.Signal.Weight)
	}
}

func TestLoadSidecar_MissingFile(t *testing.T) {
	root := t.TempDir() // no sidecar file written

	signals, err := LoadSidecar(root, nil)
	if err != nil {
		t.Fatalf("LoadSidecar: %v (missing file must not error)", err)
	}
	if len(signals) != 0 {
		t.Fatalf("got %d signals, want 0", len(signals))
	}
}

func TestLoadSidecar_InvalidFile(t *testing.T) {
	root := t.TempDir()
	writeSidecarFile(t, root, ".uf/gaze/contracts.yaml", "not: [valid: yaml\n  - broken")

	var stderr bytes.Buffer
	signals, err := LoadSidecar(root, &stderr)
	if err == nil {
		t.Fatal("expected error for malformed sidecar, got nil")
	}
	if signals != nil {
		t.Errorf("expected nil signals on malformed sidecar, got %d", len(signals))
	}
	if !strings.Contains(stderr.String(), "warning:") {
		t.Errorf("expected warning on stderr, got %q", stderr.String())
	}
}

func TestLoadSidecar_SemanticValidationSkip(t *testing.T) {
	root := t.TempDir()
	// Entries: unknown label, empty package, empty function, unknown type,
	// and one valid entry.
	writeSidecarFile(t, root, ".uf/gaze/contracts.yaml", `version: 1
contracts:
  - package: "a"
    function: "f"
    side_effect_type: "ContainerMutation"
    label: unknown
  - package: ""
    function: "f"
    side_effect_type: "ContainerMutation"
    label: contractual
  - package: "a"
    function: ""
    side_effect_type: "ContainerMutation"
    label: contractual
  - package: "a"
    function: "f"
    side_effect_type: "NotARealType"
    label: contractual
  - package: "valid"
    function: "f"
    side_effect_type: "ContainerMutation"
    label: contractual
`)

	var stderr bytes.Buffer
	signals, err := LoadSidecar(root, &stderr)
	if err != nil {
		t.Fatalf("LoadSidecar: %v", err)
	}
	if len(signals) != 1 {
		t.Fatalf("got %d signals, want 1 (invalid entries skipped)", len(signals))
	}
	if signals[0].Package != "valid" {
		t.Errorf("Package = %q, want %q", signals[0].Package, "valid")
	}

	// Verify warnings were emitted for the skipped entries.
	warn := stderr.String()
	for _, sub := range []string{"unknown label", "empty package or function", "unknown side_effect_type"} {
		if !strings.Contains(warn, sub) {
			t.Errorf("expected warning containing %q, got: %q", sub, warn)
		}
	}
}

func TestLoadSidecar_CustomReasoning(t *testing.T) {
	root := t.TempDir()
	writeSidecarFile(t, root, ".uf/gaze/contracts.yaml", `version: 1
contracts:
  - package: "mypackage"
    function: "f"
    side_effect_type: "ContainerMutation"
    label: contractual
    reasoning: "public API mutation per design doc"
`)

	signals, err := LoadSidecar(root, nil)
	if err != nil {
		t.Fatalf("LoadSidecar: %v", err)
	}
	if len(signals) != 1 {
		t.Fatalf("got %d signals, want 1", len(signals))
	}
	if signals[0].Signal.Reasoning != "public API mutation per design doc" {
		t.Errorf("Reasoning = %q, want custom reasoning", signals[0].Signal.Reasoning)
	}
}

func TestLoadSidecar_UnsupportedVersion(t *testing.T) {
	root := t.TempDir()
	writeSidecarFile(t, root, ".uf/gaze/contracts.yaml", `version: 2
contracts:
  - package: "pkg"
    function: "fn"
    side_effect_type: "ErrorReturn"
    label: contractual
`)

	var stderr bytes.Buffer
	signals, err := LoadSidecar(root, &stderr)
	if err != nil {
		t.Fatalf("LoadSidecar: %v (unsupported version must degrade gracefully, not error)", err)
	}
	if len(signals) != 0 {
		t.Fatalf("got %d signals, want 0 (unsupported version ignored)", len(signals))
	}
	if !strings.Contains(stderr.String(), "unsupported sidecar version") {
		t.Errorf("expected unsupported-version warning, got %q", stderr.String())
	}
}
