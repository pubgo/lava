package etcdv3

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

const Name = "etcdv3"

var DefaultCfg = clientv3.Config{
	DialTimeout: time.Second * 2,
	DialOptions: []grpc.DialOption{grpc.WithBlock()},
}

var _ json.Unmarshaler = (*duration)(nil)
var _ json.Marshaler = (*duration)(nil)

type duration int64

func (d duration) MarshalJSON() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *duration) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, "\"")

	dur, err := time.ParseDuration(string(data))
	if err != nil {
		return xerror.WrapF(err, "data: %s", data)
	}

	*d = duration(dur)

	return nil
}

type config struct {
	Endpoints            []string          `json:"endpoints"`
	AutoSyncInterval     duration          `json:"interval"`
	DialTimeout          duration          `json:"timeout"`
	DialKeepAliveTime    duration          `json:"keepalive"`
	DialKeepAliveTimeout duration          `json:"keepalive_timeout"`
	MaxCallSendMsgSize   int               `json:"max_send"`
	MaxCallRecvMsgSize   int               `json:"max_recv"`
	Username             string            `json:"username"`
	Password             string            `json:"password"`
	DialOptions          []grpc.DialOption `json:"-"`
}

// 转化为etcd config
func (t config) EtcdConfig() (cfg clientv3.Config) {
	cfg.Endpoints = t.Endpoints
	cfg.AutoSyncInterval = time.Duration(t.AutoSyncInterval)
	cfg.DialTimeout = time.Duration(t.DialTimeout)
	cfg.DialKeepAliveTime = time.Duration(t.DialKeepAliveTime)
	cfg.DialKeepAliveTimeout = time.Duration(t.DialKeepAliveTimeout)
	cfg.MaxCallSendMsgSize = t.MaxCallSendMsgSize
	cfg.MaxCallRecvMsgSize = t.MaxCallRecvMsgSize
	cfg.Username = t.Username
	cfg.Password = t.Password
	cfg.DialOptions = t.DialOptions
	return cfg
}
