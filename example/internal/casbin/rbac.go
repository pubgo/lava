package casbin

import (
	"errors"

	"go.uber.org/zap"
)

type perm struct {
	Role string
	Org  string
	Name string
	Act  map[string]bool
}

func (c *Client) GetPermissions(user, role, org string, filter func(perm string) bool) (map[string]*perm, error) {
	if user == "" && role == "" {
		return nil, errors.New("user and role is null")
	}

	var users = make(map[string][]string)
	if user != "" {
		user = HandleUserId(user)
		domain, err := c.HandleOrgDomain(org, user)
		if err != nil {
			return nil, err
		}

		for i := range domain {
			userPerm, err := c.GetImplicitRolesForUser(user, domain[i])
			if err != nil {
				return nil, err
			}
			users[domain[i]] = userPerm
		}
	} else if role != "" {
		if org == "" || org == "*" {
			return nil, errors.New("when using role, you need to specify the value of org")
		}
		users[HandleOrgId(org)] = []string{HandleRoleId(role)}
	}

	c.logger.Debug("GetUsers", zap.Any("params", users))

	var permList = make(map[string]*perm)
	for domainName, roles := range users {
		for i := range roles {
			var userPerms = c.GetPermissionsForUser(roles[i], domainName)
			for _, p := range userPerms {
				if len(p) != 4 {
					continue
				}

				if !filter(p[2]) {
					continue
				}

				if permList[p[2]] == nil {
					permList[p[2]] = &perm{
						Org:  domainName,
						Role: roles[i],
						Name: p[2],
						Act:  map[string]bool{},
					}
				}
				permList[p[2]].Act[p[3]] = true
			}
		}
	}

	return permList, nil
}

func (c *Client) GetUserMethodPerms(sub, org string) map[string]bool {
	var mthList = make(map[string]bool)
	for _, perms := range c.GetPermissionsForUserInDomain(sub, org) {
		if len(perms) != 4 {
			continue
		}

		if perms[1] != org {
			continue
		}

		if !IsMethod(perms[2]) {
			continue
		}

		mthList[perms[2]] = true
	}
	return mthList
}

func (c *Client) EnforceMth(sub string, domain string, method string) (bool, error) {
	return c.Enforce(sub, domain, method, ActAccess)
}

func (c *Client) HandleOrgDomain(org string, sub string) ([]string, error) {
	var domain []string
	var err error

	switch org {
	case "":
		return nil, errors.New("org is null")
	case "*":
		domain, err = c.Enforcer.GetDomainsForUser(sub)
		if err != nil {
			return nil, err
		}
	default:
		domain = append(domain, HandleOrgId(org))
	}

	return domain, nil
}
