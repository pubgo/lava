package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
)

var commitID string
var buildTime string
var data string
var domain string
var version string
var tag string
var project string

func init() {
	defer recovery.Exit()

	assert.If(version == "", "version is null")
	assert.If(project == "", "project is null")

	assert.Exit1(ver.NewVersion(version))
}

func CommitID() string {
	return commitID
}

func Version() string {
	return version
}

func BuildTime() string {
	return buildTime
}

func Data() string {
	return data
}

func Domain() string {
	return domain
}

func Tag() string {
	return tag
}

func Project() string {
	return project
}
