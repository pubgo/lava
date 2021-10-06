package request_id

import (
	"context"

	"github.com/segmentio/ksuid"
)

type (
	reqIdKey struct{}
)

func ctxWithReqId(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, reqIdKey{}, val)
}

func GetReqID(ctx context.Context) string {
	var reqId, ok = ctx.Value(reqIdKey{}).(string)
	if ok {
		return reqId
	}
	return ksuid.New().String()
}
