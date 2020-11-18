package golug_redis

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/pubgo/xlog"
)

var resMap sync.Map

type config struct {
	Addr        string
	Db          int
	Password    string
	PoolSize    int
	IdleTimeout int
}

//组装转换数据
func buildRedisData(prefix string, value []byte) error {
	if len(value) == 0 {
		return errors.New("watch时读取json出错,prefix=" + prefix)
	}

	var redisConfig config
	err := json.Unmarshal(value, &redisConfig)
	if err != nil {
		return errors.New("json unmarshal, error=" + err.Error())
	}
	if redisConfig.Addr == "" {
		return errors.New("未找到addr配置， " + prefix + "无效")
	}

	redisClient := newRedisPool(&redisConfig)
	ping := redisClient.Ping()
	if ping.Val() == "" {
		return errors.New("redis连接池连接失败,error=" + ping.Err().Error())
	}
	resMap.Store(prefix, redisClient)
	xlog.Info("rebuild redis pool done - " + prefix)

	return nil
}

//new redis pool from config center by watch
func newRedisPool(redisConfig *config) *redis.Client {
	redisCli := redis.NewClient(&redis.Options{
		Addr:        redisConfig.Addr,
		DB:          redisConfig.Db,
		PoolSize:    redisConfig.PoolSize,
		IdleTimeout: time.Duration(redisConfig.IdleTimeout) * time.Millisecond,
		Password:    redisConfig.Password,
	})
	return redisCli
}

func PickupRedisClient(prefix string) (*redis.Client, error) {
	result, ok := resMap.Load(prefix)
	if ok {
		return result.(*redis.Client), nil
	}

	xlog.Errorf("can not get redis client ,prefix=" + prefix)
	return nil, errors.New("can not get redis client ,prefix=" + prefix)
}
