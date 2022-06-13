package permrpc

import (
	"context"
	"errors"
	"strings"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"

	"github.com/pubgo/lava/example/pkg/casbin"
	"github.com/pubgo/lava/example/pkg/menuservice"
	"github.com/pubgo/lava/example/pkg/models"
	"github.com/pubgo/lava/example/pkg/proto/permpb"
)

func New() permpb.PermServiceServer {
	return &server{}
}

type server struct {
	Logger *logging.Logger
	Casbin *casbin.Client
	M      *menuservice.Menu
	Db     *orm.Client
}

func (s *server) Init() {
	s.Logger = s.Logger.Named("perm.srv.rbac")
}

func (s *server) ListResources(ctx context.Context, req *permpb.PermServiceListResourcesRequest) (*permpb.PermServiceListResourcesResponse, error) {
	sub, err := casbin.HandleSub(req.UserId, req.RoleId)
	if err != nil {
		return nil, err
	}

	domain, err := s.Casbin.HandleOrgDomain(casbin.HandleOrgId(req.OrgId), sub)
	if err != nil {
		return nil, err
	}

	if len(domain) == 0 {
		return nil, errors.New("domain is null")
	}

	resp := &permpb.PermServiceListResourcesResponse{}
	var act = req.Act

	for i := range domain {
		resources, err := s.Casbin.ListResources(sub, domain[i], req.ResType, "*")
		if err != nil {
			return nil, err
		}

		for k, acts := range resources {
			if casbin.IsNode(k) {
				continue
			}

			if !acts[act] && !acts["access"] && !acts["*"] {
				continue
			}

			k = casbin.GetResId(k)
			var res = &permpb.Resource{ResId: k}
			for m := range acts {
				res.Acts = append(res.Acts, m)
			}
			resp.Resources = append(resp.Resources, res)
		}
	}
	return resp, nil
}

func (s *server) ListMenus(ctx context.Context, req *permpb.PermServiceListMenusRequest) (*permpb.PermServiceListMenusResponse, error) {
	sub, err := casbin.HandleSub(req.UserId, req.RoleId)
	if err != nil {
		return nil, err
	}

	domain, err := s.Casbin.HandleOrgDomain(casbin.HandleOrgId(req.OrgId), sub)
	if err != nil {
		return nil, err
	}

	if len(domain) == 0 {
		return nil, errors.New("domain is null")
	}

	mthList := make(map[string]bool)
	for i := range domain {
		for name := range s.Casbin.GetUserMethodPerms(sub, domain[i]) {
			mthList[name] = true
		}
	}

	var names []string
	for name := range mthList {
		names = append(names, casbin.GetLast(strings.Split(name, "/")...))
	}

	var menuItems []*models.MenuItem
	var db = s.Db.WithContext(ctx)
	if req.Platform != "" {
		db = db.Where("platform=?", req.Platform)
	}

	if err := db.Find(&menuItems).Error; err != nil {
		return nil, err
	}

	resp := new(permpb.PermServiceListMenusResponse)
	var menus []*models.Action
	if err := s.Db.Model(&models.Action{}).Where("code in ?", names).Find(&menus).Error; err != nil {
		return nil, err
	}

	resp.Items = menuservice.HandleMenuTree(menus, menuItems, s.Logger)
	return resp, nil
}

func (s *server) ListGroups(ctx context.Context, req *permpb.PermServiceListGroupsRequest) (*permpb.PermServiceListGroupsResponse, error) {
	sub, err := casbin.HandleSub(req.UserId, req.RoleId)
	if err != nil {
		return nil, err
	}

	domain, err := s.Casbin.HandleOrgDomain(casbin.HandleOrgId(req.OrgId), sub)
	if err != nil {
		return nil, err
	}

	if len(domain) == 0 {
		return nil, errors.New("domain is null")
	}

	var act = req.Act
	var prefix = casbin.HandleNodeId(req.ResType)
	perms, err := s.Casbin.GetPermissions(req.UserId, req.RoleId, req.OrgId, func(perm string) bool {
		return strings.HasPrefix(perm, prefix)
	})
	if err != nil {
		return nil, err
	}

	s.Logger.Debug("ListCatalogs", logutil.ListField("params", perms))

	var resp = new(permpb.PermServiceListGroupsResponse)
	var nodes = make(map[string]*permpb.PermGroup)
	var allNodes = casbin.NewMapGroup()
	for i := range perms {
		if !perms[i].Act[act] && !perms[i].Act["access"] && !perms[i].Act["*"] {
			continue
		}

		for j := range domain {
			var itemId = perms[i].Name
			var node = s.Casbin.GetNode(allNodes, domain[j], req.ResType, itemId, perms[i].Act)
			if nodes[itemId] == nil {
				nodes[itemId] = node.Perm()
			} else {
				for k := range perms[i].Act {
					nodes[itemId].Acts = append(nodes[itemId].Acts, k)
				}
			}
		}
	}

	for k := range nodes {
		resp.Groups = append(resp.Groups, nodes[k])
	}

	return resp, nil
}

