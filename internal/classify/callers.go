package classify

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"

	"github.com/jflowers/gaze/internal/taxonomy"
)

// maxCallerWeight is the maximum weight for caller dependency
// signals.
const maxCallerWeight = 15

// AnalyzeCallerSignal scans TypesInfo.Uses across module packages
// to find call sites of the target function and computes a weight
// proportional to the ratio of callers that use/depend on the
// side effect.
func AnalyzeCallerSignal(
	funcObj types.Object,
	effectType taxonomy.SideEffectType,
	modulePkgs []*packages.Package,
) taxonomy.Signal {
	if funcObj == nil {
		return taxonomy.Signal{}
	}

	callerCount := countCallers(funcObj, modulePkgs)
	if callerCount == 0 {
		return taxonomy.Signal{}
	}

	// Weight is proportional to caller count, capped at max.
	// 1 caller = 5, 2-3 callers = 10, 4+ callers = 15.
	weight := 5
	if callerCount >= 4 {
		weight = maxCallerWeight
	} else if callerCount >= 2 {
		weight = 10
	}

	return taxonomy.Signal{
		Source: "caller",
		Weight: weight,
		Reasoning: fmt.Sprintf(
			"%d caller(s) reference this function",
			callerCount,
		),
	}
}

// countCallers counts the number of distinct packages that
// reference the given function object via TypesInfo.Uses.
func countCallers(funcObj types.Object, pkgs []*packages.Package) int {
	count := 0
	funcPkg := funcObj.Pkg()

	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			continue
		}
		// Skip the package that defines the function.
		if pkg.Types == funcPkg {
			continue
		}

		for _, obj := range pkg.TypesInfo.Uses {
			if obj == funcObj {
				count++
				break // Count each package only once.
			}
		}
	}

	return count
}
