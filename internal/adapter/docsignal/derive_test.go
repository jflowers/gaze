package docsignal

import (
	"testing"

	"github.com/unbound-force/gaze/v2/internal/docscan"
	"github.com/unbound-force/gaze/v2/internal/taxonomy"
)

// resultWith builds a synthetic AnalysisResult for a function with the
// given side effect types.
func resultWith(pkg, fn string, types ...taxonomy.SideEffectType) taxonomy.AnalysisResult {
	effects := make([]taxonomy.SideEffect, len(types))
	for i, ty := range types {
		effects[i] = taxonomy.SideEffect{Type: ty}
	}
	return taxonomy.AnalysisResult{
		Target: taxonomy.FunctionTarget{
			Package:  pkg,
			Function: fn,
		},
		SideEffects: effects,
	}
}

func TestDeriveSignals_EndToEnd(t *testing.T) {
	root := t.TempDir()
	// Sidecar overrides ContainerMutation → incidental, adds StreamOutput.
	writeSidecarFile(t, root, ".uf/gaze/contracts.yaml", `version: 1
contracts:
  - package: "mypackage"
    function: "process_items"
    side_effect_type: "ContainerMutation"
    label: incidental
  - package: "mypackage"
    function: "process_items"
    side_effect_type: "StreamOutput"
    label: incidental
`)

	docs := []docscan.DocumentFile{
		{
			Path:     "docs/design/containers.md",
			Content:  "<!-- gaze:contractual ContainerMutation -->\n<!-- gaze:incidental CallbackInvocation -->\n",
			Priority: docscan.PriorityOther,
		},
	}

	cached := []taxonomy.AnalysisResult{
		resultWith("mypackage", "process_items",
			taxonomy.ContainerMutation, taxonomy.CallbackInvocation, taxonomy.StreamOutput),
	}

	signals := DeriveSignals(docs, cached, root, nil)
	// Expected: 3 signals (CallbackInvocation grammar, ContainerMutation
	// sidecar override, StreamOutput sidecar).
	if len(signals) != 3 {
		t.Fatalf("got %d signals, want 3: %+v", len(signals), signals)
	}

	byType := make(map[string]DerivedSignal)
	for _, s := range signals {
		byType[s.SideEffectType] = s
	}

	// CallbackInvocation: grammar-derived incidental (architecture_doc, -25).
	cb := byType["CallbackInvocation"]
	if cb.Signal.Source != SourceArchitectureDoc {
		t.Errorf("CallbackInvocation.Source = %q, want %q", cb.Signal.Source, SourceArchitectureDoc)
	}
	if cb.Signal.Weight != -25 {
		t.Errorf("CallbackInvocation.Weight = %d, want -25", cb.Signal.Weight)
	}

	// ContainerMutation: sidecar overrides grammar (sidecar, -30).
	cm := byType["ContainerMutation"]
	if cm.Signal.Source != SourceSidecar {
		t.Errorf("ContainerMutation.Source = %q, want %q", cm.Signal.Source, SourceSidecar)
	}
	if cm.Signal.Weight != -30 {
		t.Errorf("ContainerMutation.Weight = %d, want -30", cm.Signal.Weight)
	}

	// StreamOutput: sidecar-extended (sidecar, -30).
	so := byType["StreamOutput"]
	if so.Signal.Source != SourceSidecar {
		t.Errorf("StreamOutput.Source = %q, want %q", so.Signal.Source, SourceSidecar)
	}
	if so.Signal.Weight != -30 {
		t.Errorf("StreamOutput.Weight = %d, want -30", so.Signal.Weight)
	}
}

func TestDeriveSignals_NoDocs(t *testing.T) {
	cached := []taxonomy.AnalysisResult{
		resultWith("pkg", "Foo", taxonomy.ContainerMutation),
	}

	signals := DeriveSignals(nil, cached, t.TempDir(), nil)
	if len(signals) != 0 {
		t.Fatalf("got %d signals, want 0", len(signals))
	}
}

