package redisc

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/pubgo/lug/pkg/ctxutil"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
)

var clientM sync.Map

type Option func(options *redis.Options)

func Get(ctx context.Context, prefix string, options ...Option) *redis.Client {
	val, ok := clientM.Load(prefix)
	if !ok {
		xerror.Panic(xerror.Fmt("[redis] key [%s] not found", prefix))
	}

	cc := val.(*redis.Client).WithContext(ctx)
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

func Update(name string, cfg ClientCfg) {
	var ctx, cancel = ctxutil.Timeout()
	defer cancel()

	client := redis.NewClient(&cfg)
	xerror.Panic(client.Ping(ctx).Err(), "redis连接池连接失败")

	clientM.Store(name, client)
	zap.L().Info("rebuild redis pool done - " + name)
}
