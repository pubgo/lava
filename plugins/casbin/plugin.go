package casbin

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		OnInit: func() {
			e, _ := casbin.NewEnforcer("authz_model.conf", "authz_policy.csv")
			_, err := e.DeleteRolesForUser("cathy")
			xerror.Panic(err)
		},
	})
}

// BasicAuthorizer stores the casbin handler
type BasicAuthorizer struct {
	enforcer *casbin.Enforcer
}

// GetUserName gets the user name from the request.
// Currently, only HTTP basic authentication is supported
func (a *BasicAuthorizer) GetUserName(r *http.Request) string {
	username, _, _ := r.BasicAuth()
	return username
}

// CheckPermission checks the user/method/path combination from the request.
// Returns true (permission granted) or false (permission forbidden)
func (a *BasicAuthorizer) CheckPermission(r *http.Request) bool {
	user := a.GetUserName(r)
	method := r.Method
	path := r.URL.Path

	allowed, err := a.enforcer.Enforce(user, path, method)
	if err != nil {
		panic(err)
	}

	return allowed
}
