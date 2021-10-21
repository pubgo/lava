package ossc

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/xerror"
)

func Get(names ...string) *Client {
	val := resource.Get(Name, lavax.GetDefault(names...))
	if val == nil {
		return nil
	}

	return val.(*Client)
}

func Update(name string, cfg ClientCfg) (err error) {
	defer xerror.RespErr(&err)
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	xerror.Panic(err)
	resource.Update(name, &Client{client})
	return nil
}
