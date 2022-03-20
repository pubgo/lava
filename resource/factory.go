package resource

import (
	"fmt"
	"github.com/pubgo/lava/logging/logutil"
	"go.uber.org/zap"
	"io"
	"reflect"
	"runtime"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/pkg/reflectx"
	"github.com/pubgo/lava/resource/resource_type"
)

var _ resource_type.BuilderFactory = (*Factory)(nil)

type Factory struct {
	ResID      string                                                          `json:"_id" yaml:"_id"`
	CfgBuilder resource_type.Builder                                           `json:"-" yaml:"-"`
	ResType    resource_type.Resource                                          `json:"-" yaml:"-"`
	OnDi       func(obj inject.Object, field inject.Field) (interface{}, bool) `json:"-" yaml:"-"`
}

func (f *Factory) Update(name, kind string) {
	if name == "" {
		name = consts.KeyDefault
	}

	check(kind, name)

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

		var builder = f.Builder()
		// 注入对象
		inject.Inject(builder)

		var srv = f.Wrapper(newRes(name, kind, builder.Build()))
		// 注入对象
		inject.Inject(srv)

		// 依赖注入
		inject.Register(srv, f.Di(kind))

		resourceList.Set(id, srv)

		log.Info("resource SetFinalizer")

		// 当resource被gc时, 关闭resource
		runtime.SetFinalizer(srv, func(cc io.Closer) {
			logutil.OkOrPanic(logs.L(), "resource close", cc.Close, fields...)
		})
		return
	}

	// 资源存在, 更新老资源
	// 新老对象替换, 资源内部对象不同时替换
	if oldSrv.(*baseRes).builder == f.Builder() || reflect.DeepEqual(oldSrv.(*baseRes).builder, f.Builder()) {
		// 配置未变更
		return
	}

	var srv = f.Builder().Build()
	var resSrv = f.Wrapper(newRes(name, kind, srv))
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
	if f.ResType == nil {
		return res
	}

	var v = reflectx.Indirect(reflect.New(reflectx.Indirect(reflect.ValueOf(f.ResType)).Type()))
	// find Resource field
	var v1 = v.FieldByName("Resource")
	if !v1.IsValid() {
		panic(fmt.Sprintf("resource: %#v, has not field(Resource)", f.ResType))
	}
	v1.Set(reflect.ValueOf(res))
	return v1.Interface().(resource_type.Resource)
}

func (f Factory) GetResId() string {
	if f.ResID == "" {
		return consts.KeyDefault
	}
	return f.ResID
}

func (f Factory) Builder() resource_type.Builder { return f.CfgBuilder }
