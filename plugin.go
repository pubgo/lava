package lug

import (
	"github.com/pubgo/dix"
	_ "github.com/pubgo/lug/debug"
	_ "github.com/pubgo/lug/encoding/json"
	_ "github.com/pubgo/lug/encoding/protobuf"
	_ "github.com/pubgo/lug/healthy"
	_ "github.com/pubgo/lug/internal/log"
	_ "github.com/pubgo/lug/metric"
	_ "github.com/pubgo/lug/tracing"
	"github.com/pubgo/lug/vars"
)

func init() {
	vars.Watch("dix", func() interface{} { return dix.Json() })
}
