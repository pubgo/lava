package resource

// Get 根据类型和名字获取一个资源
func Get(kind string, name string) Resource {
	check(kind, name)
	if val, ok := resourceList.Load(join(kind, name)); ok {
		return val.(Resource)
	}
	return nil
}

// GetByKind 通过资源类型获取资源列表
func GetByKind(kind string) map[string]Resource {
	check(kind, "check")
	var ss = make(map[string]Resource)
	resourceList.Range(func(name string, val interface{}) bool {
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
	resourceList.Range(func(_ string, val interface{}) bool {
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
	resourceList.Range(func(_ string, val interface{}) bool {
		set[val.(Resource).Kind()] = struct{}{}
		return true
	})

	for k := range set {
		ss = append(ss, k)
	}
	return ss
}
