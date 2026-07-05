// Package featurehttp 提供 feature flags 的 HTTP 调试接口
package featurehttp

import (
	"github.com/pubgo/funk/v2/features/featurehttp"

	"github.com/pubgo/lava/v2/core/debug"
)

func Register() {
	srv := featurehttp.NewServer("").WithPrefix("/debug/features")
	debug.App().All("/features", debug.Wrap(srv.Handler()))
	debug.App().All("/features/*", debug.Wrap(srv.Handler()))
}
