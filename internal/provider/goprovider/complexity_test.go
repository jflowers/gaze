package goprovider

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBuildEndLineMap verifies buildEndLineMap populates a correct
// {startLine: endLine} map for a real Go file.
func TestBuildEndLineMap(t *testing.T) {
	dir := t.TempDir()
	src := "package main\n\nfunc Foo() {\n\tx := 1\n\t_ = x\n}\n"
	path := filepath.Join(dir, "foo.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	m := buildEndLineMap([]string{path})
	byFile := m[path]
	if len(byFile) != 1 {
		t.Fatalf("lineMap entries = %d, want 1", len(byFile))
	}
	// Foo starts at line 3 and ends at line 6.
	if byFile[3] != 6 {
		t.Errorf("endLine for line 3 = %d, want 6", byFile[3])
	}
}

// TestBuildEndLineMap_StatAndParseFailures verifies the graceful-degradation
// branches: a nonexistent path and an unparseable file produce no entries.
func TestBuildEndLineMap_StatAndParseFailures(t *testing.T) {
	dir := t.TempDir()

	m := buildEndLineMap([]string{filepath.Join(dir, "nonexistent.go")})
	if len(m) != 0 {
		t.Errorf("map should be empty for nonexistent path, got %d entries", len(m))
	}

	bad := filepath.Join(dir, "bad.go")
	if err := os.WriteFile(bad, []byte("package main\nfunc Broken() {\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m = buildEndLineMap([]string{bad})
	if len(m) != 0 {
		t.Errorf("map should be empty for parse-failed file, got %d entries", len(m))
	}
}

// TestBuildEndLineMap_SkipsVendorTestdata verifies that vendor, testdata, and
// .git directories are skipped during the walk.
func TestBuildEndLineMap_SkipsVendorTestdata(t *testing.T) {
	dir := t.TempDir()
	for _, sub := range []string{"vendor", "testdata", ".git"} {
		subDir := filepath.Join(dir, sub)
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatal(err)
		}
		src := "package vendored\n\nfunc Skipped() {\n}\n"
		if err := os.WriteFile(filepath.Join(subDir, "skipped.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	m := buildEndLineMap([]string{dir})
	if len(m) != 0 {
		t.Errorf("map should be empty (all files under skipped dirs), got %d entries", len(m))
	}
}
