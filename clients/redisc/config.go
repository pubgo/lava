package redisc

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go/ext"
)

var Name = "redis"
var cfgMap = make(map[string]*Cfg)

const (
	DbType                  = "redis"
	SpanKind                = ext.SpanKindEnum("redis-client")
	MaxPipelineNameCmdCount = 3
	DefaultRWTimeout        = time.Second
)

type Cfg struct {
	Network            string        `json:"network" yaml:"network"`
	Addr               string        `json:"addr" yaml:"addr"`
	Username           string        `json:"username" yaml:"username"`
	Password           string        `json:"password" yaml:"password"`
	DB                 int           `json:"db" yaml:"db"`
	MaxRetries         int           `json:"max_retries" yaml:"max_retries"`
	MinRetryBackoff    time.Duration `json:"min_retry_backoff" yaml:"min_retry_backoff"`
	MaxRetryBackoff    time.Duration `json:"max_retry_backoff" yaml:"max_retry_backoff"`
	DialTimeout        time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout        time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout       time.Duration `json:"write_timeout" yaml:"write_timeout"`
	PoolSize           int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns       int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	MaxConnAge         time.Duration `json:"max_conn_age" yaml:"max_conn_age"`
	PoolTimeout        time.Duration `json:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	IdleCheckFrequency time.Duration `json:"idle_check_frequency" yaml:"idle_check_frequency"`
}

func DefaultCfg() *redis.Options {
	return &redis.Options{}
}
