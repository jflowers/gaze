// Package cognitive computes per-function cognitive complexity
// following the SonarSource cognitive complexity specification,
// adapted for Go AST.
//
// Cognitive complexity measures how hard code is to understand
// by penalizing nesting, recursion, and breaks in linear flow.
// Unlike cyclomatic complexity, it increases with each nested
// control structure and with recursive calls.
package cognitive

import (
	"go/ast"
	"go/token"
)

// FuncCognitiveComplexity holds the cognitive complexity for a
// single function.
type FuncCognitiveComplexity struct {
	// Package is the package name.
	Package string `json:"package"`

	// Function is the function or method name.
	Function string `json:"function"`

	// File is the absolute filesystem path to the source file.
	File string `json:"file"`

	// Line is the line number of the function declaration.
	Line int `json:"line"`

	// CognitiveComplexity is the cognitive complexity value.
	CognitiveComplexity int `json:"cognitive_complexity"`
}

// AnalyzeFunc computes the cognitive complexity of a single
// function declaration following the SonarSource specification
// adapted for Go AST.
//
// Increment rules:
//   - if: +1 (+ nesting level)
//   - else if: +1 (no nesting penalty)
//   - else: +1 (no nesting penalty)
//   - switch, type switch: +1 (+ nesting level)
//   - for, range: +1 (+ nesting level)
//   - select: +1 (+ nesting level, treated like switch)
//   - &&, ||: +1 per mixed-operator sequence
//   - goto: +1
//   - recursion: +1
func AnalyzeFunc(fset *token.FileSet, funcDecl *ast.FuncDecl) int {
	if funcDecl.Body == nil {
		return 0
	}
	a := &analyzer{
		funcName: funcDecl.Name.Name,
	}
	a.walkBody(funcDecl.Body, 0)
	return a.total
}

// AnalyzeFile computes cognitive complexity for all function
// declarations in a file.
func AnalyzeFile(fset *token.FileSet, file *ast.File) []FuncCognitiveComplexity {
	var results []FuncCognitiveComplexity
	pkgName := file.Name.Name
	fpos := fset.Position(file.Pos())
	fileName := fpos.Filename
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		cc := AnalyzeFunc(fset, funcDecl)
		fp := fset.Position(funcDecl.Pos())
		results = append(results, FuncCognitiveComplexity{
			Package:             pkgName,
			Function:            funcDecl.Name.Name,
			File:                fileName,
			Line:                fp.Line,
			CognitiveComplexity: cc,
		})
	}
	return results
}

type analyzer struct {
	funcName string
	total    int
}

func (a *analyzer) walkBody(body *ast.BlockStmt, nesting int) {
	if body == nil {
		return
	}
	for _, stmt := range body.List {
		a.walkStmt(stmt, nesting)
	}
}

func (a *analyzer) walkStmt(stmt ast.Stmt, nesting int) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		a.walkIf(s, nesting)
	case *ast.SwitchStmt:
		a.total += 1 + nesting
		a.walkBody(s.Body, nesting+1)
	case *ast.TypeSwitchStmt:
		a.total += 1 + nesting
		a.walkBody(s.Body, nesting+1)
	case *ast.SelectStmt:
		a.total += 1 + nesting
		a.walkBody(s.Body, nesting+1)
	case *ast.ForStmt:
		a.total += 1 + nesting
		a.walkBody(s.Body, nesting+1)
	case *ast.RangeStmt:
		a.total += 1 + nesting
		a.walkBody(s.Body, nesting+1)
	case *ast.BlockStmt:
		a.walkBody(s, nesting)
	case *ast.BranchStmt:
		if s.Tok == token.GOTO {
			a.total++
		}
	default:
		a.walkExprStmt(stmt, nesting)
	}
}

func (a *analyzer) walkIf(s *ast.IfStmt, nesting int) {
	a.total += 1 + nesting
	a.walkBody(s.Body, nesting+1)
	a.walkElse(s.Else, nesting)
}

