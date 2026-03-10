package aireport

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildFakeGemini compiles the fake_gemini binary and returns its path.
func buildFakeGemini(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("subprocess tests not supported on Windows in CI")
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "gemini")

	srcDir := filepath.Join("internal", "aireport", "testdata", "fake_gemini")
	cmd := exec.Command("go", "build", "-o", bin, "./"+srcDir)
	cmd.Dir = findModuleRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building fake_gemini: %v\n%s", err, out)
	}
	return bin
}

// withGeminiOnPath temporarily makes the fake gemini binary available as
// "gemini" on PATH.
func withGeminiOnPath(t *testing.T, bin string) {
	t.Helper()
	origPath := os.Getenv("PATH")
	dir := filepath.Dir(bin)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)
}

func TestGeminiAdapter_SuccessfulInvocation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeGemini(t)
	withGeminiOnPath(t, bin)

	adapter := &GeminiAdapter{config: AdapterConfig{Name: "gemini"}}
	got, err := adapter.Format(context.Background(), "system instructions", strings.NewReader(`{"crap":{}}`))
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(got, "Fake Gemini Report") {
		t.Errorf("expected fake report in output, got: %q", got)
	}
}

func TestGeminiAdapter_GEMINIMDWrittenToTempDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeGemini(t)
	withGeminiOnPath(t, bin)

	// The fake_gemini binary checks for GEMINI.md in its working directory.
	// A successful invocation proves the file was written.
	adapter := &GeminiAdapter{config: AdapterConfig{Name: "gemini"}}
	_, err := adapter.Format(context.Background(), "# System Prompt\n\nInstructions.", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("Format failed (GEMINI.md may not have been written): %v", err)
	}
}

func TestGeminiAdapter_ModelFlagPassed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeGemini(t)
	withGeminiOnPath(t, bin)

	adapter := &GeminiAdapter{config: AdapterConfig{Name: "gemini", Model: "gemini-2.5-pro"}}
	got, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(got, "gemini-2.5-pro") {
		t.Errorf("expected model name in output, got: %q", got)
	}
}

func TestGeminiAdapter_NotOnPath_ReturnsError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	adapter := &GeminiAdapter{config: AdapterConfig{Name: "gemini"}}
	_, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected error when gemini not on PATH")
	}
	if !strings.Contains(err.Error(), "FR-012") {
		t.Errorf("expected FR-012 in error, got: %v", err)
	}
}

func TestGeminiAdapter_NonZeroExit_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeGemini(t)

	wrapDir := t.TempDir()
	geminiPath := filepath.Join(wrapDir, "gemini")
	errWrapper := "#!/bin/sh\nexec \"" + bin + "\" --exit-error \"$@\"\n"
	if err := os.WriteFile(geminiPath, []byte(errWrapper), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", wrapDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	adapter := &GeminiAdapter{config: AdapterConfig{Name: "gemini"}}
	_, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected error on non-zero exit")
	}
}

func TestGeminiAdapter_EmptyResponse_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeGemini(t)

	wrapDir := t.TempDir()
	geminiPath := filepath.Join(wrapDir, "gemini")
	// fake_gemini --empty-output writes {"response": ""}.
	emptyWrapper := "#!/bin/sh\nexec \"" + bin + "\" --empty-output \"$@\"\n"
	if err := os.WriteFile(geminiPath, []byte(emptyWrapper), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", wrapDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	adapter := &GeminiAdapter{config: AdapterConfig{Name: "gemini"}}
	_, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected FR-016 error for empty response")
	}
	if !strings.Contains(err.Error(), "FR-016") {
		t.Errorf("expected FR-016 in error, got: %v", err)
	}
}

func TestGeminiAdapter_TempDirCleanedUp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeGemini(t)
	withGeminiOnPath(t, bin)

	tmpBase := os.TempDir()
	beforeEntries, _ := os.ReadDir(tmpBase)
	before := countGazeGeminiDirs(beforeEntries)

	adapter := &GeminiAdapter{config: AdapterConfig{Name: "gemini"}}
	_, _ = adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))

	afterEntries, _ := os.ReadDir(tmpBase)
	after := countGazeGeminiDirs(afterEntries)

	if after > before {
		t.Errorf("temp directories leaked: before=%d after=%d", before, after)
	}
}

func TestGeminiAdapter_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeGemini(t)
	withGeminiOnPath(t, bin)

	// Cancel the context before calling Format so no time.Sleep is needed.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	adapter := &GeminiAdapter{config: AdapterConfig{Name: "gemini"}}
	_, err := adapter.Format(ctx, "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func countGazeGeminiDirs(entries []os.DirEntry) int {
	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "gaze-gemini-") {
			count++
		}
	}
	return count
}
