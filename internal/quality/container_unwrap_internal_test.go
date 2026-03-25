package quality

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

// parseAndTypeCheck parses Go source code and type-checks it,
// returning the AST file and populated types.Info. Used by unit
// tests for the container unwrap helper functions.
func parseAndTypeCheck(t *testing.T, src string) (*ast.File, *types.Info) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{Importer: nil}
	_, _ = conf.Check("p", fset, []*ast.File{file}, info)
	return file, info
}

// extractCallFromFunc extracts a CallExpr from the specified function
// declaration and statement index. The funcIdx is the index into
// file.Decls, and stmtIdx is the index into the function's body
// statement list. Handles both ExprStmt (bare calls) and AssignStmt
// (calls as RHS).
func extractCallFromFunc(t *testing.T, file *ast.File, funcIdx, stmtIdx int) *ast.CallExpr {
	t.Helper()
	fn, ok := file.Decls[funcIdx].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("Decls[%d] is not a FuncDecl", funcIdx)
	}
	if stmtIdx >= len(fn.Body.List) {
		t.Fatalf("stmt index %d out of range (body has %d stmts)", stmtIdx, len(fn.Body.List))
	}
	stmt := fn.Body.List[stmtIdx]
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		call, ok := s.X.(*ast.CallExpr)
		if !ok {
			t.Fatalf("ExprStmt.X is not a CallExpr")
		}
		return call
	case *ast.AssignStmt:
		if len(s.Rhs) == 0 {
			t.Fatalf("AssignStmt has no RHS")
		}
		call, ok := s.Rhs[0].(*ast.CallExpr)
		if !ok {
			t.Fatalf("AssignStmt.Rhs[0] is not a CallExpr")
		}
		return call
	default:
		t.Fatalf("stmt is %T, expected ExprStmt or AssignStmt", stmt)
		return nil
	}
}

// TestIsDataExtraction_SelectorExpr verifies that field access
// expressions (x.Field) are classified as data extraction.
func TestIsDataExtraction_SelectorExpr(t *testing.T) {
	expr := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "x"},
		Sel: &ast.Ident{Name: "Field"},
	}
	if !isDataExtraction(expr) {
		t.Error("SelectorExpr should be classified as data extraction")
	}
}

// TestIsDataExtraction_IndexExpr verifies that index access
// expressions (x[i]) are classified as data extraction.
func TestIsDataExtraction_IndexExpr(t *testing.T) {
	expr := &ast.IndexExpr{
		X:     &ast.Ident{Name: "x"},
		Index: &ast.BasicLit{Kind: token.INT, Value: "0"},
	}
	if !isDataExtraction(expr) {
		t.Error("IndexExpr should be classified as data extraction")
	}
}

// TestIsDataExtraction_TypeAssertExpr verifies that type assertion
// expressions (x.(T)) are classified as data extraction.
func TestIsDataExtraction_TypeAssertExpr(t *testing.T) {
	expr := &ast.TypeAssertExpr{
		X:    &ast.Ident{Name: "x"},
		Type: &ast.Ident{Name: "int"},
	}
	if !isDataExtraction(expr) {
		t.Error("TypeAssertExpr should be classified as data extraction")
	}
}

// TestIsDataExtraction_TypeConversion verifies that type conversion
// expressions with *ast.Ident as Fun (e.g., string(x)) are
// classified as data extraction.
func TestIsDataExtraction_TypeConversion(t *testing.T) {
	expr := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "string"},
		Args: []ast.Expr{&ast.Ident{Name: "x"}},
	}
	if !isDataExtraction(expr) {
		t.Error("single-arg CallExpr with Ident Fun should be classified as data extraction (type conversion)")
	}
}

// TestIsDataExtraction_ArrayTypeConversion verifies that slice type
// conversion expressions ([]byte(x)) are classified as data extraction.
func TestIsDataExtraction_ArrayTypeConversion(t *testing.T) {
	expr := &ast.CallExpr{
		Fun: &ast.ArrayType{
			Elt: &ast.Ident{Name: "byte"},
		},
		Args: []ast.Expr{&ast.Ident{Name: "x"}},
	}
	if !isDataExtraction(expr) {
		t.Error("single-arg CallExpr with ArrayType Fun should be classified as data extraction ([]byte conversion)")
	}
}

// TestIsDataExtraction_ParenExpr verifies that parenthesized
// expressions recurse into the inner expression.
func TestIsDataExtraction_ParenExpr(t *testing.T) {
	inner := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "x"},
		Sel: &ast.Ident{Name: "Field"},
	}
	expr := &ast.ParenExpr{X: inner}
	if !isDataExtraction(expr) {
		t.Error("ParenExpr wrapping SelectorExpr should be classified as data extraction")
	}
}

