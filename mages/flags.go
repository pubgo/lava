package mages

import (
	"fmt"
	"time"

	"github.com/pubgo/lug/consts"
)

const baseProject = "github.com/pubgo/lug"

func GoLdFlags(domain string, data string) string {
	var flags = `LDFLAGS=-ldflags " `
	flags += fmt.Sprintf(`-X '%s/version.BuildTime=%s'`, baseProject, time.Now().Format(consts.DefaultDateFormat))
	flags += fmt.Sprintf(`-X '%s/version.CommitID=%s'`, baseProject, GitHash(8))
	flags += fmt.Sprintf(`-X '%s/version.Version=%s'`, baseProject, GitTag())
	flags += fmt.Sprintf(`-X '%s/version.Domain=%s'`, baseProject, domain)
	flags += fmt.Sprintf(`-X '%s/version.Data=%s'`, baseProject, data)
	flags += ` " `
	return flags
}
