package etcd

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/x/xutil"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.etcd.io/etcd/clientv3"

	"runtime"
	"unsafe"
)

var clients types.SMap

// Get 获取etcd client
func Get(names ...string) *Client {
	c := clients.Get(consts.GetDefault(names...))
	if c == nil {
		return nil
	}
	return c.(*Client)
}

// List etcd client list
func List() (dt map[string]*Client) {
	xerror.Panic(clients.Map(&dt))
	return
}

func newClient(cfg clientv3.Config) (c *clientv3.Client, err error) {
	defer xerror.RespErr(&err)

	// 创建etcd client对象
	var etcdClient *clientv3.Client
	err = retry(3, func() error { etcdClient, err = clientv3.New(cfg); return err })
	xerror.PanicF(err, "[etcd] newClient error, err: %v, cfgList: %#v", err, cfg)

	return etcdClient, nil
}

// updateClient 更新etcd client
func updateClient(name string, cfg Cfg) error {
	log.Debugf("[etcd] %s update etcd client", name)

	// 创建新的客户端
	etcdClient, err := newClient(cfg.ToEtcd())
	if err != nil {
		return err
	}

	// 获取老的客户端
	oldClient, ok := clients.Load(name)
	if !ok || oldClient == nil {
		// 老客户端不存在就直接保存
		clients.Set(name, &Client{etcdClient})
		return nil
	}

	// 老的客户端存在就更新
	oldClient.(*Client).Client = etcdClient
	return nil
}

// initClient 创建或者初始化etcd client
func initClient(name string, cfg Cfg) error {
	return xutil.Try(func() {
		xerror.Assert(name == "", "[name] should not be null")
		xerror.Assert(clients.Has(name), "[etcd] %s already exists", name)

		etcdClient, err := newClient(cfg.ToEtcd())
		xerror.Panic(err)

		clients.Set(name, &Client{Client: etcdClient})
	})
}

// delClient 删除etcd client, 并关闭etcd client
func delClient(name string) {
	c := Get(name)
	clients.Delete(name)

	if c == nil {
		return
	}

	// 当old etcd client没有被使用的时候, 那么就关闭
	runtime.SetFinalizer(c, func(cc *Client) {
		log.Infof("[etcd] old etcd client %s object %d gc", name, uintptr(unsafe.Pointer(cc)))
		if err := cc.Close(); err != nil {
			log.Error("[etcd] old etcd client close error", xlog.String("name", name), xlog.Any("err", err))
		}
	})
}
