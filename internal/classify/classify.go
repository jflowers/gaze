package classify

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"

	"github.com/jflowers/gaze/internal/config"
	"github.com/jflowers/gaze/internal/taxonomy"
)

// Options configures the classification engine.
type Options struct {
	// Config is the Gaze configuration. If nil, defaults are used.
	Config *config.GazeConfig

	// ModulePackages is the list of all packages in the module,
	// used for interface satisfaction and caller analysis.
	ModulePackages []*packages.Package

	// TargetPkg is the loaded target package (for AST access).
	TargetPkg *packages.Package

	// Verbose controls whether signal detail fields (SourceFile,
	// Excerpt, Reasoning) are populated.
	Verbose bool
}

// Classify classifies each side effect in the given analysis
// results using mechanical signal analyzers. It attaches a
// Classification to each SideEffect and returns the modified
// results.
func Classify(results []taxonomy.AnalysisResult, opts Options) []taxonomy.AnalysisResult {
	if opts.Config == nil {
		opts.Config = config.DefaultConfig()
	}

	// Build a lookup from function name to AST declaration and
	// types.Object for the target package.
	funcDecls := buildFuncDeclMap(opts.TargetPkg)
	funcObjs := buildFuncObjMap(opts.TargetPkg)

	for i := range results {
		result := &results[i]
		funcName := result.Target.Function
		funcDecl := funcDecls[funcName]
		funcObj := funcObjs[funcName]

		// Determine receiver type if this is a method.
		var receiverType types.Type
		if funcObj != nil {
			if sig, ok := funcObj.Type().(*types.Signature); ok && sig.Recv() != nil {
				receiverType = sig.Recv().Type()
				// Unwrap pointer for interface checks.
				if ptr, ok := receiverType.(*types.Pointer); ok {
					receiverType = ptr.Elem()
				}
			}
		}

		for j := range result.SideEffects {
			se := &result.SideEffects[j]
			signals := classifySideEffect(
				funcName, funcDecl, funcObj,
				receiverType, se.Type,
				opts,
			)

			classification := ComputeScore(signals, opts.Config)

			// Strip detail fields if not verbose.
			if !opts.Verbose {
				for k := range classification.Signals {
					classification.Signals[k].SourceFile = ""
					classification.Signals[k].Excerpt = ""
					classification.Signals[k].Reasoning = ""
				}
			}

			se.Classification = &classification
		}
	}

	return results
}

// classifySideEffect runs all five mechanical signal analyzers
// for a single side effect and returns the collected signals.
func classifySideEffect(
	funcName string,
	funcDecl *ast.FuncDecl,
	funcObj types.Object,
	receiverType types.Type,
	effectType taxonomy.SideEffectType,
	opts Options,
) []taxonomy.Signal {
	var signals []taxonomy.Signal

	// 1. Interface satisfaction.
	if s := AnalyzeInterfaceSignal(funcName, receiverType, effectType, opts.ModulePackages); s.Source != "" {
		signals = append(signals, s)
	}

	// 2. API surface visibility.
	if s := AnalyzeVisibilitySignal(funcDecl, funcObj, effectType); s.Source != "" {
		signals = append(signals, s)
	}

	// 3. Caller dependency.
	if s := AnalyzeCallerSignal(funcObj, effectType, opts.ModulePackages); s.Source != "" {
		signals = append(signals, s)
	}

	// 4. Naming convention.
	if s := AnalyzeNamingSignal(funcName, effectType); s.Source != "" {
		signals = append(signals, s)
	}

	// 5. Godoc comment.
	if s := AnalyzeGodocSignal(funcDecl, effectType); s.Source != "" {
		signals = append(signals, s)
	}

	return signals
}

// buildFuncDeclMap creates a lookup from function/method name to
// its AST declaration in the given package.
func buildFuncDeclMap(pkg *packages.Package) map[string]*ast.FuncDecl {
	m := make(map[string]*ast.FuncDecl)
	if pkg == nil {
		return m
	}
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			m[fd.Name.Name] = fd
		}
	}
	return m
}

// buildFuncObjMap creates a lookup from function/method name to
// its types.Object in the given package.
func buildFuncObjMap(pkg *packages.Package) map[string]types.Object {
	m := make(map[string]types.Object)
	if pkg == nil {
		return m
	}
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if _, ok := obj.(*types.Func); ok {
			m[name] = obj
		}
	}

	// Also look up methods on named types.
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		for i := 0; i < named.NumMethods(); i++ {
			method := named.Method(i)
			m[method.Name()] = method
		}
	}

	return m
}
