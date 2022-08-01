package version

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	"github.com/pubgo/funk/assert"
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
