package rolerpc

import (
	"context"
	"strings"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/internal/casbin"
	"github.com/pubgo/lava/example/internal/models"
	"github.com/pubgo/lava/example/pkg/proto/permpb"
)

func New() permpb.RoleServiceServer {
	return &server{}
}

type server struct {
	Logger *logging.Logger
	Casbin *casbin.Client
	Db     *orm.Client
}

func (s *server) Init() {
	s.Logger = s.Logger.Named("perm.srv.role")
}

func (s *server) CreateRole(ctx context.Context, req *permpb.CreateRoleRequest) (*permpb.CreateRoleResponse, error) {
	var role = models.RoleFromProto(req.Role)
	err := s.Db.Upsert(ctx, role, "name=? and org_id=?", role.Name, casbin.HandleOrgId(role.OrgId))
	if err != nil {
		return nil, err
	}
	return &permpb.CreateRoleResponse{Role: role.Proto()}, nil
}

func (s *server) DeleteRole(ctx context.Context, req *permpb.DeleteRoleRequest) (*permpb.DeleteRoleResponse, error) {
	var role = models.Role{ID: uint(req.Id)}
	if req.Id == 0 {
		err := s.Db.WithContext(ctx).Where("name=? and org_id=?", req.Name, casbin.HandleOrgId(req.OrgId)).First(&role).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := s.Db.WithContext(ctx).First(&role).Error
		if err != nil {
			return nil, err
		}
	}

	var domain = casbin.HandleOrgId(role.OrgId)

	// del rbac role
	for _, u := range s.Casbin.Enforcer.GetAllUsersByDomain(domain) {
		_, _ = s.Casbin.DeleteRoleForUserInDomain(u, casbin.HandleOrgRole(role.Name), domain)
	}

	err := s.Db.WithContext(ctx).Delete(&role).Error
	if err != nil {
		return nil, err
	}
	return &permpb.DeleteRoleResponse{}, nil
}

func (s *server) UpdateRole(ctx context.Context, req *permpb.UpdateRoleRequest) (*permpb.UpdateRoleResponse, error) {
	var role = models.RoleFromProto(req.Role)
	if req.Role.Id == 0 {
		err := s.Db.WithContext(ctx).Where("name=? and org_id=?", req.Role.Name, casbin.HandleOrgId(req.Role.OrgId)).Updates(role).Error
		if err != nil {
			return nil, err
		}
		return &permpb.UpdateRoleResponse{}, nil
	}

	err := s.Db.WithContext(ctx).Updates(role).Error
	if err != nil {
		return nil, err
	}
	return &permpb.UpdateRoleResponse{Role: role.Proto()}, nil
}

func (s *server) GetRole(ctx context.Context, req *permpb.GetRoleRequest) (*permpb.GetRoleResponse, error) {
	var role = &models.Role{ID: uint(req.Id)}
	err := s.Db.WithContext(ctx).Where("id=?", req.Id).First(role).Error
	if orm.ErrNotFound(err) {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	return &permpb.GetRoleResponse{Role: role.Proto()}, nil
}

func (s *server) ListRoles(ctx context.Context, req *permpb.ListRolesRequest) (*permpb.ListRolesResponse, error) {
	var db = s.Db.WithContext(ctx)
	if req.OrgId != "" {
		db = db.Where("org_id=?", req.OrgId)
	}

	var roles []*models.Role
	err := db.Find(&roles).Error
	if err != nil {
		return nil, err
	}

	var resp = new(permpb.ListRolesResponse)
	for i := range roles {
		resp.Roles = append(resp.Roles, roles[i].Proto())
	}
	return resp, nil
}

func (s *server) AddRoleForUser(ctx context.Context, req *permpb.AddRoleForUserRequest) (_ *permpb.AddRoleForUserResponse, err error) {
	var resp = new(permpb.AddRoleForUserResponse)
	resp.Ok, err = s.Casbin.AddRoleForUser(
		casbin.HandleUserId(req.UserId),
		casbin.HandleRoleId(req.RoleId),
		casbin.HandleOrgId(req.OrgId),
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *server) DelRoleForUser(ctx context.Context, req *permpb.DelRoleForUserRequest) (_ *permpb.DelRoleForUserResponse, err error) {
	var resp = new(permpb.DelRoleForUserResponse)
	user := casbin.HandleUserId(req.UserId)
	domain := casbin.HandleOrgId(req.OrgId)

	if req.RoleId == "*" {
		resp.Ok, err = s.Casbin.DeleteRolesForUser(user, domain)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	resp.Ok, err = s.Casbin.DeleteRoleForUser(
		user,
		casbin.HandleRoleId(req.RoleId),
		domain,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *server) GetRolesForUser(ctx context.Context, req *permpb.GetRolesForUserRequest) (_ *permpb.GetRolesForUserResponse, err error) {
	var resp = new(permpb.GetRolesForUserResponse)
	resp.Roles, err = s.Casbin.GetImplicitRolesForUser(
		casbin.HandleUserId(req.UserId),
		casbin.HandleOrgId(req.OrgId),
	)
	if err != nil {
		return nil, err
	}

	for i := range resp.Roles {
		resp.Roles[i] = casbin.GetLast(strings.Split(resp.Roles[i], "/")...)
	}
	return resp, nil
}

func (s *server) GetUsersForRole(ctx context.Context, req *permpb.GetUsersForRoleRequest) (*permpb.GetUsersForRoleResponse, error) {
	var resp = new(permpb.GetUsersForRoleResponse)

	if req.RoleId == "" {
		resp.Users = s.Casbin.GetAllUsersByDomain(casbin.HandleOrgId(req.OrgId))
	} else {
		resp.Users = s.Casbin.GetUsersForRoleInDomain(
			casbin.HandleRoleId(req.RoleId),
			casbin.HandleOrgId(req.OrgId),
		)
	}
	for i := range resp.Users {
		resp.Users[i] = casbin.GetLast(strings.Split(resp.Users[i], "/")...)
	}

	return resp, nil
}
