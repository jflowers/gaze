package cognitive

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func cc(t *testing.T, src string) int {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	for _, decl := range file.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			return AnalyzeFunc(fset, fd)
		}
	}
	t.Fatal("no function declaration found")
	return -1
}

func TestSimpleIf(t *testing.T) {
	got := cc(t, `package p
func f(x bool) {
	if x {}
}`)
	if got != 1 {
		t.Errorf("simple if: got %d, want 1", got)
	}
}

func TestNestedIf(t *testing.T) {
	got := cc(t, `package p
func f(x, y bool) {
	if x {
		if y {}
	}
}`)
	if got != 3 {
		t.Errorf("nested if: got %d, want 3", got)
	}
}

func TestElseIfChain(t *testing.T) {
	got := cc(t, `package p
func f(a, b, c bool) {
	if a {
	} else if b {
	} else if c {
	} else {
	}
}`)
	if got != 4 {
		t.Errorf("else if chain: got %d, want 4", got)
	}
}

func TestForLoop(t *testing.T) {
	got := cc(t, `package p
func f() {
	for i := 0; i < 10; i++ {}
}`)
	if got != 1 {
		t.Errorf("for loop: got %d, want 1", got)
	}
}

func TestRangeLoop(t *testing.T) {
	got := cc(t, `package p
func f(s []int) {
	for range s {}
}`)
	if got != 1 {
		t.Errorf("range loop: got %d, want 1", got)
	}
}

func TestSwitch(t *testing.T) {
	got := cc(t, `package p
func f(x int) {
	switch x {
	case 1:
	case 2:
	}
}`)
	if got != 1 {
		t.Errorf("switch: got %d, want 1", got)
	}
}

func TestTypeSwitch(t *testing.T) {
	got := cc(t, `package p
func f(x interface{}) {
	switch x.(type) {
	case int:
	case string:
	}
}`)
	if got != 1 {
		t.Errorf("type switch: got %d, want 1", got)
	}
}

func TestSelect(t *testing.T) {
	got := cc(t, `package p
func f(ch chan int) {
	select {
	case <-ch:
	}
}`)
	if got != 1 {
		t.Errorf("select: got %d, want 1", got)
	}
}

func TestLogicalOperatorSameSequence(t *testing.T) {
	got := cc(t, `package p
func f(a, b, c bool) {
	_ = a && b && c
}`)
	if got != 1 {
		t.Errorf("same logical operator sequence: got %d, want 1", got)
	}
}

func TestMixedLogicalOperators(t *testing.T) {
	got := cc(t, `package p
func f(a, b, c bool) {
	_ = a && b || c
}`)
	if got != 2 {
		t.Errorf("mixed logical operators: got %d, want 2", got)
	}
}

func TestGoto(t *testing.T) {
	got := cc(t, `package p
func f() {
	goto end
	end:
}`)
	if got != 1 {
		t.Errorf("goto: got %d, want 1", got)
	}
}

func TestRecursion(t *testing.T) {
	got := cc(t, `package p
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}`)
	if got != 2 {
		t.Errorf("recursion: got %d, want 2", got)
	}
}

func TestDeeplyNested(t *testing.T) {
	got := cc(t, `package p
func f(x int, s []int) {
	for range s {
		if x > 0 {
			switch x {
			case 1:
			}
		}
	}
}`)
	if got != 6 {
		t.Errorf("deeply nested: got %d, want 6", got)
	}
}

func TestEmptyFunction(t *testing.T) {
	got := cc(t, `package p
func f() {}`)
	if got != 0 {
		t.Errorf("empty function: got %d, want 0", got)
	}
}

func TestNilBody(t *testing.T) {
	fset := token.NewFileSet()
	fd := &ast.FuncDecl{
		Name: ast.NewIdent("f"),
	}
	got := AnalyzeFunc(fset, fd)
	if got != 0 {
		t.Errorf("nil body: got %d, want 0", got)
	}
}

func TestAnalyzeFile(t *testing.T) {
	src := `package p
func a() {
	if true {}
}
func b() {
	for i := 0; i < 10; i++ {}
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	results := AnalyzeFile(fset, file)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Function != "a" || results[0].CognitiveComplexity != 1 {
		t.Errorf("func a: got %+v", results[0])
	}
	if results[1].Function != "b" || results[1].CognitiveComplexity != 1 {
		t.Errorf("func b: got %+v", results[1])
	}
}

func TestNestedForIf(t *testing.T) {
	got := cc(t, `package p
func f(items []int) {
	for range items {
		if true {}
	}
}`)
	if got != 3 {
		t.Errorf("nested for>if: got %d, want 3", got)
	}
}

func TestElseOnly(t *testing.T) {
	got := cc(t, `package p
func f(x bool) {
	if x {
	} else {
	}
}`)
	if got != 2 {
		t.Errorf("if/else: got %d, want 2", got)
	}
}

func TestLogicalOrSequence(t *testing.T) {
	got := cc(t, `package p
func f(a, b, c bool) {
	_ = a || b || c
}`)
	if got != 1 {
		t.Errorf("same || sequence: got %d, want 1", got)
	}
}

func TestMixedLogicalThreeOperators(t *testing.T) {
	got := cc(t, `package p
func f(a, b, c, d bool) {
	_ = a && b || c && d
}`)
	if got != 3 {
		t.Errorf("mixed three operators: got %d, want 3", got)
	}
}

func TestNestedSwitchInIf(t *testing.T) {
	got := cc(t, `package p
func f(x, y int) {
	if x > 0 {
		switch y {
		case 1:
		}
	}
}`)
	if got != 3 {
		t.Errorf("nested switch in if: got %d, want 3", got)
	}
}

func TestNoIncrementForCaseLabels(t *testing.T) {
	got := cc(t, `package p
func f(x int) {
	switch x {
	case 1:
	case 2:
	case 3:
	case 4:
	}
}`)
	if got != 1 {
		t.Errorf("case labels no increment: got %d, want 1", got)
	}
}
