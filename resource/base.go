package resource

import (
	"io"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/resource/resource_type"
)

func newRes(name string, kind string, val io.Closer) resource_type.Resource {
	return &baseRes{
		name: name,
		kind: kind,
		v:    val,
		log:  zap.L().Named(logkey.Component).Named(kind).With(zap.String("name", name)),
	}
}

type baseRes struct {
	kind    string
	name    string
	v       io.Closer
	rw      sync.RWMutex
	builder resource_type.Builder
	counter atomic.Uint32
	log     *zap.Logger
}

func (t *baseRes) Log() *zap.Logger { return t.log }

func (t *baseRes) GetRes() interface{} {
	t.rw.RLock()
	t.counter.Inc()
	// TODO counter check
	return t.v
}

func (t *baseRes) Done() {
	t.counter.Dec()
	t.rw.RUnlock()
}

func (t *baseRes) Name() string { return t.name }
func (t *baseRes) Kind() string { return t.kind }

func (t *baseRes) getObj() io.Closer {
	t.rw.RLock()
	var v = t.v
	t.rw.RUnlock()
	return v
}

func (t *baseRes) updateObj(obj io.Closer) {
	// 资源更新5s超时, 打印log
	// 方便log查看和监控
	syncx.Monitor(
		time.Second*5,
		func() {
			t.rw.Lock()
			t.v = obj
			t.rw.Unlock()
			t.log.Info("resource update ok")
		},
		func(err error) {
			t.log.Error("resource update failed", logutil.ErrField(err, zap.Uint32("curConn", t.counter.Load()))...)
		},
	)
}
