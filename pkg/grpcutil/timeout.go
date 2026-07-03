package grpcutil

import "time"

const (
	// MinRequestTimeout is the shortest deadline accepted from client metadata.
	MinRequestTimeout = 10 * time.Millisecond
	// MaxRequestTimeout is the longest deadline accepted from client metadata.
	MaxRequestTimeout = 30 * time.Second
)

// CapRequestTimeout clamps client-provided RPC deadlines to a safe range.
func CapRequestTimeout(d time.Duration) time.Duration {
	if d < MinRequestTimeout {
		return MinRequestTimeout
	}
	if d > MaxRequestTimeout {
		return MaxRequestTimeout
	}
	return d
}
