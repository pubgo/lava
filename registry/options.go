package registry

import (
	"time"

	"github.com/pubgo/xerror"
)

func TTL(t string) RegisterOption {
	return func(o *RegisterOptions) {
		dur, err := time.ParseDuration(t)
		xerror.Panic(err)
		o.TTL = dur
	}
}

// Watch a service
func WatchService(name string) WatchOption {
	return func(o *WatchOptions) {
		o.Service = name
	}
}
