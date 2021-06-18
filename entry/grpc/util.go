package grpc

import (
	"encoding/json"
	"net/url"
	"os"
	"reflect"

	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/xgen"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

func register(server *grpc.Server, handler interface{}) error {
	xerror.Assert(server == nil, "[server] should not be nil")

	var v = checkHandle(handler)
	if v.IsValid() {
		_ = fx.WrapValue(v, server, handler)
		return nil
	}

	return xerror.Fmt("register [%#v] 没有找到匹配的interface", handler)
}

func getHostname() string {
	if name, err := os.Hostname(); err != nil {
		return "unknown"
	} else {
		return name
	}
}

func checkHandle(handler interface{}) reflect.Value {
	xerror.Assert(handler == nil, "[handler] should not be nil")

	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for v := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 2 || v1.In(1).Kind() != reflect.Interface {
			continue
		}

		if !hd.Implements(v1.In(1)) || v1.In(0).String() != "*grpc.Server" {
			continue
		}

		return v
	}

	return reflect.Value{}
}

func init() {
	encoding.RegisterCodec(&uriCodec{})
}

// 解析http get请求的query参数
type uriCodec struct{}

func (c *uriCodec) Name() string                          { return "uri" }
func (c *uriCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *uriCodec) Unmarshal(data []byte, v interface{}) error {
	var u, err = url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	return gutil.MapFormByTag(v, u, "json")
}
