package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/xerror"
)

var CommitID = ""
var BuildTime = ""
var Data = ""
var Domain = ""
var Version = "v0.0.1-dev"
var Tag = ""

func init() {
	if Version == "" {
		Version = "v0.0.1-dev"
	}
	xerror.PanicErr(ver.NewVersion(Version))

	if Domain != "" {
		runmode.Domain = Domain
	}
}
