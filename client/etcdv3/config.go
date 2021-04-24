package etcdv3

import (
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

const Name = "etcdv3"

var log = xlog.Named(Name)
var cfgList = make(map[string]Cfg)

type Cfg struct {
	Endpoints            []string          `json:"endpoints"`
	AutoSyncInterval     time.Duration     `json:"auto_sync_interval"`
	DialTimeout          time.Duration     `json:"dial_timeout"`
	DialKeepAliveTime    time.Duration     `json:"dial_keep_alive_time"`
	DialKeepAliveTimeout time.Duration     `json:"dial_keep_alive_timeout"`
	MaxCallSendMsgSize   int               `json:"max_send_size"`
	MaxCallRecvMsgSize   int               `json:"max_recv_size"`
	Username             string            `json:"username"`
	Name                 string            `json:"name"`
	Password             string            `json:"password"`
	RejectOldCluster     bool              `json:"reject-old-cluster"`
	PermitWithoutStream  bool              `json:"permit-without-stream"`
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
