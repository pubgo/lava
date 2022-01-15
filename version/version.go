package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/consts"
)

var CommitID = ""
var BuildTime = "2021-03-20 16:52:09"
var Data = ""
var Domain = consts.Domain
var Version = "v0.0.1"
var Tag = "v0.0.1"

func init() {
	xerror.ExitErr(ver.NewVersion(Version))
}
