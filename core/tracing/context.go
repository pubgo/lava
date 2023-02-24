package tracing

import (
	"context"

	"github.com/pubgo/lava/internal/requestid"
)

func RequestID(ctx context.Context) string {
	return requestid.Ctx(ctx)
}
