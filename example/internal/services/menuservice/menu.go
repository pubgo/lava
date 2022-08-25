package menuservice

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/example/internal/models"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
)

const Name = "menu"

var trim = strings.TrimSpace

func New(c *Config, db *orm.Client, l *logging.Logger) *Menu {
	var m = &Menu{c: c, db: db, defaultMenus: make(map[string]bool), logger: l.Named("menu"), mux: mux.NewRouter()}
	return m
}

type Menu struct {
	logger       *logging.Logger
	db           *orm.Client
	c            *Config
	defaultMenus map[string]bool
	mux          *mux.Router
}

func (m *Menu) IsDefaultMenu(name string) bool {
	return m.defaultMenus[name]
}

func (m *Menu) ListEndpointsWithCode(args ...interface{}) ([]*models.Endpoint, error) {
	var menus []*models.Endpoint
	var ctx = m.db.Model(&models.Endpoint{})
	if len(args) != 0 {
		ctx = ctx.Where(args[0], args[1:]...)
	}
	return menus, ctx.Preload("Action").Find(&menus).Error
}

func (m *Menu) GetAllCode() map[string]bool {
	var codes = make(map[string]bool)
	_ = m.mux.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		codes[route.GetName()] = true
		return nil
	})
	return codes
}

// LoadMenusFromDb load menu, casbin method enforce (method,path)=>code
func (m *Menu) LoadMenusFromDb() error {
	var routes, err = m.ListEndpointsWithCode()
	if err != nil && strings.Contains(err.Error(), "no such table") {
		return nil
	}

	if err != nil {
		return err
	}

	sort.Slice(routes, func(i, j int) bool {
		return !strings.Contains(routes[i].Path, "}")
	})

	var router = mux.NewRouter()
	for _, route := range routes {

		code := joinPath(route.Action.Type, route.ApiCode)

		// default menu
		for _, mth := range m.c.DefaultMenus {
			if strings.EqualFold(mth.Method, route.Method) && mth.Path == route.Path {
				m.defaultMenus[code] = true
			}
		}

		var r = m.handleRoute(router, route.Path, code, route.Method)
		if r == nil {
			return fmt.Errorf("handle route failed, route=%v", route)
		}

		if err := r.GetError(); err != nil {
			return err
		}
	}

	m.mux = router
	if m.c.PrintRoute {
		m.logger.Info("defaultMenus", zap.Any("defaultMenus", m.defaultMenus))
		_ = router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			var path, _ = route.GetPathTemplate()
			var mth, _ = route.GetMethods()
			m.logger.Info("route",
				zap.String("name", route.GetName()),
				zap.String("path", path),
				zap.Any("method", mth),
				zap.Any("default", m.defaultMenus[route.GetName()]),
			)
			return nil
		})
	}

	return nil
}

func (m *Menu) handleRoute(router *mux.Router, path string, code string, method string) *mux.Route {
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return nil
	}

	path = handlePath("", path)
	// trim '/'
	path = strings.TrimRight(path, "/")

	var methods []string
	if method == "access" || method == "*" {
		methods = allMethods[:]
	} else {
		methods = append(methods, method)
	}

	// /api/camera/*
	if path[len(path)-1] == '*' {
		return router.Methods(methods...).PathPrefix(path[:len(path)-1]).Name(code)
	} else {
		return router.Methods(methods...).Path(path).Name(code)
	}
}

func (m *Menu) SaveLocalMenusToDb() {
	var menuItems, err = parseMenuItems(m.c.Path)
	xerror.Panic(err)

	for _, item := range menuItems {
		xerror.Assert(item.Path == "" || item.Method == "" || item.Code == "", "path or method or code is null")

		xerror.Panic(m.db.Upsert(context.Background(), &models.Endpoint{
			TargetType: item.TargetType,
			Path:       item.Path,
			Method:     item.Method,
			ApiCode:    item.Code,
			Action: models.Action{
				Code: item.Code,
				Type: item.ResType,
				Name: item.DisplayName,
			},
		}, "path=? and method=?", item.Path, item.Method))

		for parentCode := range item.Parent {
			if parentCode == "" {
				continue
			}

			xerror.Panic(m.db.Upsert(context.Background(), &models.MenuItem{
				Code:       item.Code,
				ParentCode: parentCode,
				Platform:   "ka",
			}, "code=? and parent_code=?", item.Code, parentCode))
		}
	}
}

func (m *Menu) GetMethodName(method, url string) (string, error) {
	var match mux.RouteMatch
	method = strings.ToUpper(method)
	var req, err = http.NewRequest(method, url, nil)
	if err != nil {
		return "", err
	}

	if !m.mux.Match(req, &match) {
		return "", fmt.Errorf("path not found, method=%s path=%s", method, url)
	}

	if _, err = match.Route.GetPathTemplate(); err != nil {
		return "", err
	}

	return match.Route.GetName(), nil
}

func parseMenuItems(path string) (map[string]*menuMapping, error) {
	var menuItems = make(map[string]*menuMapping)
	var codeAndParent = make(map[string]map[string]bool)
	var err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".csv") {
			return nil
		}

		f, err := os.Open(path)
		xerror.Panic(err)
		defer f.Close()

		var r = csv.NewReader(f)
		records, err := r.ReadAll()
		xerror.Panic(err)

		for i := range records {
			// filter header
			if i == 0 {
				continue
			}

			m := &menuMapping{
				Path:        trim(records[i][0]),
				Code:        trim(records[i][1]),
				Method:      trim(records[i][2]),
				ResType:     trim(records[i][3]),
				TargetType:  trim(records[i][4]),
				ParentCode:  trim(records[i][5]),
				DisplayName: trim(records[i][7]),
			}

			if m.Code == "" && m.Path == "" {
				continue
			}

			if codeAndParent[m.Code] == nil {
				codeAndParent[m.Code] = map[string]bool{}
			}

			if m.ParentCode != "" {
				for _, p := range strings.Split(m.ParentCode, ",") {
					codeAndParent[m.Code][trim(p)] = true
				}
			}

			if m.Code != "" && m.Path != "" && m.Method != "" && m.ResType != "" {
				menuItems[fmt.Sprintf("%s%s", m.Path, m.Method)] = m
			}
		}
		return nil
	})
	xerror.Panic(err)

	for _, m := range menuItems {
		m.Parent = codeAndParent[m.Code]
	}

	return menuItems, nil
}

type menuMapping struct {
	ResType     string          `json:"res_type"`
	DisplayName string          `json:"display_name"`
	Path        string          `json:"path"`
	Code        string          `json:"code"`
	ParentCode  string          `json:"parent_code"`
	Parent      map[string]bool `json:"-"`
	Method      string          `json:"method"`
	TargetType  string          `json:"target_type"`
}
