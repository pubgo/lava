package grpcs

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/stats"
)

var _ stats.Handler = (*statsHandler)(nil)

type GRPCStats struct {
	Duration time.Duration
	Method   string
	Failed   float64
	Success  float64
}

var consMutex sync.Mutex
var cons  = make(map[*stats.ConnTagInfo]string)

type connCtxKey struct{}

func getConnTagFromContext(ctx context.Context) (*stats.ConnTagInfo, bool) {
	tag, ok := ctx.Value(connCtxKey{}).(*stats.ConnTagInfo)
	return tag, ok
}

type statsHandler struct {
}

// TagRPC 为空.
func (s *statsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC 为空.
func (s *statsHandler) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {
	//rs := StatsFromContext(ctx)
	//if rs == nil {
	//	return
	//}
	//
	//switch t := rpcStats.(type) {
	//// case *stats.Begin:
	//// case *stats.InPayload:
	//// case *stats.InHeader:
	//// case *stats.InTrailer:
	//// case *stats.OutPayload:
	//// case *stats.OutHeader:
	//// case *stats.OutTrailer:
	//case *stats.End:
	//	rs.Duration = t.EndTime.Sub(t.BeginTime)
	//	if t.Error != nil {
	//		rs.Failed = 1
	//	} else {
	//		rs.Success = 1
	//	}
	//	c.reqCh <- rs
	//}
}

// TagConn 用来给连接打个标签，以此来标识连接(实在是找不出还有什么办法来标识连接).
// 这个标签是个指针，可保证每个连接唯一。
// 将该指针添加到上下文中去，键为 connCtxKey{}.
func (s *statsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return context.WithValue(ctx, connCtxKey{}, info)
}

// HandleConn 会在连接开始和结束时被调用，分别会输入不同的状态.
func (s *statsHandler) HandleConn(ctx context.Context, connStats stats.ConnStats) {
	tag, ok := getConnTagFromContext(ctx)
	if !ok {
		log.Fatal("can not get conn tag")
	}

	consMutex.Lock()
	defer consMutex.Unlock()

	switch connStats.(type) {
	case *stats.ConnBegin:
		cons[tag] = ""
		log.Printf("begin conn, tag = (%p)%#v, now connections = %d\n", tag, tag, len(cons))
	case *stats.ConnEnd:
		delete(cons, tag)
		log.Printf("end conn, tag = (%p)%#v, now connections = %d\n", tag, tag, len(cons))
	default:
		log.Printf("illegal ConnStats type\n")
	}
}
