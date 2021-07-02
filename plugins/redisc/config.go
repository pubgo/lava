package redisc

import (
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/xlog"
)

var Name = "redis"
var cfg = make(map[string]ClientCfg)
var logs = xlog.GetLogger(Name)

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