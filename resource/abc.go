package resource

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/reflectx"
	"github.com/pubgo/lava/pkg/syncx"
)

var _ BuilderFactory = (*Factory)(nil)

type Factory struct {
	ResID      string                                                          `json:"_id" yaml:"_id"`
	OnBuilder  Builder                                                         `json:"-" yaml:"-"`
	OnResource Resource                                                        `json:"-" yaml:"-"`
	OnDi       func(obj inject.Object, field inject.Field) (interface{}, bool) `json:"-" yaml:"-"`
}

func (t Factory) Di(kind string) func(obj inject.Object, field inject.Field) (interface{}, bool) {
	if t.OnDi == nil {
		return defaultDi(kind)
	}
	return t.OnDi
}

func (t Factory) Wrapper(res Resource) Resource {
	if t.OnResource == nil {
		return res
	}

	var v = reflectx.Indirect(reflect.New(reflectx.Indirect(reflect.ValueOf(t.OnResource)).Type()))
	// TODO Resource
	var v1 = v.FieldByName("Resource")
	if !v1.IsValid() {
		panic(fmt.Sprintf("resource: %#v, has not field(Resource)", t.OnResource))
	}
	v1.Set(reflect.ValueOf(res))
	return v1.Interface().(Resource)
}

func (t Factory) GetResId() string {
	if t.ResID == "" {
		return consts.KeyDefault
	}
	return t.ResID
}

func (t Factory) Builder() Builder { return t.OnBuilder }

type Builder interface {
	Build() io.Closer
}

type BuilderFactory interface {
	Builder() Builder
	GetResId() string
	Wrapper(res Resource) Resource
	Di(kind string) func(obj inject.Object, field inject.Field) (interface{}, bool)
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

	// GetRes 获取真实的资源对象
	//	r: 用完对象后记得 release, 不release会死锁
	GetRes() interface{}

	// Done 资源释放
	Done()

	Log() *zap.Logger
}

func newRes(name string, kind string, val io.Closer) Resource {
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
	cfg     Builder
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
			t.log.Error("resource update fail", logutil.ErrField(err, zap.Uint32("curConn", t.counter.Load()))...)
		},
	)
}
