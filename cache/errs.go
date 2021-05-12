package cache

import "errors"

var (
	ErrKeyNotFound     = errors.New("key不存在")
	ErrKeyLength       = errors.New("key长度范围设置错误")
	ErrBufSize         = errors.New("现有的缓存超过了最大的限度或者小于最小的限度")
	ErrBufExceeded     = errors.New("现有的缓存超过了最大的缓存限制")
	ErrExpiration      = errors.New("过期时间设置错误")
	ErrDataLoadTimeOut = errors.New("DataLoad执行超时")
	ErrClearTime       = errors.New("定时清理过期缓存时间设置错误")
	ErrStore           = errors.New("存储器设置失败")
)
