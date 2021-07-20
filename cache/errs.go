package cache

import (
	"github.com/pubgo/xerror"
)

var (
	Err                = xerror.New(Name)
	ErrKeyNotFound     = Err.New("key不存在")
	ErrKeyLength       = Err.New("key长度范围设置错误")
	ErrBufSize         = Err.New("现有的缓存超过了最大的限度或者小于最小的限度")
	ErrBufExceeded     = Err.New("现有的缓存超过了最大的缓存限制")
	ErrExpiration      = Err.New("过期时间设置错误")
	ErrDataLoadTimeOut = Err.New("DataLoad执行超时")
	ErrClearTime       = Err.New("定时清理过期缓存时间设置错误")
	ErrStore           = Err.New("存储器设置失败")
)
