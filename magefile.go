//go:build mage
// +build mage

package main

import (
	"github.com/pubgo/lava/mages"

	// mage:import
	_ "github.com/pubgo/lava/mages"
)

var Default = Build

// Build lava
func Build() error {
	return mages.GoBuild("lava", "cmd/lava/main.go")
}
