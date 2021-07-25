package cache

import (
	"github.com/pubgo/xerror"
)

var (
	Err                = xerror.New(Name)
	ErrNotFound        = Err.New("key不存在")
	ErrKeyExpired        = Err.New("缓存过期")
	ErrKeyLength       = Err.New("key长度范围设置错误")
	ErrBufExceeded     = Err.New("现有的缓存超过了最大的缓存限制")
	ErrExpiration      = Err.New("过期时间设置错误")
	ErrDataLoadTimeOut = Err.New("DataLoad执行超时")
	ErrClearTime       = Err.New("定时清理过期缓存时间设置错误")
)
