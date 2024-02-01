package encoding

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/typex"
)

var data typex.Map[Codec]

func Register(name string, cdc Codec) {
	defer recovery.Exit()
	assert.If(cdc == nil || name == "" || cdc.Name() == "", "codec[%s] is null", name)
	assert.If(data.Has(name), "[cdc] %s already exists", name)
	data.Set(name, cdc)
}

func Get(name string) Codec {
	val, ok := data.Load(name)
	if !ok {
		return nil
	}

	return val
}

func Keys() []string { return data.Keys() }

func Each(fn func(name string, cdc Codec)) {
	defer recovery.Exit()

	data.Each(func(name string, val Codec) {
		fn(name, val)
	})
}

// GetWithCT get codec with content type
func GetWithCT(ct string) Codec {
	return Get(cdcMapping[ct])
}

var cdcMapping = map[string]string{
	"application/json":         "json",
	"application/proto":        "proto",
	"application/protobuf":     "proto",
	"application/octet-stream": "proto",
	"application/grpc":         "proto",
	"application/grpc+json":    "json",
	"application/grpc+proto":   "proto",
	"application/grpc+bytes":   "bytes",
}
