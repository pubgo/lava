package etcdv3

import (
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

const Name = "etcdv3"

var cfgList = make(map[string]Cfg)

type Cfg struct {
	Endpoints            []string          `json:"endpoints"`
	AutoSyncInterval     time.Duration     `json:"interval"`
	DialTimeout          time.Duration     `json:"timeout"`
	DialKeepAliveTime    time.Duration     `json:"keepalive"`
	DialKeepAliveTimeout time.Duration     `json:"keepalive_timeout"`
	MaxCallSendMsgSize   int               `json:"max_send"`
	MaxCallRecvMsgSize   int               `json:"max_recv"`
	Username             string            `json:"username"`
	Name                 string            `json:"name"`
	Password             string            `json:"password"`
	DialOptions          []grpc.DialOption `json:"-"`
}

func (t Cfg) Build() (c *clientv3.Client, err error) {
	defer xerror.RespErr(&err)

	var cfg clientv3.Config
	// 转化为etcd Cfg
	xerror.Panic(merge.Copy(&cfg, &t))

	// 创建etcd client对象
	var etcdClient *clientv3.Client
	err = retry(3, func() error { etcdClient, err = clientv3.New(cfg); return err })
	xerror.PanicF(err, "[etcd] newClient error, err: %v, cfgList: %#v", err, cfg)

	return etcdClient, nil
}

func GetDefaultCfg() Cfg {
	return Cfg{
		DialTimeout: time.Second * 2,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
}
