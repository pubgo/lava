package etcdv3

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
)

var data sync.Map

// GetClient 获取etcd client
func GetClient(names ...string) *Client {
	var name = golug_consts.Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}

	val, ok := data.Load(name)
	if !ok {
		log.Errorf("[etcd] %s not found", name)
		return nil
	}

	return val.(*Client)
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
	if err = retry(3, func() (err error) { etcdClient, err = clientv3.New(cfg); return }); err != nil {
		return nil, xerror.WrapF(err, "[etcd] New error, err: %v, cfg: %#v", err, cfg)
	}

	return &Client{Client: etcdClient}, nil
}

// UpdateClient 更新etcd client
func UpdateClient(name string, cfg clientv3.Config) error {
	log.Debugf("[etcd] %s update etcd client", name)

	oldClient, ok := data.Load(name)
	etcdClient, err := newClient(cfg)
	if err != nil {
		return err
	}

	data.Store(name, etcdClient)

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

// InitClient 创建或者初始化etcd client
func InitClient(name string, cfg clientv3.Config) error {
	_, ok := data.Load(name)
	if ok {
		return fmt.Errorf("[etcd] %s already exists", name)
	}

	etcdClient, err := newClient(cfg)
	if err != nil {
		return err
	}

	data.Store(name, etcdClient)
	return nil
}

// DelClient 删除etcd client, 并关闭etcd client
func DelClient(name string) {
	c := GetClient(name)
	data.Delete(name)

	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		log.Errorf("[etcd] client close error, name:%s, err: %#v", name, err)
	}
}

// ListClient etcd client list
func ListClient() map[string]*Client {
	var clients = make(map[string]*Client)
	data.Range(func(key, value interface{}) bool {
		clients[key.(string)] = value.(*Client)
		return true
	})
	return clients
}
