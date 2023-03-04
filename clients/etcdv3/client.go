package etcdv3

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/retry"
	client3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func New(conf *Config) *Client {
	conf = merge.Copy(DefaultCfg(), conf).Unwrap()
	cfg := merge.Struct(new(client3.Config), conf).Unwrap()
	cfg.DialOptions = append(
		cfg.DialOptions,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	// 创建etcd client对象
	return &Client{Client: assert.Must1(retry.Default().DoVal(func(i int) (interface{}, error) {
		return client3.New(*cfg)
	})).(*client3.Client)}
}

type Client struct {
	*client3.Client
}
