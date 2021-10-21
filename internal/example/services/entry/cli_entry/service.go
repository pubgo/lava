package cli_entry

import (
	"fmt"
	"time"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/entry/cliEntry"
	db2 "github.com/pubgo/lava/plugins/db"

	"github.com/pubgo/x/fx"
	"go.uber.org/zap"
)

var _ cliEntry.Service = (*Service)(nil)

type Service struct {
	Db *db2.Client `dix:""`
}

func (t *Service) Run() map[string]cliEntry.Handler {
	return map[string]cliEntry.Handler{
		consts.Default: func(ctx fx.Ctx) {
			fmt.Println("db ping:", t.Db.Ping())
			zap.L().Info("cliEntry hello once")
		},
	}
}

func (t *Service) RunLoop() map[string]cliEntry.Handler {
	return map[string]cliEntry.Handler{
		"hello": func(ctx fx.Ctx) {
			fmt.Println("db ping:", t.Db.Ping())
			zap.L().Info("cliEntry hello forever")
			time.Sleep(time.Second)
		},
	}
}
