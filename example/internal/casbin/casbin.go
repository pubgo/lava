package casbin

import (
	libcasbin "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/xerror"
)

const Name = "casbin"

type Client struct {
	*libcasbin.Enforcer
	logger *logging.Logger
	db     *orm.Client
}

func New(cfg config.Config, l *logging.Logger, db *orm.Client) *Client {
	var c Config
	xerror.Panic(cfg.UnmarshalKey(Name, &c))

	xerror.Panic(c.Check())

	m, err := model.NewModelFromFile(c.Model)
	xerror.Panic(err)

	var enforcer *libcasbin.Enforcer

	if db != nil {
		ada, err := gormadapter.NewAdapterByDB(db.DB)
		xerror.Panic(err)

		e, err := libcasbin.NewEnforcer(m, ada)
		xerror.Panic(err)

		xerror.Panic(e.LoadPolicy())
		enforcer = e
	} else {
		e, err := libcasbin.NewEnforcer(m)
		xerror.Panic(err)
		enforcer = e
	}

	enforcer.EnableLog(c.EnableLog)
	enforcer.EnableAutoSave(c.EnableAutoSave)

	return &Client{
		Enforcer: enforcer,
		logger:   l.With().Named("casbin"),
		db:       db,
	}
}
