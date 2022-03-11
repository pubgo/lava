package env

import (
	"os"
	"strings"
)

var Cfg = struct {
	// Prefix 系统环境变量前缀
	Prefix string

	// Separator 分隔符
	Separator string
}{
	Prefix:    os.Getenv(strings.ToUpper("env_prefix")),
	Separator: "_",
}
