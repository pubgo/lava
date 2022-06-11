package casbin

import (
	"errors"
	"fmt"
	"strings"
)

func HandleUserId(id string) string {
	if strings.HasPrefix(id, "user/") {
		return id
	}

	return fmt.Sprintf("user/%s", id)
}

func HandleOrgRole(id string) string {
	return HandleRoleId(HandleOrgId(id))
}

func HandleRoleId(id string) string {
	if strings.HasPrefix(id, "role/") {
		return id
	}

	return fmt.Sprintf("role/%s", id)
}

func HandleOrgId(id string) string {
	if strings.HasPrefix(id, "org_") {
		return id
	}

	return "org_" + id
}

func HandleNodeId(id string) string {
	if strings.HasPrefix(id, "group/") {
		return id
	}

	return fmt.Sprintf("group/%s", id)
}

func HandleResId(id string) string {
	if strings.HasPrefix(id, "res/") {
		return id
	}

	return fmt.Sprintf("res/%s", id)
}

func HandleMethod(id string) string {
	if strings.HasPrefix(id, "mth/") {
		return id
	}

	return fmt.Sprintf("mth/%s", id)
}

func handlePath(protocol, path string) string {
	path = strings.Trim(path, "/")

	if protocol != "" {
		return "/" + protocol + "__/" + path
	}

	if strings.Contains(path, ":") {
		path = strings.ReplaceAll(path, ":", "__/")
	} else {
		path = "http__/" + path
	}

	path = "/" + strings.Trim(path, "/")
	return strings.ReplaceAll(path, "//", "/")
}

func JoinPath(paths ...string) string {
	return strings.Join(paths, "/")
}

func HandleSub(userId, roleId string) (string, error) {
	if userId == "" && roleId == "" {
		return "", errors.New("user and role is null")
	}

	if userId != "" {
		return HandleUserId(userId), nil
	}

	return HandleRoleId(roleId), nil
}

func HandleObj(resType, nodeType, resId string) (string, error) {
	if resType == "" || resId == "" {
		return "", errors.New("resType or resId is null")
	}

	if IsMth(resType) {
		return handlePath(resType, strings.Split(resId, "?")[0]), nil
	}

	if nodeType == "" {
		return HandleResId(JoinPath(resType, resId)), nil
	} else {
		return HandleNodeId(JoinPath(resType, nodeType, resId)), nil
	}
}

func OrgRootRole(orgId string) string {
	return HandleRoleId(HandleOrgId(orgId))
}

func ResRootGroup(orgId, resType string) string {
	return HandleNodeId(JoinPath(resType, "org", HandleOrgId(orgId)))
}

func IsNode(id string) bool {
	return strings.HasPrefix(id, "group/")
}

func IsRes(id string) bool {
	return strings.HasPrefix(id, "res/")
}

func IsMethod(id string) bool {
	return strings.HasPrefix(id, "mth/")
}

func GetResId(name string) string {
	var names = strings.Split(name, "/")
	return names[len(names)-1]
}

func HandleNodeTypeAndId(id string) (string, string) {
	names := strings.Split(id, "/")
	if len(names) == 4 {
		return names[len(names)-2], names[len(names)-1]
	}

	panic(fmt.Sprintf("id(%s) error", id))
}

func IsMth(resType string) bool {
	switch resType {
	case "mth", "api", "ws", "grpc", "general", "http", "ui", "action":
		return true
	default:
		return false
	}
}

func GetLast(values ...string) string {
	if len(values) == 0 {
		return ""
	}

	return values[len(values)-1]
}
