package golug_redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	DbType                  = "redis"
	SpanKind                = ext.SpanKindEnum("redis-client")
	MaxPipelineNameCmdCount = 3
	DefaultRWTimeout        = time.Second
)

type Option func(options *redis.Options)

// GetRedis get a redis client with open tracing context
func GetRedis(ctx context.Context, prefix string, options ...Option) (*redis.Client, error) {
	client, err := PickupRedisClient(prefix)
	if err != nil {
		return nil, err
	}

	cc := client.WithContext(ctx)
	opts := cc.Options()

	// 默认的读写超时时间为 1s
	opts.WriteTimeout = DefaultRWTimeout
	opts.ReadTimeout = DefaultRWTimeout

	// 处理外部进来的参数配置
	for _, o := range options {
		o(opts)
	}

	return cc, nil
}

// =============================================================================================================
// Wrap the redis command process func
// GoRedisProcessFunc is an alias of cmd process func
type GoRedisProcessFunc = func(cmd redis.Cmder) error

// GoRedisWrapProcessFunc is an alias of wrapper that wrap process
type GoRedisWrapProcessFunc = func(oldProcess GoRedisProcessFunc) GoRedisProcessFunc

func setTag(span opentracing.Span, opts *redis.Options, method, key string) {
	ext.DBType.Set(span, DbType)
	ext.PeerAddress.Set(span, opts.Addr)
	ext.SpanKind.Set(span, SpanKind)

	// add redis command
	span.SetTag("db.method", method)
	span.SetTag("db.key", key)
}

// =============================================================================================================
// Wrap the redis command process pipeline func
// GoRedisProcessPipelineFunc is an alias of process pipeline func
type GoRedisProcessPipelineFunc = func([]redis.Cmder) error

// GoRedisWrapProcessPipelineFunc is an alias of wrapper that wrap pipeline
type GoRedisWrapProcessPipelineFunc = func(oldProcess GoRedisProcessPipelineFunc) GoRedisProcessPipelineFunc
