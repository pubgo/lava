package registry

var defaultRegistry Registry

// SetDefault 设置全局默认 Registry 实例，不可为 nil。
func SetDefault(r Registry) {
	if r == nil {
		panic("[r] is nil")
	}
	defaultRegistry = r
}

// Default 返回全局默认 Registry 实例，未初始化时 panic。
func Default() Registry {
	if defaultRegistry == nil {
		panic("please init defaultRegistry")
	}
	return defaultRegistry
}
