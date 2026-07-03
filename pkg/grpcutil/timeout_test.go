package grpcutil

import (
	"testing"
	"time"
)

func TestCapRequestTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   time.Duration
		want time.Duration
	}{
		{0, MinRequestTimeout},
		{time.Millisecond, MinRequestTimeout},
		{time.Second, time.Second},
		{time.Hour, MaxRequestTimeout},
	}

	for _, tt := range tests {
		if got := CapRequestTimeout(tt.in); got != tt.want {
			t.Fatalf("CapRequestTimeout(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
