package env

import (
	"github.com/joho/godotenv"
)

// Load 加载env文件
// 	默认：.env
func Load(filenames ...string) {
	_ = godotenv.Load(filenames...)
	initEnv()
}
