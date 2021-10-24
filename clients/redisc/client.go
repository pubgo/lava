package redisc

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/pubgo/lava/resource"
)

var _ resource.Resource = (*Client)(nil)

type Client struct {
	*redis.Client
}

func (t *Client) UpdateResObj(val interface{}) { t.Client = val.(*Client).Client }
func (t *Client) Kind() string                 { return Name }
func (t *Client) Get(ctx context.Context, options ...func(*redis.Options)) *redis.Client {
	var client = t.Client
	cc := client.WithContext(ctx)
	opts := cc.Options()

	// 默认的读写超时时间为 1s
	opts.WriteTimeout = DefaultRWTimeout
	opts.ReadTimeout = DefaultRWTimeout

	// 处理外部进来的参数配置
	for _, o := range options {
		o(opts)
	}

	return cc
}
