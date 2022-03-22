package resource_type

import (
	"io"

	"go.uber.org/zap"

	"github.com/pubgo/lava/inject"
)

type Builder interface {
	Build() io.Closer
}

type BuilderFactory interface {
	Builder() Builder
	Update(name, kind string, builder Builder)
	IsValid() bool
	Wrapper(res Resource) Resource
	Di(kind string) func(obj inject.Object, field inject.Field) (interface{}, bool)
}

// Resource 资源对象接口
// 	Resource 是真实对象的wrapper, 可以通过getObj获取真实的内部对象
type Resource interface {
	// Kind 资源类型
	Kind() string

	// Name 资源名字(ID)
	Name() string

	// GetRes 获取真实的资源对象
	// 用完对象后记得release, 不然会死锁
	GetRes() interface{}

	// Done 资源释放
	Done()

	Log() *zap.Logger
}
