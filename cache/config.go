package cache

import (
	"time"
)

const Name = "cache"

// Option 可选配置
type Option func(o *options)

// options 缓存配置变量
type options struct {
	store             IStore
	DataLoadTime      time.Duration
	ClearTime         time.Duration
	MaxBufSize        uint64
	DefaultExpiration time.Duration
	MaxExpiration     time.Duration
	MaxKeySize        uint64
}
