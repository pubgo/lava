package etcdv3

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
)

var data types.SMap

// Get 获取etcd client
func Get(names ...string) *Client {
	var name = consts.GetDefault(names...)

	xerror.Assert(data.Has(name), "[etcdv3] %s not found", name)

	return data.Get(name).(*Client)
}

func New(cfg clientv3.Config) (*clientv3.Client, error) {
	c, err := newClient(cfg)
	return c.Client, err
}

func newClient(cfg clientv3.Config) (*Client, error) {
	var err error

	// etcd config处理
	cfg, err = cfgMerge(cfg)
	if err != nil {
		return nil, err
	}

	// 创建etcd client对象
	var etcdClient *clientv3.Client
	err = retry(3, func() error { etcdClient, err = clientv3.New(cfg); return err })
	if err != nil {
		return nil, xerror.WrapF(err, "[etcd] New error, err: %v, cfg: %#v", err, cfg)
	}

	return &Client{Client: etcdClient}, nil
}

// Update 更新etcd client
func Update(name string, cfg clientv3.Config) error {
	log.Debugf("[etcd] %s update etcd client", name)

	oldClient, ok := data.Load(name)
	etcdClient, err := newClient(cfg)
	if err != nil {
		return err
	}

	data.Set(name, etcdClient)

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

// Init 创建或者初始化etcd client
func Init(name string, cfg clientv3.Config) error {
	if data.Has(name) {
		return fmt.Errorf("[etcd] %s already exists", name)
	}

	etcdClient, err := newClient(cfg)
	if err != nil {
		return err
	}

	data.Set(name, etcdClient)
	return nil
}

// Del 删除etcd client, 并关闭etcd client
func Del(name string) {
	c := Get(name)
	data.Delete(name)

	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		log.Errorf("[etcd] client close error, name:%s, err: %#v", name, err)
	}
}

// List etcd client list
func List() (dt map[string]*Client) {
	data.Map(&dt)
	return
}
