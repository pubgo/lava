package discovery

import (
	"context"
	"errors"
	"testing"
)

func TestNoopDiscoveryGetServiceReturnsEmpty(t *testing.T) {
	d := NewNoopDiscovery()
	r := d.GetService(context.Background(), "any-service")
	if r.IsErr() {
		t.Fatalf("noop GetService should succeed: %v", r.GetErr())
	}
	if svcs := r.Expect("get services"); len(svcs) > 0 {
		t.Fatalf("expected empty service list, got %v", svcs)
	}
}

func TestNoopDiscoveryWatchReturnsSelf(t *testing.T) {
	d := NewNoopDiscovery()
	w := d.Watch(context.Background(), "any-service").Expect("watch")
	nw, ok := w.(*noopDiscovery)
	if !ok || nw != d.(*noopDiscovery) {
		t.Fatalf("noop Watch should return the discovery itself as watcher")
	}
}

func TestNoopDiscoveryNextReturnsStopped(t *testing.T) {
	d := NewNoopDiscovery().(*noopDiscovery)
	err := d.Next().GetErr()
	if !errors.Is(err, ErrWatcherStopped) {
		t.Fatalf("expected ErrWatcherStopped, got %v", err)
	}
}

func TestNoopDiscoveryStop(t *testing.T) {
	d := NewNoopDiscovery().(*noopDiscovery)
	if err := d.Stop(); err != nil {
		t.Fatalf("Stop should not error: %v", err)
	}
}
