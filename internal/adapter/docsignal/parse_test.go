package docsignal

import (
	"testing"

	"github.com/unbound-force/gaze/v2/internal/docscan"
)

func TestParseAnnotations_HTMLCommentForm(t *testing.T) {
	doc := docscan.DocumentFile{
		Path:     "docs/design/containers.md",
		Content:  "<!-- gaze:contractual ContainerMutation -->\n<!-- gaze:incidental CallbackInvocation -->\n",
		Priority: docscan.PriorityOther,
	}

	decls := parseAnnotations(doc)
	if len(decls) != 2 {
		t.Fatalf("got %d declarations, want 2", len(decls))
	}

	// Architecture doc → source "architecture_doc", weight ±25.
	want := []struct {
		typeName string
		label    string
		source   string
		weight   int
	}{
		{"ContainerMutation", "contractual", SourceArchitectureDoc, 25},
		{"CallbackInvocation", "incidental", SourceArchitectureDoc, -25},
	}
	for i, w := range want {
		d := decls[i]
		if d.typeName != w.typeName {
			t.Errorf("decl[%d].typeName = %q, want %q", i, d.typeName, w.typeName)
		}
		if d.label != w.label {
			t.Errorf("decl[%d].label = %q, want %q", i, d.label, w.label)
		}
		if d.source != w.source {
			t.Errorf("decl[%d].source = %q, want %q", i, d.source, w.source)
		}
		if d.weight != w.weight {
			t.Errorf("decl[%d].weight = %d, want %d", i, d.weight, w.weight)
		}
		if d.sourceFile != "docs/design/containers.md" {
			t.Errorf("decl[%d].sourceFile = %q, want %q", i, d.sourceFile, "docs/design/containers.md")
		}
	}
}

func TestParseAnnotations_FencedBlockForm(t *testing.T) {
	doc := docscan.DocumentFile{
		Path: "README.md",
		Content: "> [!gaze-contract]\n" +
			"> - `ContainerMutation` — contractual (public API mutation)\n" +
			"> - `CallbackInvocation` — incidental (internal wiring)\n",
		Priority: docscan.PriorityModuleRoot,
	}

	decls := parseAnnotations(doc)
	if len(decls) != 2 {
		t.Fatalf("got %d declarations, want 2", len(decls))
	}

	// README doc → source "readme", weight ±15.
	want := []struct {
		typeName string
		label    string
		source   string
		weight   int
	}{
		{"ContainerMutation", "contractual", SourceReadme, 15},
		{"CallbackInvocation", "incidental", SourceReadme, -15},
	}
	for i, w := range want {
		d := decls[i]
		if d.typeName != w.typeName {
			t.Errorf("decl[%d].typeName = %q, want %q", i, d.typeName, w.typeName)
		}
		if d.label != w.label {
			t.Errorf("decl[%d].label = %q, want %q", i, d.label, w.label)
		}
		if d.source != w.source {
			t.Errorf("decl[%d].source = %q, want %q", i, d.source, w.source)
		}
		if d.weight != w.weight {
			t.Errorf("decl[%d].weight = %d, want %d", i, d.weight, w.weight)
		}
	}
}

func TestParseAnnotations_CaseInsensitiveLabel(t *testing.T) {
	doc := docscan.DocumentFile{
		Path:     "docs/DESIGN.md",
		Content:  "<!-- GAZE:CONTRACTUAL ContainerMutation -->\n",
		Priority: docscan.PriorityOther,
	}

	decls := parseAnnotations(doc)
	if len(decls) != 1 {
		t.Fatalf("got %d declarations, want 1", len(decls))
	}
	if decls[0].label != "contractual" {
		t.Errorf("label = %q, want %q", decls[0].label, "contractual")
	}
	if decls[0].typeName != "ContainerMutation" {
		t.Errorf("typeName = %q, want %q", decls[0].typeName, "ContainerMutation")
	}
}

func TestParseAnnotations_MalformedSkipped(t *testing.T) {
	doc := docscan.DocumentFile{
		Path: "docs/design/mixed.md",
		Content: "<!-- gaze:contractual -->\n" + // missing type name
			"<!-- gaze:invalidLabel ContainerMutation -->\n" + // unknown label
			"<!-- gaze:contractual StreamOutput -->\n", // valid
		Priority: docscan.PriorityOther,
	}

	decls := parseAnnotations(doc)
	if len(decls) != 1 {
		t.Fatalf("got %d declarations, want 1 (malformed silently skipped)", len(decls))
	}
	if decls[0].typeName != "StreamOutput" {
		t.Errorf("typeName = %q, want %q", decls[0].typeName, "StreamOutput")
	}
}

func TestParseAnnotations_EmptyDocument(t *testing.T) {
	doc := docscan.DocumentFile{
		Path:     "docs/plain.md",
		Content:  "# Just a heading\n\nNo annotations here.\n",
		Priority: docscan.PriorityOther,
	}

	if decls := parseAnnotations(doc); len(decls) != 0 {
		t.Fatalf("got %d declarations, want 0", len(decls))
	}
}

func TestResolveConflicts_ModuleRootWins(t *testing.T) {
	// PriorityOther doc declares ContainerMutation contractual.
	nested := parseAnnotations(docscan.DocumentFile{
		Path:     "pkg/mypackage/README.md",
		Content:  "<!-- gaze:contractual ContainerMutation -->\n",
		Priority: docscan.PriorityOther,
	})
	// PriorityModuleRoot doc declares ContainerMutation incidental.
	root := parseAnnotations(docscan.DocumentFile{
		Path:     "README.md",
		Content:  "<!-- gaze:incidental ContainerMutation -->\n",
		Priority: docscan.PriorityModuleRoot,
	})

	all := append(append([]typeDecl{}, nested...), root...)
	resolved := resolveConflicts(all)

	if len(resolved) != 1 {
		t.Fatalf("got %d declarations, want 1", len(resolved))
	}
	d := resolved[0]
	if d.label != "incidental" {
		t.Errorf("label = %q, want %q (ModuleRoot must win)", d.label, "incidental")
	}
	if d.weight != -15 {
		t.Errorf("weight = %d, want -15", d.weight)
	}
	if d.sourceFile != "README.md" {
		t.Errorf("sourceFile = %q, want %q", d.sourceFile, "README.md")
	}
}

func TestResolveConflicts_NoConflictKeepsAll(t *testing.T) {
	decls := []typeDecl{
		{typeName: "ContainerMutation", priority: docscan.PriorityOther},
		{typeName: "StreamOutput", priority: docscan.PriorityModuleRoot},
	}
	resolved := resolveConflicts(decls)
	if len(resolved) != 2 {
		t.Fatalf("got %d declarations, want 2", len(resolved))
	}
	// Deterministic sort: ContainerMutation < StreamOutput.
	if resolved[0].typeName != "ContainerMutation" || resolved[1].typeName != "StreamOutput" {
		t.Errorf("unexpected order: %q, %q", resolved[0].typeName, resolved[1].typeName)
	}
}
