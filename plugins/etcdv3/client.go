package etcdv3

import (
	"go.etcd.io/etcd/client/v3"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/internal/resource"
)

// Get 获取etcd client
func Get(names ...string) *Client {
	var c = resource.Get(Name, consts.GetDefault(names...))
	if c != nil {
		return c.(*Client)
	}
	return nil
}

// Update 更新client
func Update(name string, cfg Cfg) {
	etcdCfg := cfgMerge(cfg)
	client := etcdCfg.Build()
	resource.Update(Name, name, &Client{client})
}

// Delete 删除etcd client
func Delete(name string) {
	resource.Remove(Name, name)
}

var _ resource.Resource = (*Client)(nil)

type Client struct {
	*clientv3.Client
}
