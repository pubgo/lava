package etcdv3

import (
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pubgo/funk/retry"
)

type Config struct {
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
	c                    *etcdv3.Client
}

func (t *Config) Get() *etcdv3.Client { return t.c }
func (t *Config) Build() (err error) {
	defer recovery.Err(&err)
	cfg := merge.Struct(new(etcdv3.Config), &t).
		Unwrap(func(err result.Error) result.Error { return err.WrapF("%#v", t) })
	cfg.DialOptions = append(cfg.DialOptions, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// 创建etcd client对象
	t.c = assert.Must1(t.retry.DoVal(func(i int) (interface{}, error) { return etcdv3.New(*cfg) })).(*etcdv3.Client)
	return
}

func DefaultCfg() *Config {
	return &Config{
		retry:       retry.Default(),
		DialTimeout: time.Second * 2,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
}
