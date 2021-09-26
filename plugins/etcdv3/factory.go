package etcdv3

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/pkg/typex"

	"runtime"
)

var clients typex.SMap

// Get 获取etcd client
func Get(names ...string) *Client {
	c := clients.Get(consts.GetDefault(names...))
	if c != nil {
		return c.(*Client)
	}
	return nil
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
		logs.Debug("create client", logger.Name(name))

		// 老客户端不存在就直接保存
		var client = &Client{etcdClient}
		clients.Set(name, client)
		xerror.Exit(dix.Provider(map[string]interface{}{name: client}))
		return
	}

	// 当old etcd client没有被使用的时候, 那么就关闭
	runtime.SetFinalizer(oldClient, func(cc *Client) {
		logs.Info("old client gc", logger.Name(name), logger.UIntPrt(cc))
		if err := cc.Close(); err != nil {
			logs.Error("old client close error", logger.Name(name), logger.Err(err))
		}
	})

	// 老的客户端更新
	logs.Debug("update client", logger.Name(name))
	oldClient.(*Client).Client = etcdClient
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
