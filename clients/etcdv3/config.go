package etcdv3

import (
	"github.com/pubgo/lava/internal/pkg/retry"
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	DialOptions []grpc.DialOption `json:"-"`
	retry       retry.Retry
}

func (t *Cfg) Build() *etcdv3.Client {
	var cfg etcdv3.Config
	xerror.Panic(merge.CopyStruct(&cfg, &t))
	cfg.DialOptions = append(cfg.DialOptions, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// 创建etcd client对象
	return xerror.PanicErr(t.retry.DoVal(func(i int) (interface{}, error) { return etcdv3.New(cfg) })).(*etcdv3.Client)
}

func DefaultCfg() *Cfg {
	return &Cfg{
		retry:       retry.Default(),
		DialTimeout: time.Second * 2,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
}