func (a *analyzer) walkElse(elseStmt ast.Stmt, nesting int) {
	if elseStmt == nil {
		return
	}
	if ifStmt, ok := elseStmt.(*ast.IfStmt); ok {
		a.total++
		a.walkBody(ifStmt.Body, nesting+1)
		a.walkElse(ifStmt.Else, nesting)
		return
	}
	if block, ok := elseStmt.(*ast.BlockStmt); ok {
		a.total++
		a.walkBody(block, nesting+1)
	}
}

func (a *analyzer) walkExprStmt(stmt ast.Stmt, nesting int) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		a.walkExpr(s.X, nesting, token.ILLEGAL)
	case *ast.AssignStmt:
		for _, rhs := range s.Rhs {
			a.walkExpr(rhs, nesting, token.ILLEGAL)
		}
		for _, lhs := range s.Lhs {
			a.walkExpr(lhs, nesting, token.ILLEGAL)
		}
	case *ast.ReturnStmt:
		for _, result := range s.Results {
			a.walkExpr(result, nesting, token.ILLEGAL)
		}
	case *ast.SendStmt:
		a.walkExpr(s.Chan, nesting, token.ILLEGAL)
		a.walkExpr(s.Value, nesting, token.ILLEGAL)
	case *ast.IncDecStmt:
		a.walkExpr(s.X, nesting, token.ILLEGAL)
	case *ast.DeferStmt:
		a.walkCallExpr(s.Call, nesting)
		for _, arg := range s.Call.Args {
			a.walkExpr(arg, nesting, token.ILLEGAL)
		}
	case *ast.GoStmt:
		a.walkCallExpr(s.Call, nesting)
		for _, arg := range s.Call.Args {
			a.walkExpr(arg, nesting, token.ILLEGAL)
		}
	case *ast.DeclStmt:
		if gen, ok := s.Decl.(*ast.GenDecl); ok {
			for _, spec := range gen.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for _, val := range vs.Values {
						a.walkExpr(val, nesting, token.ILLEGAL)
					}
				}
			}
		}
	}
}

func (a *analyzer) walkExpr(expr ast.Expr, nesting int, parentLogOp token.Token) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		if e.Op == token.LAND || e.Op == token.LOR {
			if parentLogOp != e.Op {
				a.total++
			}
			a.walkExpr(e.X, nesting, e.Op)
			a.walkExpr(e.Y, nesting, e.Op)
		} else {
			a.walkExpr(e.X, nesting, token.ILLEGAL)
			a.walkExpr(e.Y, nesting, token.ILLEGAL)
		}
	case *ast.CallExpr:
		a.walkCallExpr(e, nesting)
		for _, arg := range e.Args {
			a.walkExpr(arg, nesting, token.ILLEGAL)
		}
	case *ast.UnaryExpr:
		a.walkExpr(e.X, nesting, token.ILLEGAL)
	case *ast.ParenExpr:
		a.walkExpr(e.X, nesting, token.ILLEGAL)
	case *ast.IndexExpr:
		a.walkExpr(e.X, nesting, token.ILLEGAL)
		a.walkExpr(e.Index, nesting, token.ILLEGAL)
	case *ast.SliceExpr:
		a.walkExpr(e.X, nesting, token.ILLEGAL)
		if e.Low != nil {
			a.walkExpr(e.Low, nesting, token.ILLEGAL)
		}
		if e.High != nil {
			a.walkExpr(e.High, nesting, token.ILLEGAL)
		}
	case *ast.SelectorExpr:
		a.walkExpr(e.X, nesting, token.ILLEGAL)
	case *ast.FuncLit:
		a.walkBody(e.Body, nesting)
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			a.walkExpr(elt, nesting, token.ILLEGAL)
		}
	case *ast.KeyValueExpr:
		a.walkExpr(e.Value, nesting, token.ILLEGAL)
	case *ast.StarExpr:
		a.walkExpr(e.X, nesting, token.ILLEGAL)
	case *ast.TypeAssertExpr:
		a.walkExpr(e.X, nesting, token.ILLEGAL)
	case *ast.MapType:
	case *ast.ArrayType:
	case *ast.ChanType:
	}
}

func (a *analyzer) walkCallExpr(e *ast.CallExpr, nesting int) {
	if ident, ok := e.Fun.(*ast.Ident); ok {
		if ident.Name == a.funcName {
			a.total++
		}
	}
}
