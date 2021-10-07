package resource

import (
	"runtime"
	"strings"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lug/pkg/typex"
)

var sources typex.RwMap

func Remove(kind string, name string) {
	check(kind, name)
	sources.Del(join(kind, name))
}

func Has(kind string, name string) bool {
	check(kind, name)
	return sources.Has(join(kind, name))
}

func Update(kind string, name string, srv Resource) {
	defer xerror.Raise(Update)

	check(kind, name)
	xerror.Assert(srv == nil, "[srv] should not be nil")

	var id = join(kind, name)
	var oldClient, ok = sources.Load(id)
	if ok && oldClient != nil {
		// 老客户端更新
		zap.S().Infof("update old client, name=>%s", name)
		oldClient.(*resourceWrap).srv = srv
		return
	}

	zap.S().Infof("create new client, name=>%s", name)

	// 创建新客户端
	var newClient = &resourceWrap{kind: kind, srv: srv}
	sources.Set(id, newClient)

	// 依赖注入
	xerror.Panic(dix.Provider(map[string]interface{}{name: srv}))

	// 当client被gc时, 关闭client
	runtime.SetFinalizer(newClient, func(cc Resource) {
		zap.S().Infof("old client gc, name=>%s, id=>%p", name, cc)
		if err := cc.Close(); err != nil {
			zap.S().Error("old client close error",
				zap.Any("name", name),
				zap.Any("err", err),
				zap.Any("err_msg", err.Error()))
		}
	})
}

func Get(kind string, name string) Resource {
	check(kind, name)
	var id = join(kind, name)
	if val, ok := sources.Load(id); ok {
		return val.(*resourceWrap).srv
	}
	return nil
}

func GetByKind(kind string) []Resource {
	check(kind, "check")
	var ss []Resource
	sources.Each(func(name string, val interface{}) {
		if val.(*resourceWrap).kind == kind {
			ss = append(ss, val.(*resourceWrap).srv)
		}
	})
	return ss
}

func GetAllKind() []string {
	var ss []string
	var set = make(map[string]struct{})
	sources.Each(func(name string, val interface{}) {
		set[val.(*resourceWrap).kind] = struct{}{}
	})

	for k := range set {
		ss = append(ss, k)
	}
	return ss
}

func join(names ...string) string {
	return strings.Join(names, "-")
}

func check(kind string, name string) {
	if kind == "" || name == "" {
		panic(xerror.Fmt("kind or name is null, [%s,%s]", kind, name))
	}
}
