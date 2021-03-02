package etcdv3

import (
	"time"

	"github.com/pubgo/golug/types"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

const Name = "etcdv3"

var DefaultCfg = clientv3.Config{
	DialTimeout: time.Second * 2,
	DialOptions: []grpc.DialOption{grpc.WithBlock()},
}

type config struct {
	Endpoints            []string          `json:"endpoints"`
	AutoSyncInterval     types.Duration    `json:"interval"`
	DialTimeout          types.Duration    `json:"timeout"`
	DialKeepAliveTime    types.Duration    `json:"keepalive"`
	DialKeepAliveTimeout types.Duration    `json:"keepalive_timeout"`
	MaxCallSendMsgSize   int               `json:"max_send"`
	MaxCallRecvMsgSize   int               `json:"max_recv"`
	Username             string            `json:"username"`
	Password             string            `json:"password"`
	DialOptions          []grpc.DialOption `json:"-"`
}

// 转化为etcd config
func (t config) EtcdConfig() (cfg clientv3.Config) {
	cfg.Endpoints = t.Endpoints
	cfg.AutoSyncInterval = t.AutoSyncInterval.Duration
	cfg.DialTimeout = t.DialTimeout.Duration
	cfg.DialKeepAliveTime = t.DialKeepAliveTime.Duration
	cfg.DialKeepAliveTimeout = t.DialKeepAliveTimeout.Duration
	cfg.MaxCallSendMsgSize = t.MaxCallSendMsgSize
	cfg.MaxCallRecvMsgSize = t.MaxCallRecvMsgSize
	cfg.Username = t.Username
	cfg.Password = t.Password
	cfg.DialOptions = t.DialOptions
	return cfg
}
