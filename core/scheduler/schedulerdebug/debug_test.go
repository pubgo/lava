package schedulerdebug

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/core/scheduler"
)

func TestStatusCodeFromErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "job not found",
			err:  fmt.Errorf("wrapped: %w", scheduler.ErrJobNotFound),
			want: fiber.StatusNotFound,
		},
		{
			name: "job already exists",
			err:  fmt.Errorf("wrapped: %w", scheduler.ErrJobAlreadyExists),
			want: fiber.StatusConflict,
		},
		{
			name: "job name empty",
			err:  fmt.Errorf("wrapped: %w", scheduler.ErrJobNameEmpty),
			want: fiber.StatusBadRequest,
		},
		{
			name: "unknown error",
			err:  fmt.Errorf("unknown"),
			want: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusCodeFromErr(tt.err); got != tt.want {
				t.Fatalf("statusCodeFromErr() = %d, want %d", got, tt.want)
			}
		})
	}
}
