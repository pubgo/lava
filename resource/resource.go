package resource

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/typex"
)

const Name = "resource"

var sources typex.RwMap
var logs = logging.Component(Name)
var mu sync.Mutex

// Remove 删除资源
func Remove(kind string, name string) {
	logs.S().Infow("resource delete", "kind", kind, "name", name)
	check(kind, name)
	sources.Del(join(kind, name))
}

// Has 检查资源是否存在
func Has(kind string, name string) bool {
	check(kind, name)
	return sources.Has(join(kind, name))
}

// Update 更新资源
func Update(name string, srv Resource) {
	xerror.Assert(srv == nil || srv.GetObj() == nil, "[srv] should not be nil")

	if name == "" {
		name = consts.KeyDefault
	}

	kind := srv.Kind()
	check(kind, name)

	var fields = []zap.Field{
		zap.String("kind", kind),
		zap.String("name", name),
		zap.String("resource", fmt.Sprintf("%#v", srv)),
	}

	var log = logs.L().With(fields...)

	var id = join(kind, name)

	// TODO 防止资源竞争
	mu.Lock()
	defer mu.Unlock()

	var oldClient, ok = sources.Load(id)
	// 资源存在, 更新老资源
	if ok && oldClient != nil {
		// 新老对象替换, 资源内部对象不同时替换
		if oldClient.(Resource).GetObj() != srv.GetObj() {
			log.With(zap.String("old_resource", fmt.Sprintf("%#v", oldClient))).Info("resource update")
			oldClient.(Resource).updateObj(srv.GetObj())
		}
		return
	}

	// 资源不存在, 创建新资源
	log.Info("resource create")

	sources.Set(id, srv)

	// 只在资源创建的时候更新一次,依赖注入
	xerror.Panic(dix.ProviderNs(name, srv))

	log.Info("resource SetFinalizer")

	// 当resource被gc时, 关闭  resource
	runtime.SetFinalizer(srv.GetObj(), func(cc io.Closer) {
		logutil.OkOrPanic(logs.L(), "resource close", cc.Close, fields...)
	})
}

func join(names ...string) string {
	return strings.Join(names, "-")
}

func check(kind string, name string) {
	xerror.Assert(kind == "" || name == "", "resource: kind(%s) and name(%s) should not be null", kind, name)
}
