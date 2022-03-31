package resource

import (
	"fmt"
	"reflect"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/pkg/reflectx"
	"github.com/pubgo/lava/resource/resource_type"
)

var _ resource_type.BuilderFactory = (*Factory)(nil)

type Factory struct {
	DefaultCfg resource_type.Builder                                           `json:"-" yaml:"-" inject:"required"`
	ResType    resource_type.Resource                                          `json:"-" yaml:"-"`
	OnDi       func(obj inject.Object, field inject.Field) (interface{}, bool) `json:"-" yaml:"-"`
}

func (f Factory) IsValid() bool { return f.DefaultCfg != nil }

func (f Factory) Update(name, kind string, builder resource_type.Builder) {
	if name == "" {
		name = consts.KeyDefault
	}

	check(kind, name)

	// 注入对象
	inject.Inject(builder)

	var fields = []zap.Field{
		zap.String("kind", kind),
		zap.String("name", name),
	}

	var log = logs.L().With(fields...)

	var id = join(kind, name)

	mu.Lock()
	defer mu.Unlock()

	oldSrv, ok := resourceList.Load(id)
	if !ok {
		// 资源不存在, 创建资源
		log.Info("resource create")

		var srv = f.Wrapper(newRes(name, kind, builder.Build()))
		// 注入对象
		inject.Inject(srv)
		log.Info("resource inject", zap.String("resource", fmt.Sprintf("%#v", srv)))

		// 注册对象
		inject.Register(srv, f.Di(kind))

		resourceList.Set(id, srv)
		return
	}

	var resSrv = f.Wrapper(newRes(name, kind, builder.Build()))
	oldSrv.(*baseRes).updateObj(resSrv.(*baseRes).getObj())

	log.With(zap.String("old_resource", fmt.Sprintf("%#v", oldSrv))).Info("resource update")
}

func (f Factory) Di(kind string) func(obj inject.Object, field inject.Field) (interface{}, bool) {
	if f.OnDi == nil {
		return defaultDi(kind)
	}
	return f.OnDi
}

func (f Factory) Wrapper(res resource_type.Resource) resource_type.Resource {
	xerror.Assert(f.ResType == nil, "please set [ResType], kind=%s name=%s", res.Kind(), res.Name())

	var obj = reflect.New(reflectx.Indirect(reflect.ValueOf(f.ResType)).Type())

	var v = reflectx.Indirect(obj)
	// find Resource field
	var field = reflectx.FindFieldBy(v, func(field reflect.StructField) bool {
		return field.Type.String() == resourceType.String()
	})
	xerror.Assert(!field.IsValid(), "resource has not field(Resource), resType=%#v, ", f.ResType)
	field.Set(reflect.ValueOf(res))

	return obj.Interface().(resource_type.Resource)
}

func (f Factory) Builder() resource_type.Builder { return f.DefaultCfg }
