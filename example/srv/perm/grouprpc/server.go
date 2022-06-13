package grouprpc

import (
	"context"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/pkg/casbin"
	"github.com/pubgo/lava/example/pkg/proto/permpb"
)

func New() permpb.GroupServiceServer {
	return &server{}
}

type server struct {
	Logger *logging.Logger
	Db     *orm.Client
	Casbin *casbin.Client
}

func (s *server) Init() {
	s.Logger = s.Logger.Named("perm.srv.group")
}

func (s *server) CreateGroup(ctx context.Context, req *permpb.CreateGroupRequest) (*permpb.CreateGroupResponse, error) {
	var rsp = new(permpb.CreateGroupResponse)

	if req.ParentGroupId == "" {
		if _, err := s.Casbin.AddOrgRootGroupPerm(req.OrgId, req.ResType); err != nil {
			return nil, err
		}
	}

	if _, err := s.Casbin.AddNode(req.ParentGroupId, req.ParentGroupType, req.ResType, req.GroupType, req.GroupId, req.OrgId); err != nil {
		return nil, err
	}

	for i := range req.Children {
		if _, err := s.Casbin.AddRes(req.GroupId, req.ResType, req.GroupType, req.Children[i], req.OrgId); err != nil {
			return nil, err
		}
	}

	return rsp, nil
}

func (s *server) DeleteGroup(ctx context.Context, req *permpb.DeleteGroupRequest) (*permpb.DeleteGroupResponse, error) {
	var rsp = new(permpb.DeleteGroupResponse)
	if _, err := s.Casbin.DelNode(req.ParentGroupId, req.ParentGroupType, req.ResType, req.GroupType, req.GroupId, req.OrgId); err != nil {
		return nil, err
	}
	return rsp, nil
}

func (s *server) MoveGroup(ctx context.Context, req *permpb.MoveGroupRequest) (*permpb.MoveGroupResponse, error) {
	var rsp = new(permpb.MoveGroupResponse)
	_, err := s.Casbin.MoveNode(req.ToGroupId, req.ToGroupType, req.ResType, req.CurGroupId, req.CurGroupType, req.FromGroupId, req.FromGroupType, req.OrgId)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (s *server) ListGroups(ctx context.Context, req *permpb.ListGroupsRequest) (*permpb.ListGroupsResponse, error) {
	var resp = new(permpb.ListGroupsResponse)
	var group = s.Casbin.GetAllNode(req.ResType, req.OrgId)
	resp.Groups = append(resp.Groups, group.Proto())
	return resp, nil
}

func (s *server) AddResForGroup(ctx context.Context, req *permpb.AddResForGroupRequest) (*permpb.AddResForGroupResponse, error) {
	if req.GroupId == "" {
		if _, err := s.Casbin.AddOrgRootGroupPerm(req.OrgId, req.ResType); err != nil {
			return nil, err
		}
	}

	var rsp = new(permpb.AddResForGroupResponse)
	var _, err = s.Casbin.AddRes(req.GroupId, req.ResType, req.GroupType, req.ResId, req.OrgId)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (s *server) DelResForGroup(ctx context.Context, req *permpb.DelResForGroupRequest) (*permpb.DelResForGroupResponse, error) {
	var rsp = new(permpb.DelResForGroupResponse)
	var _, err = s.Casbin.DelRes(req.GroupId, req.ResType, req.GroupType, req.ResId, req.OrgId)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}
