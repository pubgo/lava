package etcdv3

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/logutil"
	"github.com/pubgo/lug/pkg/typex"

	"github.com/pubgo/xerror"

	"runtime"
)

var clients typex.SMap

// Get 获取etcd client
func Get(names ...string) *Client {
	c := clients.Get(consts.GetDefault(names...))
	if c == nil {
		return nil
	}

	return c.(*Client)
}

// Update 更新etcd client
func Update(name string, cfg Cfg) (gErr error) {
	defer xerror.RespErr(&gErr)

	xerror.Assert(name == "", "[name] should not be null")

	// 创建新的客户端
	etcdClient, err := cfg.Build()
	xerror.Panic(err)

	// 获取老的客户端
	oldClient, ok := clients.Load(name)
	if !ok || oldClient == nil {
		logs.Debug("create client", logutil.Name(name))

		// 老客户端不存在就直接保存
		clients.Set(name, &Client{etcdClient})
		return nil
	}

	// 当old etcd client没有被使用的时候, 那么就关闭
	runtime.SetFinalizer(oldClient, func(cc *Client) {
		logs.Infof("old client gc", logutil.Name(name), logutil.UIntPrt(cc))
		if err := cc.Close(); err != nil {
			logs.Error("old client close error", logutil.Name(name), logutil.Err(err))
		}
	})

	logs.Debug("update client", logutil.Name(name))
	// 老的客户端更新
	clients.Set(name, &Client{etcdClient})
	return nil
}

// Delete 删除etcd client, 并关闭etcd client
func Delete(name string) {
	clients.Delete(name)
}

// Each etcd client list
func Each(fn func(key string)) {
	xerror.Panic(clients.Each(fn))
}
