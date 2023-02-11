package requestid

import (
	"context"

	"github.com/rs/xid"
)

var reqIdKey = xid.New().String()

func CreateCtx(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, reqIdKey, reqId)
}

func Ctx(ctx context.Context) string {
	var reqId, ok = ctx.Value(reqIdKey).(string)
	if ok {
		return reqId
	}

	return xid.New().String()
}

func getReqID(ctx context.Context) string {
	var reqId, ok = ctx.Value(reqIdKey).(string)
	if ok {
		return reqId
	}

	return ""
}
