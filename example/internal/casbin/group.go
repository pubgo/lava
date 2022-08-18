package casbin

import (
	"strings"

	"github.com/pubgo/lava/logging/logutil"
	"go.uber.org/zap"

	"github.com/pubgo/lava/example/pkg/proto/permpb"
)

const ActAccess = "access"
const ActAll = "*"

func (c *Client) ListResources(sub string, domain string, resType string, act string) (map[string]map[string]bool, error) {
	var permissions, err = c.GetImplicitPermissionsForUser(sub, domain)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("getResources",
		zap.Any("params", []interface{}{sub, domain, resType, act}),
	)

	var nodePrefix = HandleNodeId(resType)
	var resPrefix = HandleResId(resType)
	var org = HandleOrgId(domain)

	// all user resources
	var resources = make(map[string]map[string]bool)
	for i := range permissions {
		if len(permissions[i]) != 4 {
			continue
		}

		// filter org
		if permissions[i][1] != org {
			continue
		}

		var _act = permissions[i][3]
		if !(_act == ActAccess || _act == ActAll || _act == act) {
			continue
		}

		var itemId = permissions[i][2]
		// itemId is res
		// res/resType/res_id
		if strings.HasPrefix(itemId, resPrefix) {
			if resources[itemId] == nil {
				resources[itemId] = make(map[string]bool)
			}
			resources[itemId][_act] = true
			continue
		}

		// itemId is node
		// group/resType/nodeType/itemId
		if strings.HasPrefix(itemId, nodePrefix) {
			var items = c.GetNodeResource(itemId, org)
			for j := range items {
				if resources[items[j]] == nil {
					resources[items[j]] = make(map[string]bool)
				}
				resources[items[j]][_act] = true
			}
		}
	}

	return resources, nil
}

func (c *Client) AddPerm(sub string, domain string, obj string, act string) (bool, error) {
	return c.Enforcer.AddPermissionForUser(sub, domain, obj, act)
}

func (c *Client) AddOrgMethodPerm(code string, domain string) error {
	var _, err = c.AddMethodPerm(
		HandleOrgRole(domain),
		HandleOrgId(domain),
		HandleMethod(code),
		ActAccess,
	)
	return err
}

func (c *Client) AddMethodPerm(sub string, domain string, name string, act string) (bool, error) {
	return c.Enforcer.AddPermissionForUser(sub, domain, name, act)
}

func (c *Client) AddOrgRootGroupPerm(orgId, resType string) (bool, error) {
	return c.Enforcer.AddPermissionForUser(
		OrgRootRole(orgId),
		HandleOrgId(orgId),
		ResRootGroup(orgId, resType),
		ActAccess,
	)
}

func (c *Client) GetAllNode(resType, orgId string) *group {
	return c.GetNode(
		make(map[string]*group),
		HandleOrgId(orgId),
		resType,
		ResRootGroup(orgId, resType),
		make(map[string]bool),
	)
}

type group struct {
	ResType   string
	GroupType string
	GroupId   string
	Resources []string
	Children  map[string]*group
	Acts      map[string]bool
}

func (g *group) Perm() *permpb.PermGroup {
	var n = &permpb.PermGroup{
		ResType:   g.ResType,
		GroupType: g.GroupType,
		GroupId:   g.GroupId,
		Resources: g.Resources,
	}

	for a := range g.Acts {
		n.Acts = append(n.Acts, a)
	}

	for _, c := range g.Children {
		n.Children = append(n.Children, c.Perm())
	}
	return n
}

func (g *group) Proto() *permpb.Group {
	var n = &permpb.Group{
		ResType:   g.ResType,
		GroupType: g.GroupType,
		GroupId:   g.GroupId,
		Resources: g.Resources,
	}
	for _, c := range g.Children {
		n.Children = append(n.Children, c.Proto())
	}
	return n
}

func NewMapGroup() map[string]*group { return make(map[string]*group) }

func (c *Client) GetNode(allNodes map[string]*group, orgId string, resType, itemId string, act map[string]bool) *group {
	if !IsNode(itemId) {
		itemId = HandleNodeId(JoinPath(resType, itemId))
	}

	c.logger.Debug("GetNode", zap.Any("info", []interface{}{orgId, resType, itemId}))
	var typ, id = HandleNodeTypeAndId(itemId)
	var node = &group{
		GroupId:   id,
		GroupType: typ,
		ResType:   resType,
		Acts:      act,
		Children:  map[string]*group{},
	}

	if allNodes[itemId] == nil {
		allNodes[itemId] = node
	} else {
		for k, v := range allNodes[itemId].Acts {
			act[k] = v
		}
		allNodes[itemId].Acts = act
	}

	for _, policy := range c.Enforcer.GetNamedGroupingPolicy("g2") {
		if len(policy) != 3 {
			continue
		}

		if orgId != "*" && orgId != policy[2] {
			continue
		}

		if policy[1] != itemId {
			continue
		}

		if IsRes(policy[0]) {
			node.Resources = append(node.Resources, GetResId(policy[0]))
		} else {
			node.Children[policy[0]] = c.GetNode(allNodes, orgId, resType, policy[0], act)
		}
	}
	return node
}