// TestIsDataExtraction_MethodCall verifies that method calls
// (x.Method()) are NOT classified as data extraction.
func TestIsDataExtraction_MethodCall(t *testing.T) {
	expr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "s"},
			Sel: &ast.Ident{Name: "Get"},
		},
		Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"key"`}},
	}
	if isDataExtraction(expr) {
		t.Error("method call (SelectorExpr as Fun) should NOT be classified as data extraction")
	}
}

// TestIsDataExtraction_MultiArgCall verifies that multi-argument
// function calls are NOT classified as data extraction.
func TestIsDataExtraction_MultiArgCall(t *testing.T) {
	expr := &ast.CallExpr{
		Fun: &ast.Ident{Name: "process"},
		Args: []ast.Expr{
			&ast.Ident{Name: "x"},
			&ast.Ident{Name: "y"},
		},
	}
	if isDataExtraction(expr) {
		t.Error("multi-arg CallExpr should NOT be classified as data extraction")
	}
}

// TestIsDataExtraction_BareIdent verifies that a bare identifier
// is NOT classified as data extraction.
func TestIsDataExtraction_BareIdent(t *testing.T) {
	expr := &ast.Ident{Name: "x"}
	if isDataExtraction(expr) {
		t.Error("bare Ident should NOT be classified as data extraction")
	}
}

// extractVarObj extracts the types.Object for a variable defined in
// the given assignment statement's LHS at position 0.
func extractVarObj(t *testing.T, file *ast.File, info *types.Info, funcIdx, stmtIdx int) types.Object {
	t.Helper()
	fn, ok := file.Decls[funcIdx].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("Decls[%d] is not a FuncDecl", funcIdx)
	}
	assign, ok := fn.Body.List[stmtIdx].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("stmt[%d] is not AssignStmt", stmtIdx)
	}
	ident, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		t.Fatal("LHS[0] is not Ident")
	}
	obj := info.Defs[ident]
	if obj == nil {
		t.Fatalf("could not find types.Object for %s", ident.Name)
	}
	return obj
}

// extractRHS extracts the RHS expression from the given statement.
func extractRHS(t *testing.T, file *ast.File, funcIdx, stmtIdx int) ast.Expr {
	t.Helper()
	fn, ok := file.Decls[funcIdx].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("Decls[%d] is not a FuncDecl", funcIdx)
	}
	assign, ok := fn.Body.List[stmtIdx].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("stmt[%d] is not AssignStmt", stmtIdx)
	}
	return assign.Rhs[0]
}

// TestContainsObject_Found verifies that containsObject returns true
// when the target object is referenced in the expression.
func TestContainsObject_Found(t *testing.T) {
	file, info := parseAndTypeCheck(t, `package p; func f() { x := 1; _ = x + 2 }`)

	xObj := extractVarObj(t, file, info, 0, 0)
	rhs := extractRHS(t, file, 0, 1)

	if !containsObject(rhs, xObj, info) {
		t.Error("containsObject should find x in the expression x + 2")
	}
}

// TestContainsObject_NotFound verifies that containsObject returns
// false when the target object is not in the expression.
func TestContainsObject_NotFound(t *testing.T) {
	file, info := parseAndTypeCheck(t, `package p; func f() { x := 1; y := 2; _ = y + 3; _ = x }`)

	xObj := extractVarObj(t, file, info, 0, 0)
	rhs := extractRHS(t, file, 0, 2) // _ = y + 3

	if containsObject(rhs, xObj, info) {
		t.Error("containsObject should not find x in the expression y + 3")
	}
}

// TestContainsObject_NilInputs verifies that containsObject handles
// nil inputs gracefully.
func TestContainsObject_NilInputs(t *testing.T) {
	if containsObject(nil, nil, nil) {
		t.Error("containsObject(nil, nil, nil) should return false")
	}

	expr := &ast.Ident{Name: "x"}
	if containsObject(expr, nil, nil) {
		t.Error("containsObject with nil target should return false")
	}
}

// TestExtractPointerDest_AddressOf verifies that extractPointerDest
// correctly unwraps &data to find the underlying variable.
func TestExtractPointerDest_AddressOf(t *testing.T) {
	src := `package p
