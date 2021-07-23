package redisc

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/atomic"
)

type Client struct {
	atomic.Value
}

func (t *Client) Get(ctx context.Context, options ...Option) *redis.Client {
	var client = t.Load().(*redis.Client)

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
