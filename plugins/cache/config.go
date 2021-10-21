package cache

import (
	"github.com/pubgo/xerror"

	"time"
)

const Name = "cache"

const (
	// DefaultExpiration 默认过期时间
	defaultExpiration time.Duration = -1
	// 最小过期时间
	defaultMinExpiration = time.Second * 2
	// 最大过期时间10m
	defaultMaxExpiration = time.Minute * 10

	// key长度限制
	// key最小长度
	defaultMinKeySize = uint32(5)
	// key最大长度
	defaultMaxKeySize = 65535

	// DataLoadTimeOutTime 默认的数据加载过期时间
	dataLoadTimeOutTime = time.Second * 15

	// 默认过期定时清理时间2m
	defaultClearTime = time.Minute * 2

	// DefaultMaxBufFactor 默认扩展因子, 用来计算全局扩展缓存
	DefaultMaxBufFactor = 0.2

	// DefaultMinBufSize 默认最小缓存
	DefaultMinBufSize = 1024 * 1024

	// DefaultMaxBufSize 默认最大缓存512M
	DefaultMaxBufSize = 512 * 1024 * 1024
)

// Cfg 缓存配置变量
type Cfg struct {
	Store             string
	DataLoadTime      time.Duration
	ClearTime         time.Duration
	MaxBufSize        uint32
	DefaultExpiration time.Duration
	MaxExpiration     time.Duration
	MaxKeySize        uint32
}

func (t Cfg) Build() (_ *cacheImpl, err error) {
	defer xerror.RespErr(&err)

	// 最大缓存大小判断
	xerror.Assert(t.MaxBufSize > DefaultMaxBufSize || t.MaxBufSize < DefaultMinBufSize,
		"现有的缓存超过了最大的限度或者小于最小的限度 ErrBufSize:%d, min:%d, max:%d", t.MaxBufSize, DefaultMinBufSize, DefaultMaxBufSize)

	// 最大过期时间判断
	xerror.Assert(t.MaxExpiration > defaultMaxExpiration || t.MaxExpiration < defaultMinExpiration,
		"过期时间设置错误 MaxExpiration:%d, min:%d, max:%d", t.DefaultExpiration, defaultMinExpiration, defaultMaxExpiration)

	// 默认过期时间判断
	xerror.Assert(t.DefaultExpiration < defaultMinExpiration || t.DefaultExpiration > t.MaxExpiration,
		"过期时间设置错误 DefaultExpiration:%d, min:%d, max:%d", t.DefaultExpiration, defaultMinExpiration, defaultMaxExpiration)

	// key 长度判断
	xerror.Assert(t.MaxKeySize > defaultMaxKeySize || t.MaxKeySize < defaultMinKeySize,
		"key长度范围设置错误 MaxKeySize:%d, min:%d, max:%d", t.MaxKeySize, defaultMinKeySize, defaultMaxKeySize)

	var store = GetStore(t.Store)

	// 存储器判断
	xerror.Assert(store == nil, "存储器设置失败 store is nil")

	var p = &cacheImpl{store: store}
	xerror.Panic(initJanitor(p))

	// 过期驱逐设置
	p.store.OnEvicted(p.defaultEvicted)
	return p, nil
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Store:             "map",
		ClearTime:         defaultClearTime,
		DataLoadTime:      dataLoadTimeOutTime,
		MaxBufSize:        DefaultMaxBufSize,
		MaxExpiration:     defaultMaxExpiration,
		MaxKeySize:        defaultMaxKeySize,
		DefaultExpiration: -1,
	}
}
