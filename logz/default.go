package logz

import (
	"sync"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/runenv"
)

var discard = zap.NewNop()
var loggerMap sync.Map
var debugLog = func() *zap.Logger {
	defer xerror.RespExit()
	cfg := xlog_config.NewDevConfig()
	cfg.EncoderConfig.EncodeCaller = "full"
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat
	cfg.Rotate = nil

	var log = cfg.Build(runenv.Project)
	return log.Named("lava")
}()

func init() {
	type sysLog struct {
		// log依赖注入, ns:lava
		Log *zap.Logger `dix:"lava"`
	}

	xerror.Exit(dix.Provider(func(s sysLog) {
		// 系统日志, 用于记录模块和组件的信息
		debugLog = s.Log.Named("lava").With(zap.Bool("system", true), zap.Namespace("fields"))
		loggerMap.Range(func(key, value interface{}) bool {
			loggerMap.Store(key, debugLog.Named(key.(string)))
			return true
		})
		xerror.Exit(dix.Provider(&Log{}))
	}))
}

type Log struct{}

func On(fn func(*Log)) {
	xerror.Exit(dix.Provider(fn))
}