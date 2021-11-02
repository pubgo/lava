package requestID

import (
	"context"

	"github.com/segmentio/ksuid"
)

type (
	reqIdKey struct{}
)

func WithReqID(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, reqIdKey{}, val)
}

func GetWith(ctx context.Context) string {
	var reqId, ok = ctx.Value(reqIdKey{}).(string)
	if ok {
		return reqId
	}
	return ksuid.New().String()
}

func getReqID(ctx context.Context) string {
	var reqId, ok = ctx.Value(reqIdKey{}).(string)
	if ok {
		return reqId
	}
	return ""
}
