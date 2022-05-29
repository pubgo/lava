package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/consts"
)

var CommitID = ""
var BuildTime = ""
var Data = ""
var Domain = consts.Domain
var Version = "v0.0.1-dev"
var Tag = ""

func init() {
	if Version == "" {
		Version = "v0.0.1-dev"
	}
	xerror.PanicErr(ver.NewVersion(Version))
}