func TestDeriveSignals_NoAnnotatedDocs(t *testing.T) {
	docs := []docscan.DocumentFile{
		{
			Path:     "docs/plain.md",
			Content:  "# Nothing here\n",
			Priority: docscan.PriorityOther,
		},
	}
	cached := []taxonomy.AnalysisResult{
		resultWith("pkg", "Foo", taxonomy.ContainerMutation),
	}

	signals := DeriveSignals(docs, cached, t.TempDir(), nil)
	if len(signals) != 0 {
		t.Fatalf("got %d signals, want 0", len(signals))
	}
}

func TestDeriveSignals_FanOutMultipleFunctions(t *testing.T) {
	docs := []docscan.DocumentFile{
		{
			Path:     "README.md",
			Content:  "<!-- gaze:contractual ContainerMutation -->\n",
			Priority: docscan.PriorityModuleRoot,
		},
	}
	cached := []taxonomy.AnalysisResult{
		resultWith("mypackage", "process_items", taxonomy.ContainerMutation),
		resultWith("mypackage", "list_items", taxonomy.ContainerMutation),
	}

	signals := DeriveSignals(docs, cached, t.TempDir(), nil)
	if len(signals) != 2 {
		t.Fatalf("got %d signals, want 2 (fan-out to both functions)", len(signals))
	}

	// readme source, weight +15 for both.
	for _, s := range signals {
		if s.Signal.Source != SourceReadme {
			t.Errorf("%s.Source = %q, want %q", s.Function, s.Signal.Source, SourceReadme)
		}
		if s.Signal.Weight != 15 {
			t.Errorf("%s.Weight = %d, want 15", s.Function, s.Signal.Weight)
		}
	}
}

func TestDeriveSignals_DedupMultipleEffectsSameType(t *testing.T) {
	docs := []docscan.DocumentFile{
		{
			Path:     "README.md",
			Content:  "<!-- gaze:contractual MapMutation -->\n",
			Priority: docscan.PriorityModuleRoot,
		},
	}
	// A function with two MapMutation effects of the same type — fan-out
	// must emit a single signal per tuple, not one per effect.
	cached := []taxonomy.AnalysisResult{
		resultWith("pkg", "Foo", taxonomy.MapMutation, taxonomy.MapMutation),
	}

	signals := DeriveSignals(docs, cached, t.TempDir(), nil)
	if len(signals) != 1 {
		t.Fatalf("got %d signals, want 1 (dedup same tuple)", len(signals))
	}
}

func TestDeriveSignals_DeterministicOrder(t *testing.T) {
	docs := []docscan.DocumentFile{
		{
			Path:     "README.md",
			Content:  "<!-- gaze:contractual StreamOutput -->\n<!-- gaze:contractual ContainerMutation -->\n",
			Priority: docscan.PriorityModuleRoot,
		},
	}
	cached := []taxonomy.AnalysisResult{
		resultWith("pkg", "Foo", taxonomy.ContainerMutation, taxonomy.StreamOutput),
	}

	// Run twice and verify identical ordering.
	first := DeriveSignals(docs, cached, t.TempDir(), nil)
	second := DeriveSignals(docs, cached, t.TempDir(), nil)
	if len(first) != len(second) {
		t.Fatalf("non-deterministic length: %d vs %d", len(first), len(second))
	}
	for i := range first {
		a, b := first[i], second[i]
		if a.SideEffectType != b.SideEffectType {
			t.Errorf("position %d: %q vs %q", i, a.SideEffectType, b.SideEffectType)
		}
	}
	// Sorted by SideEffectType: ContainerMutation < StreamOutput.
	if first[0].SideEffectType != "ContainerMutation" || first[1].SideEffectType != "StreamOutput" {
		t.Errorf("unexpected order: %q, %q", first[0].SideEffectType, first[1].SideEffectType)
	}
}
