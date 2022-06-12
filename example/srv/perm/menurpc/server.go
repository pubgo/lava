package menurpc

import (
	"context"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/pkg/menuservice"
	"github.com/pubgo/lava/example/pkg/models"
	"github.com/pubgo/lava/example/pkg/proto/permpb"
)

type server struct {
	logger *logging.Logger
	db     *orm.Client
	m      *menuservice.Menu
}

func New(l *logging.Logger, db *orm.Client, m *menuservice.Menu) permpb.MenuServiceServer {
	return &server{
		m:      m,
		db:     db,
		logger: l.Named("perm.srv.menu"),
	}
}

func (p *server) ListMenus(ctx context.Context, req *permpb.ListMenusRequest) (*permpb.ListMenusResponse, error) {
	var resp = new(permpb.ListMenusResponse)
	var menuItems []*models.MenuItem
	var db = p.db.WithContext(ctx)
	if req.Platform != "" {
		db = db.Where("platform=?", req.Platform)
	}

	err := db.Find(&menuItems).Error
	if err != nil {
		return nil, err
	}

	var menus []*models.Action
	if err := p.db.Model(&models.Action{}).Find(&menus).Error; err != nil {
		return nil, err
	}

	resp.Items = menuservice.HandleMenuTree(menus, menuItems, p.logger)
	return resp, nil
}
