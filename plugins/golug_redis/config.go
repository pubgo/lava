package golug_redis

import (
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/opentracing/opentracing-go/ext"
)

var Name = "redis"
var cfg = make(map[string]ClientCfg)
var options *redis.Options

const (
	DbType                  = "redis"
	SpanKind                = ext.SpanKindEnum("redis-client")
	MaxPipelineNameCmdCount = 3
	DefaultRWTimeout        = time.Second
)

type ClientCfg = redis.Options

func GetCfg() map[string]ClientCfg {
	return cfg
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{}
}
