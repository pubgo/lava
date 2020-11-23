package golug_env

import (
	"strconv"
)

// 默认的全局配置
var (
	Domain = "golug"
	Trace  = false
)

func init() {
	// 从环境变量中获取系统默认值
	// 获取系统默认的前缀, 环境变量前缀等
	Get(&Domain, "env_prefix")

	if trace := trim(GetEnv("trace")); trace != "" {
		Trace, _ = strconv.ParseBool(trace)
	}
}
