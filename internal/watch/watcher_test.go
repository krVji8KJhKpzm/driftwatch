package watch_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/driftwatch/internal/watch"
)

// mockRunner records calls for assertion.
type mockRunner struct {
	called int
	errAfter int
	returnErr error
}

func (m *mockRunner) Run(ctx context.Context, cfg watch.Config) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(cfg.Interval):
			m.called++
			if m.errAfter > 0 && m.called >= m.errAfter {
				return m.returnErr
			}
		}
	}
}

func TestWatch_CancelStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	r := &mockRunner{}
	cfg := watch.Config{
		ManifestPath: "manifests.yaml",
		Interval:     10 * time.Millisecond,
		Format:       "text",
	}

	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx, cfg)
	}()

	time.Sleep(35 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil after cancel, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run did not return after context cancel")
	}

	if r.called < 2 {
		t.Errorf("expected at least 2 ticks, got %d", r.called)
	}
}

func TestWatch_PropagatesError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("boom")

	r := &mockRunner{errAfter: 1, returnErr: wantErr}
	cfg := watch.Config{
		Interval: 5 * time.Millisecond,
		Format:   "text",
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	err := r.Run(ctxTimeout, cfg)
	if !errors.Is(err, wantErr) {
		t.Errorf("expected %v, got %v", wantErr, err)
	}
}
