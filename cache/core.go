package cache

import (
	"sync"

	"go.uber.org/atomic"
)

// 默认扩展因子, 用来计算全局扩展缓存
const DefaultMaxBufFactor = 0.2

// 默认最小缓存
const DefaultMinBufSize = 1024 * 1024

// 默认最大缓存512M
var DefaultMaxBufSize = uint64(512 * 1024 * 1024)

// 当前cache实例数
var CacheCount atomic.Int32

// 最大cache实例数
var MaxCacheCount = 20

// GlobalMaxBufSize 全局最大缓存
// 全局最大缓存被所有缓存实例共享, 是所有缓存实例的缓存总和的最大值上限
var GlobalMaxBufSize uint64

// GlobalMaxBufExpand 全局最大的扩展缓存, 超过最大扩展缓存就会直接返回调用者错误信息
var GlobalMaxBufExpand float64
var mutex sync.Mutex

// GlobalBufSize 全局所有缓存实例的缓存总和
var GlobalBufSize atomic.Uint64

func init() {
	GlobalMaxBufSize = 500 * 1024 * 1024
	GlobalMaxBufExpand = (DefaultMaxBufFactor + 1) * float64(GlobalMaxBufSize)
}

// SetDefaultMaxBufSize 设置默认最大缓存, 设置时请仔细思考和计算[warning]
func SetDefaultMaxBufSize(maxBufSize uint64) {
	mutex.Lock()
	defer mutex.Unlock()

	if maxBufSize < DefaultMaxBufSize {
		return
	}
	DefaultMaxBufSize = maxBufSize
}

// SetGlobalMaxBufSize 全局最大缓存设置
func SetGlobalMaxBufSize(maxBufSize uint64) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if maxBufSize < DefaultMinBufSize || maxBufSize > DefaultMaxBufSize {
		return false
	}

	GlobalMaxBufSize = maxBufSize
	GlobalMaxBufExpand = (DefaultMaxBufFactor + 1) * float64(GlobalMaxBufSize)
	return true
}
