package etcdv3

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/internal/resource"
)

// Get 获取etcd client
func Get(names ...string) *Client {
	var c = resource.Get(Name, consts.GetDefault(names...))
	if c != nil {
		return c.(*Client)
	}
	return nil
}

// Update 更新etcd client
func Update(name string, cfg Cfg) {
	etcdClient, err := cfg.Build()
	xerror.Panic(err)
	resource.Update(Name, name, &Client{etcdClient})
}

// Delete 删除etcd client
func Delete(name string) {
	resource.Remove(Name, name)
}
