// Package testhelper provides polling helpers for tunnel integration tests.
package testhelper

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const pollInterval = 25 * time.Millisecond

// WaitTCP dials addr until the listener accepts or ctx is cancelled.
func WaitTCP(ctx context.Context, addr string) error {
	var d net.Dialer
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait tcp %s: %w", addr, ctx.Err())
		case <-ticker.C:
		}
	}
}

// WaitHTTPGET polls url until GET returns nil error or ctx is cancelled.
func WaitHTTPGET(ctx context.Context, url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait http get %s: %w", url, ctx.Err())
		case <-ticker.C:
		}
	}
}

// WaitServiceCount polls countFn until it returns at least min or ctx is cancelled.
func WaitServiceCount(ctx context.Context, min int, countFn func() int) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if countFn() >= min {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait service count >= %d (got %d): %w", min, countFn(), ctx.Err())
		case <-ticker.C:
		}
	}
}
