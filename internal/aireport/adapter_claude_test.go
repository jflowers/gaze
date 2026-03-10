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

// buildFakeClaude compiles the fake_claude binary and returns its path.
// The binary is placed in t.TempDir() and cleaned up automatically.
func buildFakeClaude(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("subprocess tests not supported on Windows in CI")
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "claude")

	srcDir := filepath.Join("internal", "aireport", "testdata", "fake_claude")
	cmd := exec.Command("go", "build", "-o", bin, "./"+srcDir)
	cmd.Dir = findModuleRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building fake_claude: %v\n%s", err, out)
	}
	return bin
}

// findModuleRoot returns the module root directory.
func findModuleRoot(t *testing.T) string {
	t.Helper()
	// Walk upward from the current directory to find go.mod.
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

// withClaudeOnPath temporarily adds the directory containing the fake claude
// binary to PATH, restoring the original value on cleanup.
func withClaudeOnPath(t *testing.T, bin string) {
	t.Helper()
	origPath := os.Getenv("PATH")
	dir := filepath.Dir(bin)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)
}

func TestClaudeAdapter_SuccessfulInvocation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeClaude(t)
	withClaudeOnPath(t, bin)

	adapter := &ClaudeAdapter{config: AdapterConfig{Name: "claude"}}
	got, err := adapter.Format(context.Background(), "system instructions", strings.NewReader(`{"crap":{}}`))
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(got, "Fake Claude Report") {
		t.Errorf("expected fake report in output, got: %q", got)
	}
}

func TestClaudeAdapter_ModelFlagPassed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeClaude(t)
	withClaudeOnPath(t, bin)

	adapter := &ClaudeAdapter{config: AdapterConfig{Name: "claude", Model: "claude-3-haiku"}}
	got, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(got, "claude-3-haiku") {
		t.Errorf("expected model name in output, got: %q", got)
	}
}

func TestClaudeAdapter_NotOnPath_ReturnsError(t *testing.T) {
	// Override PATH with an empty directory so claude is not found.
	t.Setenv("PATH", t.TempDir())

	adapter := &ClaudeAdapter{config: AdapterConfig{Name: "claude"}}
	_, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected error when claude not on PATH")
	}
	if !strings.Contains(err.Error(), "FR-012") {
		t.Errorf("expected FR-012 in error, got: %v", err)
	}
}

func TestClaudeAdapter_NonZeroExit_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeClaude(t)

	// Replace PATH with a dir containing a "claude" that exits non-zero.
	wrapDir := t.TempDir()
	claudePath := filepath.Join(wrapDir, "claude")
	errWrapper := "#!/bin/sh\nexec \"" + bin + "\" --exit-error \"$@\"\n"
	if err := os.WriteFile(claudePath, []byte(errWrapper), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", wrapDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	adapter := &ClaudeAdapter{config: AdapterConfig{Name: "claude"}}
	_, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected error on non-zero exit")
	}
}

func TestClaudeAdapter_EmptyOutput_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeClaude(t)

	// Create a "claude" that returns empty output.
	wrapDir := t.TempDir()
	claudePath := filepath.Join(wrapDir, "claude")
	emptyWrapper := "#!/bin/sh\nexec \"" + bin + "\" --empty-output \"$@\"\n"
	if err := os.WriteFile(claudePath, []byte(emptyWrapper), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", wrapDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	adapter := &ClaudeAdapter{config: AdapterConfig{Name: "claude"}}
	_, err := adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))
	if err == nil {
		t.Fatal("expected FR-016 error for empty output")
	}
	if !strings.Contains(err.Error(), "FR-016") {
		t.Errorf("expected FR-016 in error, got: %v", err)
	}
}

func TestClaudeAdapter_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeClaude(t)
	withClaudeOnPath(t, bin)

	// Cancel the context before calling Format so the subprocess is never started
	// or is immediately killed — no time.Sleep needed for synchronization.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	adapter := &ClaudeAdapter{config: AdapterConfig{Name: "claude"}}
	_, err := adapter.Format(ctx, "prompt", strings.NewReader(`{}`))
	// Either context.Canceled or a subprocess kill error is acceptable.
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func TestClaudeAdapter_TempFileCleanedUp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in -short mode")
	}
	bin := buildFakeClaude(t)
	withClaudeOnPath(t, bin)

	// Count temp files before.
	tmpDir := os.TempDir()
	beforeEntries, _ := os.ReadDir(tmpDir)
	before := countGazeClaudeFiles(beforeEntries)

	adapter := &ClaudeAdapter{config: AdapterConfig{Name: "claude"}}
	_, _ = adapter.Format(context.Background(), "prompt", strings.NewReader(`{}`))

	afterEntries, _ := os.ReadDir(tmpDir)
	after := countGazeClaudeFiles(afterEntries)

	if after > before {
		t.Errorf("temp files leaked: before=%d after=%d", before, after)
	}
}

func countGazeClaudeFiles(entries []os.DirEntry) int {
	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "gaze-claude-prompt-") {
			count++
		}
	}
	return count
}
