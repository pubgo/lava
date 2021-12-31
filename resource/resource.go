package resource

import (
	"runtime"
	"strings"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/pkg/typex"
)

const Name = "resource"

var sources typex.SMap
var logs = logz.Component(Name)

// Remove 删除资源
func Remove(kind string, name string) {
	logs.Infow("delete resource", "kind", kind, "name", name)
	check(kind, name)
	sources.Delete(join(kind, name))
}

// Has 检查资源是否存在
func Has(kind string, name string) bool {
	check(kind, name)
	return sources.Has(join(kind, name))
}

// Update 更新资源
func Update(name string, srv Resource) {
	xerror.Assert(srv == nil, "[srv] should not be nil")

	if name == "" {
		name = consts.KeyDefault
	}

	kind := srv.Kind()
	check(kind, name)

	var id = join(kind, name)
	var oldClient, ok = sources.Load(id)

	var log = logs.With(zap.String("kind", kind), zap.String("name", name))

	// 资源存在, 更新老资源
	if ok && oldClient != nil {
		log.Info("update resource")
		oldClient.(Resource).UpdateResObj(srv)
		return
	}

	// 资源不存在, 创建新资源
	log.Info("create resource")

	sources.Set(id, srv)

	// 只在资源创建的时候更新一次,依赖注入
	xerror.Panic(dix.ProviderNs(name, srv))

	// 当resource被gc时, 关闭resource
	runtime.SetFinalizer(srv, func(cc Resource) {
		logs.Logs("old resource close", cc.Close,
			zap.String("kind", kind),
			zap.String("name", name),
		)
	})
}

func join(names ...string) string {
	return strings.Join(names, "-")
}

func check(kind string, name string) {
	xerror.Assert(kind == "" || name == "", "resource: kind(%s) and name(%s) should not be null", kind, name)
}
