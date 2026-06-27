// Package discovery 提供服务发现的抽象接口，供 gRPC resolver 等客户端组件使用。
//
// Discovery 与 registry 互补：registry 负责「注册自身」，
// discovery 负责「发现其他服务实例」。
//
// 默认提供 noop 实现（不返回任何服务），适用于无需服务发现的场景。
package discovery

import (
	"context"
	"time"

	"github.com/pubgo/funk/v2/result"

	"github.com/pubgo/lava/v2/core/service"
)

type (
	WatchOpt func(*WatchOpts)
	GetOpt   func(*GetOpts)
)

// Discovery 是服务发现接口。
type Discovery interface {
	String() string
	Watch(ctx context.Context, srv string, opts ...WatchOpt) result.Result[Watcher]
	GetService(ctx context.Context, srv string, opts ...GetOpt) result.Result[[]*service.Service]
}

// Watcher 监听服务实例变更，Next 为阻塞调用。
type Watcher interface {
	Next() result.Result[*Result]
	Stop() error
}

// Result 是 Watcher.Next 返回的单次变更事件。
type Result struct {
	Action  EventType
	Service *service.Service
}

type WatchOpts struct {
	Service string
}

type GetOpts struct {
	Timeout time.Duration
}

// EventType 表示服务变更类型。
type EventType int32

const (
	EventType_UNKNOWN EventType = 0
	EventType_CREATE  EventType = 1
	EventType_UPDATE  EventType = 2
	EventType_DELETE  EventType = 3
)
