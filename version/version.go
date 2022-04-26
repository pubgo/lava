package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/xerror"
)

var CommitID = ""
var BuildTime = ""
var Data = ""
var Domain = consts.Domain
var Version = "v0.0.1-dev"
var Tag = ""

func init() {
	xerror.PanicErr(ver.NewVersion(Version))
}
