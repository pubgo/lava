package resource

import (
	"github.com/pubgo/lava/internal/logz"
	"runtime"
	"strings"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/typex"
)

const Name = "resource"

var sources typex.SMap

// Remove 删除资源
func Remove(kind string, name string) {
	logz.Named(Name).Infof("delete resource, kind=>%s, name=>%s", kind, name)
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
		name = consts.Default
	}

	kind := srv.Kind()
	check(kind, name)

	var id = join(kind, name)
	var oldClient, ok = sources.Load(id)

	// 资源存在, 更新老资源
	if ok && oldClient != nil {
		logz.Named(Name).Infof("update resource[%s,%s]", kind, name)
		oldClient.(Resource).UpdateResObj(srv)
		return
	}

	// 资源不存在, 创建新资源
	logz.Named(Name).Infof("create resource[%s,%s]", kind, name)

	sources.Set(id, srv)

	// 依赖注入
	xerror.Panic(dix.ProviderNs(name, srv))

	// 当resource被gc时, 关闭resource
	runtime.SetFinalizer(srv, func(cc Resource) {
		defer xerror.Resp(func(err xerror.XErr) {
			logz.Named(Name).Error("old resource close error",
				zap.Any("name", name),
				zap.Any("err", err),
				zap.Any("err_msg", err.Error()))
		})

		xerror.Panic(cc.Close())
		logz.Named(Name).Infof("old resource[%s,%p] close ok", name, cc)
	})
}

// Get 根据类型和名字获取一个资源
func Get(kind string, name string) Resource {
	check(kind, name)
	var id = join(kind, name)
	if val, ok := sources.Load(id); ok {
		return val.(Resource)
	}
	return nil
}

// GetByKind 通过资源类型获取资源列表
func GetByKind(kind string) map[string]Resource {
	check(kind, "check")
	var ss = make(map[string]Resource)
	sources.Range(func(key, val interface{}) bool {
		var name = key.(string)
		if val.(Resource).Kind() == kind {
			ss[name] = val.(Resource)
		}
		return true
	})
	return ss
}

// GetOne 根据类型获取一个资源
func GetOne(kind string) Resource {
	check(kind, "check")
	var ss Resource
	sources.Range(func(_, val interface{}) bool {
		if val.(Resource).Kind() == kind {
			ss = val.(Resource)
			return false
		}
		return true
	})
	return ss
}

// GetAllKind 获取所有的资源类型
func GetAllKind() []string {
	var ss []string
	var set = make(map[string]struct{})
	sources.Range(func(_, val interface{}) bool {
		set[val.(Resource).Kind()] = struct{}{}
		return true
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
		xerror.Panic(ErrKindNull, "kind:", kind, "name:", name)
	}
}
