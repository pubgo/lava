package golug_codex_thrift

import (
	"github.com/pubgo/golug/golug_codec"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xlog"
)

func init() {
	if golug_env.Trace {
		xlog.Debug("init ok")
	}
}

var Name = "thrift"

func init() {
	golug_codec.Register(Name, ThriftCodec{})
}
