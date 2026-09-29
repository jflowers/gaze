package adapter

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestSessionDiscoverTimeoutFailsFast(t *testing.T) {
	bin := buildCallTestFakeAnalyzer(t)

	var stderr bytes.Buffer
	session := NewSession(bin, []string{"--stdio", "--hang-discover"}, "/tmp/project", []string{"./..."}, &stderr, nil)
	session.discoverTimeout = 50 * time.Millisecond

	_, err := session.Initialize()
	if err == nil {
		t.Fatalf("Initialize() = nil error, want error when discover times out")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Initialize() error = %v, want to wrap context.DeadlineExceeded", err)
	}
	_ = session.Close()
}
