package resource

import (
	"io"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/syncx"
)

type Base struct {
	ResID     string                      `json:"_id" yaml:"_id"`
	OnCfg     CfgBuilder                  `json:"-" yaml:"-"`
	OnWrapper func(res Resource) Resource `json:"-" yaml:"-"`
}

func (t Base) Wrapper(res Resource) Resource {
	if t.OnWrapper == nil {
		return res
	}
	return t.OnWrapper(res)
}

func (t Base) GetResId() string {
	if t.ResID == "" {
		return consts.KeyDefault
	}
	return t.ResID
}

func (t Base) Cfg() CfgBuilder { return t.OnCfg }

type CfgBuilder interface {
	Build() io.Closer
}

type Builder interface {
	Cfg() CfgBuilder
	GetResId() string
	Wrapper(res Resource) Resource
}

// Release 释放资源
type Release interface {
	Release()
}

// Resource 资源对象接口
// 	Resource 是真实对象的wrapper, 可以通过getObj获取真实的内部对象
type Resource interface {
	// getObj 获取资源内部真实对象
	getObj() io.Closer

	// updateObj 更新资源对象
	updateObj(obj io.Closer)

	// Kind 资源类型
	Kind() string

	// Name 资源名字(ID)
	Name() string

	// LoadObj 获取真实的资源对象
	//	r: 用完对象后记得 release, 不release会死锁
	LoadObj() (obj interface{}, r Release)
}

func New(name string, kind string, val io.Closer) Resource {
	return &baseRes{name: name, kind: kind, v: val}
}

type baseRes struct {
	kind    string
	name    string
	v       io.Closer
	rw      sync.RWMutex
	cfg     CfgBuilder
	counter atomic.Uint32
}

func (t *baseRes) Name() string { return t.name }

func (t *baseRes) Kind() string { return t.kind }

func (t *baseRes) Release() {
	t.rw.RUnlock()
	t.counter.Dec()
}

func (t *baseRes) LoadObj() (interface{}, Release) {
	t.rw.RLock()
	t.counter.Inc()
	return t.v, t
}

func (t *baseRes) getObj() io.Closer {
	t.rw.RLock()
	var v = t.v
	t.rw.RUnlock()
	return v
}

func (t *baseRes) updateObj(obj io.Closer) {
	// 如果长久未释放，就log error
	syncx.GoMonitor(
		time.Second*5,
		func() bool { return t.counter.Load() != 0 },
		func(err error) {
			logs.L().Error("resource update timeout",
				logutil.ErrField(err, zap.String("kind", t.kind), zap.String("name", t.name))...)
		},
	)

	t.rw.Lock()
	t.v = obj
	t.rw.Unlock()
}
