package etcdv3

import (
	"io"
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/pkg/retry"
)

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
	retry                retry.Retry
}

func (t Cfg) Build() io.Closer {
	var cfg clientv3.Config
	xerror.Panic(merge.CopyStruct(&cfg, &t))
	cfg.DialOptions = append(cfg.DialOptions, grpc.WithBlock())

	// 创建etcd client对象
	return xerror.PanicErr(t.retry.DoVal(
		func(i int) (interface{}, error) { return clientv3.New(cfg) }),
	).(*clientv3.Client)
}

func DefaultCfg() *Cfg {
	return &Cfg{
		retry:       retry.Default(),
		DialTimeout: time.Second * 2,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
}
