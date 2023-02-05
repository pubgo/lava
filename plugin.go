package lava

// 加载插件
import (
	// set GOMAXPROCS
	_ "github.com/pubgo/lava/modules/automaxprocs"

	// gc plugin
	_ "github.com/pubgo/lava/modules/gcnotifier"

	// 用于系统诊断
	_ "github.com/pubgo/lava/modules/gops"
)

// 加载插件
import (
	// 加载protobuf编码
	_ "github.com/pubgo/lava/encoding/protobuf"

	// 默认注册中心
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"

	_ "github.com/pubgo/lava/logging/logext/grpclog"
)
