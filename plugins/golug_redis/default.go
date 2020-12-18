package golug_redis

import (
	"context"
	"errors"
	"sync"

	"github.com/go-redis/redis/v7"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var clientM sync.Map

type Option func(options *redis.Options)

func GetClient(ctx context.Context, prefix string, options ...Option) *redis.Client {
	val, ok := clientM.Load(prefix)
	if !ok {
		xerror.Next().Panic(xerror.Fmt("%s not found", prefix))
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
	redisClient := redis.NewClient(&cfg)
	ping := redisClient.Ping()
	if ping.Val() == "" {
		xerror.Exit(errors.New("redis连接池连接失败,error=" + ping.Err().Error()))
	}
	clientM.Store(name, redisClient)
	xlog.Info("rebuild redis pool done - " + name)
}
