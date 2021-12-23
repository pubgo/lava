package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/xerror"
)

var CommitID = ""
var BuildTime = "2021-03-20 16:52:09"
var Data = ""
var Domain = "lava"
var Version = "v0.0.1"

func init() {
	xerror.ExitErr(ver.NewVersion(Version))
}
