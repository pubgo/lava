package env

import (
	"github.com/joho/godotenv"
	"github.com/pubgo/xerror"
)

// Load 加载env文件
func Load(filenames ...string) {
	xerror.Panic(godotenv.Load(filenames...))
	initEnv()
}
