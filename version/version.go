package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/env"
)

var CommitID = ""
var BuildTime = ""
var Data = ""
var Domain = consts.Domain
var Project = ""
var Version = "v0.0.1-dev"
var Tag = ""

func init() {
	xerror.PanicErr(ver.NewVersion(Version))

	if Project == "" {
		Project = env.Get("lava_project", "app_name", "project_name", "service_name")
	}
}
