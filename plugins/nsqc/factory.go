package nsqc

import (
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/xerror"
)

func Get(names ...string) *Client {
	var name = lavax.GetDefault(names...)
	val := resource.Get(Name, name)
	if val == nil {
		return nil
	}
	return val.(*Client)
}

func Update(name string, cfg Cfg) error {
	return xerror.Try(func() {
		xerror.Assert(name == "", "[name] should not be null")

		// 创建新的客户端
		client, err := cfg.Build()
		xerror.Panic(err)
		resource.Update(name, client)
	})
}
