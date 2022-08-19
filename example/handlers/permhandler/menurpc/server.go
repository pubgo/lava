package menurpc

import (
	"context"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/internal/menuservice"
	"github.com/pubgo/lava/example/internal/models"
	"github.com/pubgo/lava/example/pkg/proto/permpb"
)

type server struct {
	Logger *logging.Logger
	Db     *orm.Client
	M      *menuservice.Menu
}

func New() permpb.MenuServiceServer {
	return &server{}
}

func (p *server) Init() {
	p.Logger = p.Logger.Named("perm.srv.menu")
}

func (p *server) ListMenus(ctx context.Context, req *permpb.ListMenusRequest) (*permpb.ListMenusResponse, error) {
	var resp = new(permpb.ListMenusResponse)
	var menuItems []*models.MenuItem
	var db = p.Db.WithContext(ctx)
	if req.Platform != "" {
		db = db.Where("platform=?", req.Platform)
	}

	err := db.Find(&menuItems).Error
	if err != nil {
		return nil, err
	}

	var menus []*models.Action
	if err := p.Db.Model(&models.Action{}).Find(&menus).Error; err != nil {
		return nil, err
	}

	resp.Items = menuservice.HandleMenuTree(menus, menuItems, p.Logger)
	return resp, nil
}