func g(x []byte, dst *int) {}
func f() {
	var data int
	g(nil, &data)
}`
	file, info := parseAndTypeCheck(t, src)
	call := extractCallFromFunc(t, file, 1, 1) // f(), stmt 1: g(nil, &data)

	dest := extractPointerDest(call, 1, info)
	if dest == nil {
		t.Fatal("extractPointerDest should find the variable from &data")
	}
	if dest.Name() != "data" {
		t.Errorf("expected variable name 'data', got %q", dest.Name())
	}
}

// TestExtractPointerDest_NilInputs verifies that extractPointerDest
// handles nil and out-of-bounds inputs gracefully.
func TestExtractPointerDest_NilInputs(t *testing.T) {
	if extractPointerDest(nil, 0, nil) != nil {
		t.Error("extractPointerDest(nil, 0, nil) should return nil")
	}

	call := &ast.CallExpr{Args: []ast.Expr{&ast.Ident{Name: "x"}}}
	if extractPointerDest(call, 5, nil) != nil {
		t.Error("extractPointerDest with out-of-bounds index should return nil")
	}
}

// TestExtractPointerDest_BareIdent verifies that extractPointerDest
// handles a bare identifier argument (already a pointer variable).
func TestExtractPointerDest_BareIdent(t *testing.T) {
	src := `package p
func f() {
	var p *int
	g(p)
}
func g(x *int) {}`
	file, info := parseAndTypeCheck(t, src)
	call := extractCallFromFunc(t, file, 0, 1) // f(), stmt 1: g(p)

	dest := extractPointerDest(call, 0, info)
	if dest == nil {
		t.Fatal("extractPointerDest should find the variable from bare pointer ident")
	}
	if dest.Name() != "p" {
		t.Errorf("expected variable name 'p', got %q", dest.Name())
	}
}

// TestIsTransformationCall_ByteSliceAndPointer verifies that a
// function with []byte and *T parameters is detected as a
// transformation call.
func TestIsTransformationCall_ByteSliceAndPointer(t *testing.T) {
	src := `package p
func unmarshal(data []byte, dst *int) {}
func f() { var x int; unmarshal(nil, &x) }`
	file, info := parseAndTypeCheck(t, src)
	call := extractCallFromFunc(t, file, 1, 1) // f(), stmt 1: unmarshal(nil, &x)

	byteIdx, ptrIdx, ok := isTransformationCall(call, info)
	if !ok {
		t.Fatal("isTransformationCall should match func([]byte, *int)")
	}
	if byteIdx != 0 {
		t.Errorf("byteArgIdx = %d, want 0", byteIdx)
	}
	if ptrIdx != 1 {
		t.Errorf("ptrDestIdx = %d, want 1", ptrIdx)
	}
}

// TestIsTransformationCall_StringAndPointer verifies that a function
// with string and *T parameters is detected as a transformation call.
func TestIsTransformationCall_StringAndPointer(t *testing.T) {
	src := `package p
func decode(s string, dst *int) {}
func f() { var x int; decode("", &x) }`
	file, info := parseAndTypeCheck(t, src)
	call := extractCallFromFunc(t, file, 1, 1) // f(), stmt 1: decode("", &x)

	byteIdx, ptrIdx, ok := isTransformationCall(call, info)
	if !ok {
		t.Fatal("isTransformationCall should match func(string, *int)")
	}
	if byteIdx != 0 {
		t.Errorf("byteArgIdx = %d, want 0", byteIdx)
	}
	if ptrIdx != 1 {
		t.Errorf("ptrDestIdx = %d, want 1", ptrIdx)
	}
}

// TestIsTransformationCall_NoPointer verifies that a function without
// a pointer parameter is not detected as a transformation call.
func TestIsTransformationCall_NoPointer(t *testing.T) {
	src := `package p
func process(data []byte, n int) {}
func f() { process(nil, 0) }`
	file, info := parseAndTypeCheck(t, src)
	call := extractCallFromFunc(t, file, 1, 0) // f(), stmt 0: process(nil, 0)

	_, _, ok := isTransformationCall(call, info)
	if ok {
		t.Error("isTransformationCall should not match func([]byte, int) — no pointer dest")
	}
}

// TestIsTransformationCall_NoByteLike verifies that a function without
// a byte-like parameter is not detected as a transformation call.
func TestIsTransformationCall_NoByteLike(t *testing.T) {
	src := `package p
func store(n int, dst *int) {}
func f() { var x int; store(0, &x) }`
	file, info := parseAndTypeCheck(t, src)
	call := extractCallFromFunc(t, file, 1, 1) // f(), stmt 1: store(0, &x)

	_, _, ok := isTransformationCall(call, info)
	if ok {
		t.Error("isTransformationCall should not match func(int, *int) — no byte-like input")
	}
}

// TestIsTransformationCall_NilInputs verifies that isTransformationCall
// handles nil inputs gracefully.
func TestIsTransformationCall_NilInputs(t *testing.T) {
	_, _, ok := isTransformationCall(nil, nil)
	if ok {
		t.Error("isTransformationCall(nil, nil) should return ok=false")
	}
}
