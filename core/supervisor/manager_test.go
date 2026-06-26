package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/lifecycle/lifecyclebuilder"
)

func TestStartServiceAfterNaturalExit(t *testing.T) {
	mgr := NewManager("test-supervisor", lifecyclebuilder.New(nil).Getter)

	err := mgr.Add(
		NewService("once-service", func(ctx context.Context) error {
			return nil
		}),
		WithRestartPolicy(RestartOnFailure),
		WithRestartDelay(5*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("add service: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := mgr.ServeBackground(ctx)
	t.Cleanup(func() {
		cancel()
		waitManagerDone(t, errCh, 2*time.Second)
	})

	waitForCondition(t, 2*time.Second, func() bool {
		info, getErr := mgr.GetServiceInfo("once-service")
		if getErr != nil {
			return false
		}
		return info.StartCount >= 1 && info.Status != StatusRunning
	}, "service should naturally exit and not stay in running status")

	if err = mgr.StartService("once-service"); err != nil {
		t.Fatalf("start service after natural exit: %v", err)
	}

	waitForCondition(t, 2*time.Second, func() bool {
		info, getErr := mgr.GetServiceInfo("once-service")
		if getErr != nil {
			return false
		}
		return info.StartCount >= 2
	}, "service should be startable again after natural exit")
}

func TestMaxRestartsAllowsConfiguredAttempts(t *testing.T) {
	mgr := NewManager("test-supervisor", lifecyclebuilder.New(nil).Getter)

	err := mgr.Add(
		NewService("flaky-service", func(ctx context.Context) error {
			return errors.New("boom")
		}),
		WithRestartPolicy(RestartAlways),
		WithMaxRestarts(1),
		WithRestartDelay(time.Millisecond),
		WithBackoff(time.Millisecond, 1),
		WithRestartWindow(time.Hour, 0),
	)
	if err != nil {
		t.Fatalf("add service: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := mgr.ServeBackground(ctx)
	t.Cleanup(func() {
		cancel()
		waitManagerDone(t, errCh, 2*time.Second)
	})

	waitForCondition(t, 2*time.Second, func() bool {
		info, getErr := mgr.GetServiceInfo("flaky-service")
		if getErr != nil {
			return false
		}
		return info.Failed
	}, "service should be marked failed after exceeding max restarts")

	info, err := mgr.GetServiceInfo("flaky-service")
	if err != nil {
		t.Fatalf("get service info: %v", err)
	}
	if info.StartCount != 2 {
		t.Fatalf("expected exactly 2 starts (initial + 1 restart), got %d", info.StartCount)
	}
}

func TestNormalizeServiceConfig(t *testing.T) {
	cfg := normalizeServiceConfig(ServiceConfig{
		RestartDelay:      0,
		MaxRestartDelay:   0,
		RestartWindow:     0,
		BackoffMultiplier: 0,
	})

	if cfg.RestartDelay <= 0 {
		t.Fatalf("RestartDelay should be normalized to > 0")
	}
	if cfg.MaxRestartDelay < cfg.RestartDelay {
		t.Fatalf("MaxRestartDelay should be >= RestartDelay")
	}
	if cfg.RestartWindow <= 0 {
		t.Fatalf("RestartWindow should be normalized to > 0")
	}
	if cfg.BackoffMultiplier < 1 {
		t.Fatalf("BackoffMultiplier should be normalized to >= 1")
	}
}

func TestManagerErrorSentinels(t *testing.T) {
	mgr := NewManager("test-supervisor", lifecyclebuilder.New(nil).Getter)

	if _, err := mgr.GetServiceInfo("missing"); !errors.Is(err, ErrServiceNotFound) {
		t.Fatalf("GetServiceInfo should wrap ErrServiceNotFound, got: %v", err)
	}

	if err := mgr.StopService("missing"); !errors.Is(err, ErrServiceNotFound) {
		t.Fatalf("StopService should wrap ErrServiceNotFound, got: %v", err)
	}

	err := mgr.Add(NewService("svc", func(ctx context.Context) error { return nil }))
	if err != nil {
		t.Fatalf("first add service failed: %v", err)
	}

	err = mgr.Add(NewService("svc", func(ctx context.Context) error { return nil }))
	if !errors.Is(err, ErrServiceAlreadyExists) {
		t.Fatalf("second add should wrap ErrServiceAlreadyExists, got: %v", err)
	}
}

func TestNormalizeServiceConfigInvalidValues(t *testing.T) {
	cfg := normalizeServiceConfig(ServiceConfig{
		RestartPolicy:       RestartPolicy(99),
		MaxRestarts:         -2,
		MaxRestartsInWindow: -3,
		RestartDelay:        -time.Second,
		MaxRestartDelay:     -time.Minute,
		RestartWindow:       -time.Minute,
		BackoffMultiplier:   0.5,
	})

	if cfg.RestartPolicy != RestartAlways {
		t.Fatalf("RestartPolicy should be normalized to RestartAlways, got: %v", cfg.RestartPolicy)
	}
	if cfg.MaxRestarts != 0 {
		t.Fatalf("MaxRestarts should be normalized to 0, got: %d", cfg.MaxRestarts)
	}
	if cfg.MaxRestartsInWindow != 0 {
		t.Fatalf("MaxRestartsInWindow should be normalized to 0, got: %d", cfg.MaxRestartsInWindow)
	}
	if cfg.BackoffMultiplier < 1 {
		t.Fatalf("BackoffMultiplier should be normalized to >= 1, got: %f", cfg.BackoffMultiplier)
	}
}

func TestWaitDelay(t *testing.T) {
	t.Run("canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if ok := waitDelay(ctx, 10*time.Millisecond); ok {
			t.Fatalf("waitDelay should return false for canceled context")
		}
	})

	t.Run("zero delay", func(t *testing.T) {
		if ok := waitDelay(context.Background(), 0); !ok {
			t.Fatalf("waitDelay should return true for active context with zero delay")
		}
	})
}

func waitForCondition(t *testing.T, timeout time.Duration, fn func() bool, msg string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timeout waiting condition: %s", msg)
}

func waitManagerDone(t *testing.T, errCh <-chan error, timeout time.Duration) {
	t.Helper()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("manager exited with error: %v", err)
		}
	case <-time.After(timeout):
		t.Fatalf("timeout waiting manager exit")
	}
}
