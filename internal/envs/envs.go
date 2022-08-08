package envs

import (
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/internal/pkg/env"
	"github.com/pubgo/lava/internal/pkg/utils"
	"github.com/pubgo/lava/version"
)

var (
	EnvPrefix = utils.GetDefault(env.Get(consts.EnvCfgPrefix), version.Project())
)
