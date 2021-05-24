package redisc

import (
	"context"
	"errors"
	"github.com/pubgo/lug/pkg/ctxutil"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var clientM sync.Map

type Option func(options *redis.Options)

func GetClient(ctx context.Context, prefix string, options ...Option) *redis.Client {
	val, ok := clientM.Load(prefix)
	if !ok {
		xerror.Panic(xerror.Fmt("%s not found", prefix))
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

func initClient(name string, cfg ClientCfg) {
	var ctx, cancel = ctxutil.Timeout()
	defer cancel()

	redisClient := redis.NewClient(&cfg)
	ping := redisClient.Ping(ctx)
	if ping.Val() == "" {
		xerror.Exit(errors.New("redis连接池连接失败,error=" + ping.Err().Error()))
	}

	clientM.Store(name, redisClient)
	xlog.Info("rebuild redis pool done - " + name)
}
