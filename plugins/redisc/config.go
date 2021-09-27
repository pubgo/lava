package redisc

import (
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go/ext"
)

var Name = "redis"
var cfgMap = make(map[string]ClientCfg)

const (
	DbType                  = "redis"
	SpanKind                = ext.SpanKindEnum("redis-client")
	MaxPipelineNameCmdCount = 3
	DefaultRWTimeout        = time.Second
)

type ClientCfg = redis.Options

func DefaultCfg() ClientCfg {
	return ClientCfg{}
}
