package resource

import (
	"io"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/resource/resource_type"
)

func newRes(name string, kind string, val io.Closer) resource_type.Resource {
	var res = &baseRes{
		name: name,
		kind: kind,
		v:    val,
		log:  zap.L().Named(logkey.Component).Named(kind).With(zap.String("name", name)),
	}

	go res.loop()
	return res
}

type baseRes struct {
	kind    string
	name    string
	v       io.Closer
	rw      sync.RWMutex
	counter atomic.Uint32
	log     *zap.Logger
}

func (t *baseRes) loop() {

}

func (t *baseRes) Log() *zap.Logger { return t.log }

func (t *baseRes) GetRes() interface{} {
	t.rw.RLock()
	t.counter.Inc()

	go func() {
		if t.counter.Load() > 10 {
			t.log.Error("curConn should be release", zap.Uint32("curConn", t.counter.Load()))
		}
	}()
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
	go syncx.Monitor(
		time.Second*5,
		func() {
			t.rw.Lock()
			var oldPbj = t.v
			t.v = obj
			logutil.OkOrErr(t.log, "resource close", oldPbj.Close)
			t.rw.Unlock()
			t.log.Info("resource update ok")
		},
		func(err error) {
			t.log.Error("resource update failed", logutil.ErrField(err, zap.Uint32("curConn", t.counter.Load()))...)
		},
	)
}
