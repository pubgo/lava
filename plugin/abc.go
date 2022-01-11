package plugin

import (
	"encoding/json"

	"github.com/pubgo/lava/types"
)

const Name = "plugin"

type Plugin interface {
	Process
	json.Marshaler
	String() string
	UniqueName() string
	Flags() types.Flags
	Commands() *types.Command
	Init() error
	Watch() types.Watcher
	Vars(types.Vars) error
	Health() types.Healthy
	Middleware() types.Middleware
}

type Process interface {
	BeforeStart(fn func())
	AfterStart(fn func())
	BeforeStop(fn func())
	AfterStop(fn func())
}
