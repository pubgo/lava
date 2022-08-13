package envs

import (
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/internal/pkg/env"
	"github.com/pubgo/lava/internal/pkg/utils"
)

var (
	AppEnv = utils.GetDefault(env.Get(consts.AppEnv), "local")
)
