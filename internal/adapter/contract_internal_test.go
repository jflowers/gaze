package adapter

import (
	"testing"

	"github.com/unbound-force/gaze/internal/protocol"
	"github.com/unbound-force/gaze/internal/taxonomy"
)

// ---------------------------------------------------------------------------
// confidenceRange tests (table-driven)
// ---------------------------------------------------------------------------

func TestConfidenceRange(t *testing.T) {
	tests := []struct {
		name    string
		effects []taxonomy.SideEffect
		wantMin int
		wantMax int
		wantOK  bool
	}{
		{
			name:    "empty slice",
			effects: nil,
			wantMin: 0, wantMax: 0, wantOK: false,
		},
		{
			name: "all nil classification",
			effects: []taxonomy.SideEffect{
				{Classification: nil},
				{Classification: nil},
			},
			wantMin: 0, wantMax: 0, wantOK: false,
		},
		{
			name: "all classified",
			effects: []taxonomy.SideEffect{
				{Classification: &taxonomy.Classification{Confidence: 78}},
				{Classification: &taxonomy.Classification{Confidence: 85}},
				{Classification: &taxonomy.Classification{Confidence: 79}},
			},
			wantMin: 78, wantMax: 85, wantOK: true,
		},
		{
			name: "mixed nil and classified",
			effects: []taxonomy.SideEffect{
				{Classification: nil},
				{Classification: &taxonomy.Classification{Confidence: 60}},
				{Classification: nil},
				{Classification: &taxonomy.Classification{Confidence: 72}},
			},
			wantMin: 60, wantMax: 72, wantOK: true,
		},
		{
			name: "single effect",
			effects: []taxonomy.SideEffect{
				{Classification: &taxonomy.Classification{Confidence: 50}},
			},
			wantMin: 50, wantMax: 50, wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minC, maxC, found := confidenceRange(tt.effects)
			if found != tt.wantOK {
				t.Errorf("found = %v, want %v", found, tt.wantOK)
			}
			if minC != tt.wantMin {
				t.Errorf("minConf = %d, want %d", minC, tt.wantMin)
			}
			if maxC != tt.wantMax {
				t.Errorf("maxConf = %d, want %d", maxC, tt.wantMax)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// deriveCoverageReason tests (table-driven)
// ---------------------------------------------------------------------------

func TestDeriveCoverageReason(t *testing.T) {
	tests := []struct {
		name       string
		effects    []taxonomy.SideEffect
		cc         taxonomy.ContractCoverage
		wantReason string
		wantMin    int
		wantMax    int
	}{
		{
			name:       "no effects",
			effects:    nil,
			cc:         taxonomy.ContractCoverage{},
			wantReason: "no_effects_detected",
			wantMin:    0, wantMax: 0,
		},
		{
			name: "all nil classification",
			effects: []taxonomy.SideEffect{
				{Classification: nil},
				{Classification: nil},
			},
			cc:         taxonomy.ContractCoverage{TotalContractual: 0},
			wantReason: "all_effects_unclassified",
			wantMin:    0, wantMax: 0,
		},
		{
			name: "all ambiguous (classified but below threshold)",
			effects: []taxonomy.SideEffect{
				{Classification: &taxonomy.Classification{Confidence: 78}},
				{Classification: &taxonomy.Classification{Confidence: 79}},
			},
			cc:         taxonomy.ContractCoverage{TotalContractual: 0},
			wantReason: "all_effects_ambiguous",
			wantMin:    78, wantMax: 79,
		},
		{
			name: "with contractual effects",
			effects: []taxonomy.SideEffect{
				{Classification: &taxonomy.Classification{Confidence: 90}},
			},
			cc:         taxonomy.ContractCoverage{TotalContractual: 1},
			wantReason: "",
			wantMin:    0, wantMax: 0,
		},
		{
			name: "contractual > 0 with all nil classification",
			effects: []taxonomy.SideEffect{
				{Classification: nil},
				{Classification: nil},
			},
			cc:         taxonomy.ContractCoverage{TotalContractual: 1},
			wantReason: "",
			wantMin:    0, wantMax: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason, minC, maxC := deriveCoverageReason(tt.effects, tt.cc)
			if reason != tt.wantReason {
				t.Errorf("reason = %q, want %q", reason, tt.wantReason)
			}
			if minC != tt.wantMin {
				t.Errorf("minConf = %d, want %d", minC, tt.wantMin)
			}
			if maxC != tt.wantMax {
				t.Errorf("maxConf = %d, want %d", maxC, tt.wantMax)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// computeMappingClassificationConfidence tests (table-driven)
// ---------------------------------------------------------------------------

func TestComputeMappingClassificationConfidence(t *testing.T) {
	tests := []struct {
		name     string
		mappings []protocol.AssertionMappingData
		pkg      string
		fn       string
		want     int
	}{
		{
			name:     "nil mappings",
			mappings: nil,
			pkg:      "pkg", fn: "Foo",
			want: 0,
		},
		{
			name:     "empty slice",
			mappings: []protocol.AssertionMappingData{},
			pkg:      "pkg", fn: "Foo",
			want: 0,
		},
		{
			name: "all recognized",
			mappings: []protocol.AssertionMappingData{
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: "equality"},
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: "error_check"},
			},
			pkg: "pkg", fn: "Foo",
			want: 100,
		},
		{
			name: "none recognized",
			mappings: []protocol.AssertionMappingData{
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: ""},
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: ""},
			},
			pkg: "pkg", fn: "Foo",
			want: 0,
		},
		{
			name: "mixed recognized and unknown",
			mappings: []protocol.AssertionMappingData{
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: "equality"},
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: ""},
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: "error_check"},
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: ""},
			},
			pkg: "pkg", fn: "Foo",
			want: 50,
		},
		{
			name: "integer truncation (1/3 = 33)",
			mappings: []protocol.AssertionMappingData{
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: "equality"},
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: ""},
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: ""},
			},
			pkg: "pkg", fn: "Foo",
			want: 33,
		},
		{
			name: "filters by target function",
			mappings: []protocol.AssertionMappingData{
				{TargetPackage: "pkg", TargetFunction: "Foo", AssertionType: "equality"},
				{TargetPackage: "pkg", TargetFunction: "Bar", AssertionType: ""},
				{TargetPackage: "other", TargetFunction: "Foo", AssertionType: ""},
			},
			pkg: "pkg", fn: "Foo",
			want: 100, // only 1 mapping matches, and it's recognized
		},
		{
			name: "multiple test functions aggregated",
			mappings: []protocol.AssertionMappingData{
				{TargetPackage: "pkg", TargetFunction: "Foo", TestFunction: "TestA", AssertionType: "equality"},
				{TargetPackage: "pkg", TargetFunction: "Foo", TestFunction: "TestA", AssertionType: "equality"},
				{TargetPackage: "pkg", TargetFunction: "Foo", TestFunction: "TestB", AssertionType: ""},
			},
			pkg: "pkg", fn: "Foo",
			want: 66, // 2/3 = 66 (integer truncation)
		},
		{
			name: "no matching mappings in non-empty slice",
			mappings: []protocol.AssertionMappingData{
				{TargetPackage: "other", TargetFunction: "Bar", AssertionType: "equality"},
			},
			pkg: "pkg", fn: "Foo",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeMappingClassificationConfidence(tt.mappings, tt.pkg, tt.fn)
			if got != tt.want {
				t.Errorf("computeMappingClassificationConfidence() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// MappingClassificationConfidence method tests
// ---------------------------------------------------------------------------

func TestMappingClassificationConfidence(t *testing.T) {
	t.Run("before Build (nil map)", func(t *testing.T) {
		p := &ExternalContractCoverageProvider{}
		got := p.MappingClassificationConfidence("pkg", "Foo")
		if got != 0 {
			t.Errorf("MappingClassificationConfidence before Build = %d, want 0", got)
		}
	})

	t.Run("returns stored value", func(t *testing.T) {
		p := &ExternalContractCoverageProvider{
			classificationConfidence: map[string]int{
				"pkg/Foo": 75,
				"pkg/Bar": 100,
			},
		}
		if got := p.MappingClassificationConfidence("pkg", "Foo"); got != 75 {
			t.Errorf("MappingClassificationConfidence(pkg, Foo) = %d, want 75", got)
		}
		if got := p.MappingClassificationConfidence("pkg", "Bar"); got != 100 {
			t.Errorf("MappingClassificationConfidence(pkg, Bar) = %d, want 100", got)
		}
	})

	t.Run("unknown function returns 0", func(t *testing.T) {
		p := &ExternalContractCoverageProvider{
			classificationConfidence: map[string]int{
				"pkg/Foo": 75,
			},
		}
		got := p.MappingClassificationConfidence("pkg", "Unknown")
		if got != 0 {
			t.Errorf("MappingClassificationConfidence(pkg, Unknown) = %d, want 0", got)
		}
	})
}

// TestBuild_ResetsClassificationConfidence verifies that a degraded Build
// (test_mapping unsupported) clears any confidence cached by a prior
// successful build, so MappingClassificationConfidence never returns stale
// values on a subsequent lookup.
func TestBuild_ResetsClassificationConfidence(t *testing.T) {
	p := &ExternalContractCoverageProvider{
		caps:                     protocol.Capabilities{TestMapping: false},
		classificationConfidence: map[string]int{"pkg/Foo": 75},
	}

	lookup, _, err := p.Build(nil, "")
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}
	if lookup == nil {
		t.Fatal("Build() returned nil lookup")
	}

	if got := p.MappingClassificationConfidence("pkg", "Foo"); got != 0 {
		t.Errorf("MappingClassificationConfidence after degraded Build = %d, want 0", got)
	}
}
