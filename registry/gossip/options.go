package gossip

import (
	"context"
	"time"

	"github.com/hashicorp/memberlist"
	registry "github.com/pubgo/lug/registry"
)

type secretKey struct{}
type addressKey struct{}
type configKey struct{}
type advertiseKey struct{}
type connectTimeoutKey struct{}
type connectRetryKey struct{}

// helper for setting registry options
func setRegistryOption(k, v interface{}) registry.Opt {
	return func(o *registry.Opts) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, k, v)
	}
}

// Secret specifies an encryption key. The value should be either
// 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func Secret(k []byte) registry.Opt {
	return setRegistryOption(secretKey{}, k)
}

// Address to bind to - host:port
func Address(a string) registry.Opt {
	return setRegistryOption(addressKey{}, a)
}

// Config sets *memberlist.Config for configuring gossip
func Config(c *memberlist.Config) registry.Opt {
	return setRegistryOption(configKey{}, c)
}

// The address to advertise for other gossip members to connect to - host:port
func Advertise(a string) registry.Opt {
	return setRegistryOption(advertiseKey{}, a)
}

// ConnectTimeout sets the registry connect timeout. Use -1 to specify infinite timeout
func ConnectTimeout(td time.Duration) registry.Opt {
	return setRegistryOption(connectTimeoutKey{}, td)
}

// ConnectRetry enables reconnect to registry then connection closed,
// use with ConnectTimeout to specify how long retry
func ConnectRetry(v bool) registry.Opt {
	return setRegistryOption(connectRetryKey{}, v)
}