func (s *server) SaveRolePerm(ctx context.Context, req *permpb.PermServiceSaveRolePermRequest) (*permpb.PermServiceSaveRolePermResponse, error) {
	const separator = ":::"

	var role = casbin.HandleRoleId(req.RoleId)
	var domain = casbin.HandleOrgId(req.OrgId)

	menus, err := s.M.ListEndpointsWithCode("api_code in ?", req.Menus)
	if err != nil {
		return nil, err
	}

	var reqPerms = make(map[string]bool)
	var resMap = make(map[string][]string)
	for i := range menus {
		code := menus[i].Action.Code
		targetType := menus[i].TargetType
		resMap[targetType] = append(resMap[targetType], code)
		var data = []string{role, domain, casbin.HandleMethod(casbin.JoinPath(menus[i].Action.Type, code)), casbin.ActAccess}
		reqPerms[strings.Join(data, separator)] = true
	}

	for i := range req.Groups {
		resType := req.Groups[i].ResType
		groupType := req.Groups[i].GroupType
		groupId := req.Groups[i].GroupId

		for _, act := range resMap[resType] {
			var obj = casbin.HandleResId(casbin.JoinPath(resType, groupId))
			if groupType != "" {
				obj = casbin.HandleNodeId(casbin.JoinPath(resType, groupType, groupId))
			}
			var data = []string{role, domain, obj, act}
			reqPerms[strings.Join(data, separator)] = true
		}
	}

	var perms = make(map[string]bool)
	perm, err := s.Casbin.GetImplicitPermissionsForUser(role, domain)
	if err != nil {
		return nil, err
	}

	for _, p := range perm {
		perms[strings.Join(p, separator)] = true
	}

	// add perm
	for p1 := range reqPerms {
		if !perms[p1] {
			if _, err := s.Casbin.AddPermissionForUser(role, strings.Split(p1, separator)[1:]...); err != nil {
				return nil, err
			}
		}
	}

	// del perm
	for p1 := range perms {
		if !reqPerms[p1] {
			if _, err := s.Casbin.DeletePermissionForUser(role, strings.Split(p1, separator)[1:]...); err != nil {
				return nil, err
			}
		}
	}

	return &permpb.PermServiceSaveRolePermResponse{}, nil
}

func (s *server) Enforce(ctx context.Context, req *permpb.EnforceRequest) (*permpb.EnforceResponse, error) {
	sub, err := casbin.HandleSub(req.UserId, req.RoleId)
	if err != nil {
		return nil, err
	}

	obj, err := casbin.HandleObj(req.ResType, req.GroupType, req.ResId)
	if err != nil {
		return nil, err
	}

	domain, err := s.Casbin.HandleOrgDomain(casbin.HandleOrgId(req.OrgId), sub)
	if err != nil {
		return nil, err
	}

	if len(domain) == 0 {
		return nil, errors.New("domain is null")
	}

	var act = req.Act
	var name string
	if casbin.IsMth(req.ResType) {
		mthName, err := s.M.GetMethodName(act, obj)
		if err != nil {
			return nil, err
		}
		name = mthName
	}

	for i := range domain {
		var ok bool
		var err error
		if name != "" {
			s.Logger.Debug("method enforce", logutil.ListField("params", sub, domain[i], casbin.HandleMethod(name)))
			ok, err = s.Casbin.EnforceMth(sub, domain[i], casbin.HandleMethod(name))
		} else {
			s.Logger.Debug("resource enforce", logutil.ListField("params", sub, domain[i], obj, act))
			ok, err = s.Casbin.Enforce(sub, domain[i], obj, act)
		}

		if err != nil {
			return nil, err
		}

		if ok {
			return &permpb.EnforceResponse{Ok: ok, Code: casbin.GetLast(strings.Split(name, "/")...)}, nil
		}
	}

	return &permpb.EnforceResponse{Ok: false, Code: casbin.GetLast(strings.Split(name, "/")...)}, nil
}
