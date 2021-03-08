package etcdv3

import (
	"runtime"
	"unsafe"

	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
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

func newClient(cfg clientv3.Config) (c *Client, err error) {
	defer xerror.RespErr(&err)

	// etcd config处理
	cfg, err = cfgMerge(cfg)
	xerror.Panic(err)

	// 创建etcd client对象
	var etcdClient *clientv3.Client
	err = retry(3, func() error { etcdClient, err = clientv3.New(cfg); return err })
	xerror.PanicF(err, "[etcd] New error, err: %v, cfgList: %#v", err, cfg)

	return &Client{Client: etcdClient}, nil
}

// updateClient 更新etcd client
func updateClient(name string, cfg clientv3.Config) error {
	log.Debugf("[etcd] %s update etcd client", name)

	oldClient, ok := clients.Load(name)
	etcdClient, err := newClient(cfg)
	if err != nil {
		return err
	}

	clients.Set(name, etcdClient)

	if !ok || oldClient == nil {
		return nil
	}

	// 当old etcd client没有被使用的时候, 那么就关闭
	runtime.SetFinalizer(oldClient, func(cc *Client) {
		log.Infof("[etcd] old etcd client %s object %d gc", name, uintptr(unsafe.Pointer(cc)))
		if err := cc.Close(); err != nil {
			log.Errorf("[etcd] old etcd client close error, name: %s, err:%#v", name, err)
		}
	})

	return nil
}

// initClient 创建或者初始化etcd client
func initClient(name string, cfg clientv3.Config) {
	xerror.Assert(clients.Has(name), "[etcd] %s already exists", name)

	etcdClient, err := newClient(cfg)
	xerror.Panic(err)

	clients.Set(name, etcdClient)
}

// delClient 删除etcd client, 并关闭etcd client
func delClient(name string) {
	c := Get(name)
	clients.Delete(name)

	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		log.Errorf("[etcd] client close error, name:%s, err: %#v", name, err)
	}
}
