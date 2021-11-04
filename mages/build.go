package mages

import (
	"fmt"
	"time"

	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/shutil"
)

func GoBuild(domain string, dir string) (err error) {
	defer xerror.RespErr(&err)

	var commitID = shutil.MustRun("git rev-parse --short=8 HEAD")
	var buildTime = time.Now().Format(consts.DefaultTimeFormat)
	var version = shutil.MustRun("git describe --abbrev=0 --tags")

	ldFlags := `-ldflags "`
	ldFlags += fmt.Sprintf(" -X 'github.com/pubgo/lava/version.BuildTime=%s'", buildTime)
	ldFlags += fmt.Sprintf(" -X 'github.com/pubgo/lava/version.CommitID=%s'", commitID)
	ldFlags += fmt.Sprintf(" -X 'github.com/pubgo/lava/version.Version=%s'", version)
	ldFlags += fmt.Sprintf(" -X 'github.com/pubgo/lava/version.Domain=%s'", domain)
	ldFlags += fmt.Sprintf(" -X 'github.com/pubgo/lava/version.Data=hello'")
	ldFlags += `"`

	fmt.Println(shutil.MustRun("go", "build", ldFlags, "-mod vendor", "-v", "-o main", dir))
	return
}
