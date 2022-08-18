package version

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
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
var deviceID = assert.Exit1(machineid.ID())
var instanceID = uuid.New().String()

func init() {
	defer recovery.Exit()
	assert.If(version == "", "version is null")
	assert.If(project == "", "project is null")
	assert.If(commitID == "", "commit id is null")
	assert.If(buildTime == "", "build time is null")
	assert.Exit1(ver.NewVersion(version))
}
