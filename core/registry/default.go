package registry

var defaultRegistry Registry

func SetDefault(r Registry) {
	if r == nil {
		panic("[r] is nil")
	}
	defaultRegistry = r
}

func Default() Registry {
	if defaultRegistry == nil {
		panic("please init defaultRegistry")
	}
	return defaultRegistry
}
