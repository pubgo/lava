package resource

import "github.com/pubgo/lava/resource/resource_type"

// Get 根据类型和名字获取一个资源
func Get(kind string, name string) resource_type.Resource {
	check(kind, name)
	if val, ok := resourceList.Load(join(kind, name)); ok {
		return val.(resource_type.Resource)
	}
	return nil
}

// GetByKind 通过资源类型获取资源列表
func GetByKind(kind string) map[string]resource_type.Resource {
	check(kind, "check")
	var ss = make(map[string]resource_type.Resource)
	resourceList.Range(func(name string, val interface{}) bool {
		if val.(resource_type.Resource).Kind() == kind {
			ss[name] = val.(resource_type.Resource)
		}
		return true
	})
	return ss
}

// GetOne 根据类型获取一个资源
func GetOne(kind string) resource_type.Resource {
	check(kind, "check")
	var ss resource_type.Resource
	resourceList.Range(func(_ string, val interface{}) bool {
		if val.(resource_type.Resource).Kind() == kind {
			ss = val.(resource_type.Resource)
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
	resourceList.Range(func(_ string, val interface{}) bool {
		set[val.(resource_type.Resource).Kind()] = struct{}{}
		return true
	})

	for k := range set {
		ss = append(ss, k)
	}
	return ss
}
