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

// buildFakeOpenCode compiles the fake_opencode binary and returns its path.
// The binary is placed in t.TempDir() and cleaned up automatically.
func buildFakeOpenCode(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("subprocess tests not supported on Windows in CI")
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "opencode")

	srcDir := filepath.Join("internal", "aireport", "testdata", "fake_opencode")
	cmd := exec.Command("go", "build", "-o", bin, "./"+srcDir)
	cmd.Dir = findModuleRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building fake_opencode: %v\n%s", err, out)
	}
	return bin
}

// withOpenCodeOnPath temporarily adds the directory containing the fake opencode
// binary to PATH, restoring the original value on cleanup.
func withOpenCodeOnPath(t *testing.T, bin string) {
	t.Helper()
	origPath := os.Getenv("PATH")
	dir := filepath.Dir(bin)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)
}

func TestOpenCodeAdapter_SuccessfulInvocation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)
	withOpenCodeOnPath(t, bin)

	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	got, err := adapter.Format(context.Background(), "system instructions", strings.NewReader(`{"crap":{}}`))
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(got, "Fake OpenCode Report") {
		t.Errorf("expected fake report in output, got: %q", got)
	}
}

func TestOpenCodeAdapter_AgentFileWrittenToTempDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)
	withOpenCodeOnPath(t, bin)

	// The fake_opencode binary verifies that .opencode/agents/gaze-reporter.md
	// exists under --dir. A successful invocation proves the file was written.
	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	_, err := adapter.Format(context.Background(), "# System Prompt\n\nInstructions.", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("Format failed (agent file may not have been written): %v", err)
	}
}

func TestOpenCodeAdapter_FrontmatterWritten(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)
	withOpenCodeOnPath(t, bin)

	// The fake_opencode binary reads the agent file and asserts it starts with
	// "---" (frontmatter present). A successful invocation proves frontmatter
	// was prepended by the adapter (FR-003).
	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	_, err := adapter.Format(context.Background(), "prompt without frontmatter", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("Format failed (frontmatter may not have been written): %v", err)
	}
}

func TestOpenCodeAdapter_ModelFlagPassed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)
	withOpenCodeOnPath(t, bin)

	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode", Model: "claude-3-5-sonnet"}}
	got, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(got, "claude-3-5-sonnet") {
		t.Errorf("expected model name in output, got: %q", got)
	}
}

func TestOpenCodeAdapter_NoModelFlag_Succeeds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)
	withOpenCodeOnPath(t, bin)

	// When cfg.Model is empty, no -m flag is passed; fake_opencode succeeds
	// without a model name in output.
	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	got, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if strings.Contains(got, "model:") {
		t.Errorf("unexpected model string in output when no model specified, got: %q", got)
	}
}

func TestOpenCodeAdapter_NotOnPath_ReturnsError(t *testing.T) {
	// Override PATH with an empty temp directory so opencode is not found.
	// No testing.Short() guard needed — no subprocess is compiled or started.
	t.Setenv("PATH", t.TempDir())

	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	_, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected error when opencode not on PATH")
	}
	if !strings.Contains(err.Error(), "FR-007") {
		t.Errorf("expected FR-007 in error, got: %v", err)
	}
}

func TestOpenCodeAdapter_NonZeroExit_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)

	// Create a shell wrapper named "opencode" that injects --exit-error after
	// the "run" subcommand arg so that the fake binary's flag parser sees it.
	// The adapter passes args as: run --dir ... --agent ... --format ... ""
	// so $1="run" and the rest are flags. We forward $1 (run) then inject
	// --exit-error before the remaining args.
	wrapDir := t.TempDir()
	opencodePath := filepath.Join(wrapDir, "opencode")
	errWrapper := "#!/bin/sh\nshift\nexec \"" + bin + "\" run --exit-error \"$@\"\n"
	if err := os.WriteFile(opencodePath, []byte(errWrapper), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", wrapDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	_, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected error on non-zero exit")
	}
}

func TestOpenCodeAdapter_EmptyOutput_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)

	// Create a shell wrapper that injects --empty-output after the "run"
	// subcommand arg so that the fake binary's flag parser sees it correctly.
	wrapDir := t.TempDir()
	opencodePath := filepath.Join(wrapDir, "opencode")
	emptyWrapper := "#!/bin/sh\nshift\nexec \"" + bin + "\" run --empty-output \"$@\"\n"
	if err := os.WriteFile(opencodePath, []byte(emptyWrapper), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", wrapDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	_, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected FR-009 error for empty output")
	}
	if !strings.Contains(err.Error(), "FR-009") {
		t.Errorf("expected FR-009 in error, got: %v", err)
	}
}

func TestOpenCodeAdapter_TempDirCleanedUp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)
	withOpenCodeOnPath(t, bin)

	tmpBase := os.TempDir()
	beforeEntries, _ := os.ReadDir(tmpBase)
	before := countGazeOpenCodeDirs(beforeEntries)

	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	_, _ = adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))

	afterEntries, _ := os.ReadDir(tmpBase)
	after := countGazeOpenCodeDirs(afterEntries)

	if after > before {
		t.Errorf("temp directories leaked: before=%d after=%d", before, after)
	}
}

func TestOpenCodeAdapter_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)
	withOpenCodeOnPath(t, bin)

	// Cancel the context before calling Format so the subprocess is never
	// started or is immediately killed — no time.Sleep needed.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	_, err := adapter.Format(ctx, "prompt", strings.NewReader(`{}`))
	// Either context.Canceled or a subprocess kill error is acceptable.
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

// TestOpenCodeAdapter_ValidateBinary_NotOnPath verifies that ValidateBinary
// returns an error containing "FR-007" when opencode is not on PATH.
func TestOpenCodeAdapter_ValidateBinary_NotOnPath(t *testing.T) {
	// Isolate PATH so opencode is definitely not found.
	t.Setenv("PATH", t.TempDir())

	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	err := adapter.ValidateBinary()
	if err == nil {
		t.Fatal("expected error when opencode not on PATH")
	}
	if !strings.Contains(err.Error(), "FR-007") {
		t.Errorf("expected FR-007 in ValidateBinary error, got: %v", err)
	}
}

// TestOpenCodeAdapter_ValidateBinary_OnPath verifies that ValidateBinary
// returns nil when opencode is on PATH.
func TestOpenCodeAdapter_ValidateBinary_OnPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeOpenCode(t)
	withOpenCodeOnPath(t, bin)

	adapter := &OpenCodeAdapter{config: AdapterConfig{Name: "opencode"}}
	if err := adapter.ValidateBinary(); err != nil {
		t.Errorf("expected nil from ValidateBinary when opencode on PATH, got: %v", err)
	}
}

func countGazeOpenCodeDirs(entries []os.DirEntry) int {
	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "gaze-opencode-") {
			count++
		}
	}
	return count
}
