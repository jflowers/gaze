package diff

import (
	"reflect"
	"testing"
)

func TestParse_SingleFileSingleHunk(t *testing.T) {
	input := `diff --git a/pkg/foo.go b/pkg/foo.go
index abc123..def456 100644
--- a/pkg/foo.go
+++ b/pkg/foo.go
@@ -10,6 +10,6 @@ func Foo() {
`
	result := Parse(input)
	if len(result) != 1 {
		t.Fatalf("got %d file changes, want 1", len(result))
	}
	if result[0].Path != "pkg/foo.go" {
		t.Errorf("path = %q, want %q", result[0].Path, "pkg/foo.go")
	}
	if len(result[0].Ranges) != 1 {
		t.Fatalf("got %d ranges, want 1", len(result[0].Ranges))
	}
	want := LineRange{Start: 10, End: 15}
	if result[0].Ranges[0] != want {
		t.Errorf("range = %+v, want %+v", result[0].Ranges[0], want)
	}
}

func TestParse_MultipleFilesMultipleHunks(t *testing.T) {
	input := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1,3 +1,3 @@
@@ -10,2 +10,2 @@
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -5,4 +5,4 @@
@@ -20,1 +20,1 @@
`
	result := Parse(input)
	if len(result) != 2 {
		t.Fatalf("got %d file changes, want 2", len(result))
	}
	if result[0].Path != "a.go" {
		t.Errorf("file[0].Path = %q, want %q", result[0].Path, "a.go")
	}
	if len(result[0].Ranges) != 2 {
		t.Fatalf("file[0] got %d ranges, want 2", len(result[0].Ranges))
	}
	if result[1].Path != "b.go" {
		t.Errorf("file[1].Path = %q, want %q", result[1].Path, "b.go")
	}
	if len(result[1].Ranges) != 2 {
		t.Fatalf("file[1] got %d ranges, want 2", len(result[1].Ranges))
	}
}

func TestParse_EmptyDiff(t *testing.T) {
	result := Parse("")
	if len(result) != 0 {
		t.Errorf("got %d file changes, want 0", len(result))
	}
}

func TestParse_BinaryFile(t *testing.T) {
	input := `diff --git a/image.png b/image.png
index abc..def 100644
Binary files a/image.png and b/image.png differ
`
	result := Parse(input)
	if len(result) != 1 {
		t.Fatalf("got %d file changes, want 1", len(result))
	}
	if result[0].Path != "image.png" {
		t.Errorf("path = %q, want %q", result[0].Path, "image.png")
	}
	if len(result[0].Ranges) != 0 {
		t.Errorf("got %d ranges, want 0", len(result[0].Ranges))
	}
}

func TestParse_AddedFile(t *testing.T) {
	input := `diff --git a/new.go b/new.go
new file mode 100644
--- /dev/null
+++ b/new.go
@@ -0,0 +1,5 @@
+package new
+
+func Hello() {
+}
`
	result := Parse(input)
	if len(result) != 1 {
		t.Fatalf("got %d file changes, want 1", len(result))
	}
	if result[0].Path != "new.go" {
		t.Errorf("path = %q, want %q", result[0].Path, "new.go")
	}
	if len(result[0].Ranges) != 1 {
		t.Fatalf("got %d ranges, want 1", len(result[0].Ranges))
	}
	want := LineRange{Start: 1, End: 5}
	if !reflect.DeepEqual(result[0].Ranges[0], want) {
		t.Errorf("range = %+v, want %+v", result[0].Ranges[0], want)
	}
}

func TestParse_DeletedFile(t *testing.T) {
	input := `diff --git a/old.go b/old.go
deleted file mode 100644
--- a/old.go
+++ /dev/null
@@ -1,5 +0,0 @@
-package old
-
-func Goodbye() {
-}
`
	result := Parse(input)
	if len(result) != 1 {
		t.Fatalf("got %d file changes, want 1", len(result))
	}
	if result[0].Path != "old.go" {
		t.Errorf("path = %q, want %q", result[0].Path, "old.go")
	}
	if len(result[0].Ranges) != 0 {
		t.Errorf("got %d ranges, want 0 (deletion has +0 lines)", len(result[0].Ranges))
	}
}
