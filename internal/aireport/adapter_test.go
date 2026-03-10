package aireport

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

// TestFakeAdapter_ReturnsConfiguredResponse verifies that FakeAdapter returns
// the configured Response string and nil error.
func TestFakeAdapter_ReturnsConfiguredResponse(t *testing.T) {
	fa := &FakeAdapter{Response: "# Report\n\nSome content."}
	got, err := fa.Format(context.Background(), "system", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != fa.Response {
		t.Errorf("expected %q, got %q", fa.Response, got)
	}
}

// TestFakeAdapter_ReturnsConfiguredError verifies that FakeAdapter returns the
// configured Err and an empty string when Err is non-nil.
func TestFakeAdapter_ReturnsConfiguredError(t *testing.T) {
	want := errors.New("fake error")
	fa := &FakeAdapter{Err: want, Response: "should be ignored"}
	got, err := fa.Format(context.Background(), "system", strings.NewReader(`{}`))
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
	if got != "" {
		t.Errorf("expected empty string on error, got %q", got)
	}
}

// TestFakeAdapter_CallsGrowWithEachInvocation verifies that the Calls slice
// grows with each Format call and that call fields are recorded correctly.
func TestFakeAdapter_CallsGrowWithEachInvocation(t *testing.T) {
	fa := &FakeAdapter{Response: "ok"}

	payload1 := `{"call":"one"}`
	payload2 := `{"call":"two"}`

	_, _ = fa.Format(context.Background(), "prompt1", strings.NewReader(payload1))
	_, _ = fa.Format(context.Background(), "prompt2", strings.NewReader(payload2))

	if len(fa.Calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(fa.Calls))
	}
	if fa.Calls[0].SystemPrompt != "prompt1" {
		t.Errorf("expected SystemPrompt %q, got %q", "prompt1", fa.Calls[0].SystemPrompt)
	}
	if string(fa.Calls[0].Payload) != payload1 {
		t.Errorf("expected Payload %q, got %q", payload1, fa.Calls[0].Payload)
	}
	if fa.Calls[1].SystemPrompt != "prompt2" {
		t.Errorf("expected SystemPrompt %q, got %q", "prompt2", fa.Calls[1].SystemPrompt)
	}
}

// TestFakeAdapter_RecordsFullPayloadBytes verifies that FakeAdapterCall
// captures the complete payload bytes read from the io.Reader.
func TestFakeAdapter_RecordsFullPayloadBytes(t *testing.T) {
	payload := bytes.Repeat([]byte("x"), 4096)
	fa := &FakeAdapter{Response: "ok"}
	_, _ = fa.Format(context.Background(), "sys", bytes.NewReader(payload))

	if len(fa.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fa.Calls))
	}
	if !bytes.Equal(fa.Calls[0].Payload, payload) {
		t.Errorf("payload mismatch: got %d bytes, want %d", len(fa.Calls[0].Payload), len(payload))
	}
}

// TestFakeAdapter_NilPayloadReader verifies that FakeAdapter handles a nil
// payload reader without panicking.
func TestFakeAdapter_NilPayloadReader(t *testing.T) {
	fa := &FakeAdapter{Response: "ok"}
	got, err := fa.Format(context.Background(), "sys", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Errorf("expected %q, got %q", "ok", got)
	}
	if fa.Calls[0].Payload != nil {
		t.Errorf("expected nil Payload for nil reader, got %v", fa.Calls[0].Payload)
	}
}

// TestNewAdapter_RejectsUnknownName verifies that NewAdapter returns a
// descriptive error for unknown adapter names.
func TestNewAdapter_RejectsUnknownName(t *testing.T) {
	_, err := NewAdapter(AdapterConfig{Name: "badai"})
	if err == nil {
		t.Fatal("expected error for unknown adapter name")
	}
	if !strings.Contains(err.Error(), "badai") {
		t.Errorf("expected error to mention adapter name, got: %v", err)
	}
	for _, valid := range []string{"claude", "gemini", "ollama"} {
		if !strings.Contains(err.Error(), valid) {
			t.Errorf("expected error to list valid adapter %q, got: %v", valid, err)
		}
	}
}

// TestNewAdapter_AcceptsAllValidNames verifies that NewAdapter returns a
// non-nil adapter for each valid name without error.
func TestNewAdapter_AcceptsAllValidNames(t *testing.T) {
	for _, name := range []string{"claude", "gemini", "ollama"} {
		adapter, err := NewAdapter(AdapterConfig{Name: name, Model: "m"})
		if err != nil {
			t.Errorf("NewAdapter(%q): unexpected error: %v", name, err)
		}
		if adapter == nil {
			t.Errorf("NewAdapter(%q): expected non-nil adapter", name)
		}
		_ = adapter
	}
}
