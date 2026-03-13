package quality

// ResolveExprRoot exports resolveExprRoot for testing.
var ResolveExprRoot = resolveExprRoot

// SafeSSABuild is exported for testing. See safeSSABuild.
func SafeSSABuild(buildFn func()) any {
	return safeSSABuild(buildFn)
}
