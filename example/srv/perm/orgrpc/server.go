package orgrpc

import (
	"context"
	"errors"

	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/pkg/casbin"
	"github.com/pubgo/lava/example/pkg/menuservice"
	"github.com/pubgo/lava/example/pkg/proto/permpb"
)

func New() permpb.OrgServiceServer {
	return &server{}
}

type server struct {
	Logger *logging.Logger
	Casbin *casbin.Client
	M      *menuservice.Menu
}

func (s *server) Init() {
	s.Logger = s.Logger.Named("perm.srv.org")
}

func (s *server) CreateOrg(ctx context.Context, req *permpb.CreateOrgRequest) (*permpb.CreateOrgResponse, error) {
	if req.OrgId == "" {
		return nil, errors.New("orgId is null")
	}

	// Add permission rule for org
	for code := range s.M.GetAllCode() {
		if err := s.Casbin.AddOrgMethodPerm(code, req.OrgId); err != nil {
			return nil, err
		}
	}

	// org root user
	if req.UserId != "" {
		var orgId = casbin.HandleOrgId(req.OrgId)
		var _, err = s.Casbin.AddRoleForUserInDomain(casbin.HandleUserId(req.UserId), casbin.HandleRoleId(orgId), orgId)
		if err != nil {
			return nil, err
		}
	}

	return &permpb.CreateOrgResponse{}, nil
}

func (s *server) DeleteOrg(ctx context.Context, req *permpb.DeleteOrgRequest) (*permpb.DeleteOrgResponse, error) {
	if req.OrgId == "" {
		return nil, errors.New("orgId is null")
	}

	var _, err = s.Casbin.DeleteDomains(casbin.HandleOrgId(req.OrgId))
	if err != nil {
		return nil, err
	}

	return &permpb.DeleteOrgResponse{}, nil
}

func (s *server) TransferOrg(ctx context.Context, req *permpb.TransferOrgRequest) (*permpb.TransferOrgResponse, error) {
	if req.OrgId == "" || req.UserId == "" || req.NewUserId == "" {
		return nil, errors.New("org_id or user_id or new_user_id is null")
	}

	var newUserId = casbin.HandleUserId(req.NewUserId)
	var userId = casbin.HandleUserId(req.UserId)
	var orgId = casbin.HandleOrgId(req.OrgId)
	var orgRole = casbin.HandleRoleId(orgId)

	// add new user to org role
	if _, err := s.Casbin.AddRoleForUserInDomain(newUserId, orgRole, orgId); err != nil {
		return nil, err
	}

	// del user
	if _, err := s.Casbin.DeleteRoleForUser(userId, orgRole, orgId); err != nil {
		return nil, err
	}

	return &permpb.TransferOrgResponse{}, nil
}

func (s *server) ListOrg(ctx context.Context, req *permpb.ListOrgRequest) (*permpb.ListOrgResponse, error) {
	var resp = &permpb.ListOrgResponse{}

	var domains, err = s.Casbin.GetAllDomains()
	if err != nil {
		return nil, err
	}

	resp.Orgs = domains
	return resp, nil
}
