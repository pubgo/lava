package etcdv3

import (
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/pubgo/lug/pkg/retry"
)

const Name = "etcdv3"

var logs = xlog.GetLogger(Name)
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
	RejectOldCluster     bool              `json:"reject_old_cluster"`
	PermitWithoutStream  bool              `json:"permit_without_stream"`
	DialOptions          []grpc.DialOption `json:"-"`
	backoff              retry.Handler
}

func (t Cfg) Build() (c *clientv3.Client, err error) {
	defer xerror.RespErr(&err)

	var cfg clientv3.Config
	xerror.Panic(merge.Copy(&cfg, &t))
	cfg.DialOptions = append(cfg.DialOptions, grpc.WithBlock())

	// 创建etcd client对象
	xerror.PanicF(t.backoff.Do(func(i int) error {
		c, err = clientv3.New(cfg)
		return xerror.Wrap(err)
	}), "[etcd] newClient error, err: %v, cfgList: %#v", err, cfg)

	return
}

func GetDefaultCfg() Cfg {
	return Cfg{
		backoff:     retry.New(),
		DialTimeout: time.Second * 2,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
}
