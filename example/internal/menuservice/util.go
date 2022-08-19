package menuservice

import (
	"go.uber.org/zap"
	"net/http"
	"strings"

	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/internal/models"
	"github.com/pubgo/lava/example/pkg/proto/permpb"
)

var allMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

// handlePath http:/api/example=>/http__/api/example
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

func joinPath(paths ...string) string {
	return strings.Join(paths, "/")
}

func HandleMenuTree(menus []*models.Action, menuItems []*models.MenuItem, log *logging.Logger) []*permpb.MenuItem {
	var codeMap = make(map[string]*permpb.MenuItem)
	for _, item := range menus {
		codeMap[item.Code] = &permpb.MenuItem{
			Code: item.Code,
			Type: item.Type,
			Name: item.Name,
		}
	}

	var codeAndParent = make(map[string][]string)
	for i := range menuItems {
		codeAndParent[menuItems[i].Code] = append(codeAndParent[menuItems[i].Code], menuItems[i].ParentCode)
	}

	var nodeList []*permpb.MenuItem
	for _, item := range codeMap {
		if len(codeAndParent[item.Code]) == 0 {
			nodeList = append(nodeList, item)
			continue
		}

		for _, pp := range codeAndParent[item.Code] {
			pp = strings.TrimSpace(pp)
			if pp == "" {
				continue
			}

			if codeMap[pp] == nil {
				log.Error("menu parent code not found",
					zap.String("code", item.Code),
					zap.String("parent_code", pp),
				)
				nodeList = append(nodeList, item)
				continue
			}

			codeMap[pp].Children = append(codeMap[pp].Children, item)
		}
	}

	return nodeList
}
