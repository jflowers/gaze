package adapter

import (
	"testing"

	"github.com/unbound-force/gaze/internal/crap"
)

func TestFilterTestFiles(t *testing.T) {
	mk := func(file string) crap.FunctionComplexity {
		return crap.FunctionComplexity{Package: "pkg", Function: "fn", File: file}
	}
	tests := []struct {
		name      string
		funcs     []crap.FunctionComplexity
		testFiles map[string]bool
		want      int
		wantFile  string
	}{
		{
			name:      "nil set disables filtering",
			funcs:     []crap.FunctionComplexity{mk("a.py"), mk("b.py")},
			testFiles: nil,
			want:      2,
		},
		{
			name:      "empty set disables filtering",
			funcs:     []crap.FunctionComplexity{mk("a.py"), mk("b.py")},
			testFiles: map[string]bool{},
			want:      2,
		},
		{
			name:      "populated set with relative paths drops matches",
			funcs:     []crap.FunctionComplexity{mk("src/a.py"), mk("tests/test_a.py")},
			testFiles: map[string]bool{"tests/test_a.py": true},
			want:      1,
			wantFile:  "src/a.py",
		},
		{
			name:      "populated set with no match keeps all",
			funcs:     []crap.FunctionComplexity{mk("src/a.py"), mk("src/b.py")},
			testFiles: map[string]bool{"tests/test_a.py": true},
			want:      2,
		},
		{
			name:      "dot-prefixed file path normalizes to match",
			funcs:     []crap.FunctionComplexity{mk("./tests/test_a.py")},
			testFiles: map[string]bool{"tests/test_a.py": true},
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterTestFiles(tt.funcs, tt.testFiles)
			if len(got) != tt.want {
				t.Errorf("filterTestFiles() = %d funcs, want %d", len(got), tt.want)
			}
			if tt.wantFile != "" && (len(got) != 1 || got[0].File != tt.wantFile) {
				t.Errorf("surviving entry = %v, want File %q", got, tt.wantFile)
			}
		})
	}
}