func (c *Client) MoveNode(toItemId, toNodeType, resType, itemId, nodeType, fromItemId, fromNodeType string, domain string) (bool, error) {
	if fromItemId == "" {
		fromItemId = ResRootGroup(domain, resType)
	}

	if !IsNode(toItemId) {
		toItemId = HandleNodeId(JoinPath(resType, toNodeType, toItemId))
	}

	if !IsNode(itemId) {
		itemId = HandleNodeId(JoinPath(resType, nodeType, itemId))
	}

	if !IsNode(fromItemId) {
		fromItemId = HandleNodeId(JoinPath(resType, fromNodeType, fromItemId))
	}

	_, err := c.Enforcer.AddNamedGroupingPolicy("g2", itemId, toItemId, HandleOrgId(domain))
	if err != nil {
		return false, err
	}

	_, err = c.Enforcer.RemoveNamedGroupingPolicy("g2", itemId, fromItemId, HandleOrgId(domain))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Client) DelNode(parentItemId, parentNodeType, resType, nodeType, itemId string, domain string) (bool, error) {
	if parentItemId == "" {
		parentItemId = ResRootGroup(domain, resType)
	}

	if !IsNode(parentItemId) {
		parentItemId = HandleNodeId(JoinPath(resType, parentNodeType, parentItemId))
	}

	if !IsNode(itemId) {
		itemId = HandleNodeId(JoinPath(resType, nodeType, itemId))
	}

	c.logger.Debug("DelNode", logutil.ListField("params", itemId, parentItemId))
	var ok, err = c.Enforcer.RemoveNamedGroupingPolicy("g2", itemId, parentItemId, HandleOrgId(domain))
	if err != nil {
		return false, err
	}

	for _, policy := range c.Enforcer.GetNamedGroupingPolicy("g2") {
		if policy[1] == itemId {
			c.logger.Debug("DelNode", logutil.ListField("params", policy[0], itemId))
			ok, err = c.Enforcer.RemoveNamedGroupingPolicy("g2", policy[0], itemId, HandleOrgId(domain))
			if err != nil {
				return false, err
			}
		}
	}

	return ok, nil
}

func (c *Client) AddNode(parentItemId, parentNodeType, resType, nodeType, itemId string, domain string) (bool, error) {
	if parentItemId == "" {
		parentItemId = ResRootGroup(domain, resType)
	}

	if !IsNode(parentItemId) {
		parentItemId = HandleNodeId(JoinPath(resType, parentNodeType, parentItemId))
	}

	if !IsNode(itemId) {
		itemId = HandleNodeId(JoinPath(resType, nodeType, itemId))
	}

	c.logger.Debug("AddNode", logutil.ListField("params", itemId, parentItemId))
	return c.Enforcer.AddNamedGroupingPolicy("g2", itemId, parentItemId, HandleOrgId(domain))
}

func (c *Client) AddRes(itemId, resType, nodeType, resId string, domain string) (bool, error) {
	if itemId == "" {
		itemId = ResRootGroup(domain, resType)
	}

	if !IsRes(resId) {
		resId = HandleResId(JoinPath(resType, resId))
	}

	if !IsNode(itemId) {
		itemId = HandleNodeId(JoinPath(resType, nodeType, itemId))
	}

	c.logger.Debug("addRes", logutil.ListField("params", resId, itemId))
	return c.Enforcer.AddNamedGroupingPolicy("g2", resId, itemId, HandleOrgId(domain))
}

func (c *Client) DelRes(itemId, resType, nodeType, resId string, domain string) (bool, error) {
	if itemId == "" {
		itemId = ResRootGroup(domain, resType)
	}

	if !IsRes(resId) {
		resId = HandleResId(JoinPath(resType, resId))
	}

	if !IsNode(itemId) {
		itemId = HandleNodeId(JoinPath(resType, nodeType, itemId))
	}

	c.logger.Debug("delRes", logutil.ListField("params", resId, itemId))
	return c.Enforcer.RemoveNamedGroupingPolicy("g2", resId, itemId, HandleOrgId(domain))
}

func (c *Client) GetNodeResource(itemId string, domain string) []string {
	var res []string
	var items = c.Enforcer.GetNamedGroupingPolicy("g2")
	for i := range items {
		if items[i][2] != domain {
			continue
		}

		if items[i][1] != itemId {
			continue
		}

		var val = items[i][0]
		if IsRes(val) {
			res = append(res, val)
		} else {
			res = append(res, c.GetNodeResource(val, domain)...)
		}
	}
	return res
}
