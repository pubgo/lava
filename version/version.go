package version

import (
	"fmt"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
)

var commitID string
var buildTime string
var data string
var domain string
var version = "v0.0.1-dev-99"
var tag string
var project string
var deviceID = assert.Exit1(machineid.ID())
var instanceID = uuid.New().String()

func init() {
	defer recovery.Exit(func() {
		fmt.Println(
			project,
			version,
			commitID,
			buildTime,
		)
	})

	assert.If(project == "", "project is null")
	assert.If(version == "", "version is null")
	assert.If(commitID == "", "commitID is null")
	assert.If(buildTime == "", "buildTime is null")
}
