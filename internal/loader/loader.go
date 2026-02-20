// Package loader wraps go/packages to load Go packages with full
// type information for static analysis.
package loader

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// LoadMode is the minimum set of flags needed for SSA-ready analysis.
const LoadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedTypesSizes

// Result holds the loaded package along with convenience accessors.
type Result struct {
	// Pkg is the loaded package.
	Pkg *packages.Package

	// Fset is the shared file set for position information.
	Fset *token.FileSet
}

// FuncInfo holds a function declaration with its type information.
type FuncInfo struct {
	// Decl is the AST function declaration.
	Decl *ast.FuncDecl

	// Obj is the types.Func object for this function.
	Obj *types.Func

	// Pkg is the package this function belongs to.
	Pkg *packages.Package
}

// Load loads a Go package at the given import path or file pattern.
// It returns the loaded package result or an error if loading or
// type-checking fails.
func Load(pattern string) (*Result, error) {
	cfg := &packages.Config{
		Mode:  LoadMode,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, fmt.Errorf("loading package %q: %w", pattern, err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found for pattern %q", pattern)
	}

	pkg := pkgs[0]

	// Check for package-level errors (syntax, type errors, etc.).
	var errs []string
	for _, e := range pkg.Errors {
		errs = append(errs, e.Error())
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("package %q has errors:\n  %s",
			pattern, strings.Join(errs, "\n  "))
	}

	return &Result{
		Pkg:  pkg,
		Fset: pkg.Fset,
	}, nil
}

// FindFunction looks up a function or method by name within the
// loaded package. Returns nil if not found.
func (r *Result) FindFunction(name string) *FuncInfo {
	for _, file := range r.Pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Name == nil {
				continue
			}
			if fd.Name.Name == name {
				obj := r.Pkg.TypesInfo.Defs[fd.Name]
				if obj == nil {
					continue
				}
				fn, ok := obj.(*types.Func)
				if !ok {
					continue
				}
				return &FuncInfo{
					Decl: fd,
					Obj:  fn,
					Pkg:  r.Pkg,
				}
			}
		}
	}
	return nil
}

// AllFunctions returns all function declarations in the package.
// If exportedOnly is true, only exported functions are returned.
func (r *Result) AllFunctions(exportedOnly bool) []*FuncInfo {
	var funcs []*FuncInfo
	for _, file := range r.Pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Name == nil {
				continue
			}
			if exportedOnly && !fd.Name.IsExported() {
				continue
			}
			obj := r.Pkg.TypesInfo.Defs[fd.Name]
			if obj == nil {
				continue
			}
			fn, ok := obj.(*types.Func)
			if !ok {
				continue
			}
			funcs = append(funcs, &FuncInfo{
				Decl: fd,
				Obj:  fn,
				Pkg:  r.Pkg,
			})
		}
	}
	return funcs
}

// FormatPos returns a "file:line:col" string for the given position.
func (r *Result) FormatPos(pos token.Pos) string {
	if !pos.IsValid() {
		return "<unknown>"
	}
	p := r.Fset.Position(pos)
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}
