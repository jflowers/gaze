package analysis_test

import (
	"testing"

	"github.com/jflowers/gaze/internal/analysis"
)

func BenchmarkAnalyzeFunction_Returns(b *testing.B) {
	pkg := loadTestPackageBench(b, "returns")
	fd := analysis.FindFuncDecl(pkg, "ErrorReturn")
	if fd == nil {
		b.Fatal("ErrorReturn not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analysis.AnalyzeFunction(pkg, fd)
	}
}

func BenchmarkAnalyzeFunction_Mutation(b *testing.B) {
	pkg := loadTestPackageBench(b, "mutation")
	fd := analysis.FindMethodDecl(pkg, "*Counter", "Increment")
	if fd == nil {
		b.Fatal("(*Counter).Increment not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analysis.AnalyzeFunction(pkg, fd)
	}
}

func BenchmarkAnalyze_AllFunctions(b *testing.B) {
	pkg := loadTestPackageBench(b, "returns")

	opts := analysis.Options{IncludeUnexported: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analysis.Analyze(pkg, opts)
	}
}

func BenchmarkAnalyze_MutationPackage(b *testing.B) {
	pkg := loadTestPackageBench(b, "mutation")

	opts := analysis.Options{IncludeUnexported: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analysis.Analyze(pkg, opts)
	}
}
