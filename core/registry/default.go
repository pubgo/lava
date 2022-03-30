package registry

import (
	"github.com/pubgo/lava/core/registry/registry_type"
)

var defaultRegistry registry_type.Registry

func SetDefault(r registry_type.Registry) {
	if r == nil {
		panic("[r] is nil")
	}
	defaultRegistry = r
}

func Default() registry_type.Registry {
	if defaultRegistry == nil {
		panic("please init defaultRegistry")
	}
	return defaultRegistry
}
