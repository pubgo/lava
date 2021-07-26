// +build mage

package main

import (
	"context"
	"fmt"
	"github.com/magefile/mage/mg"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

// Build A build step that requires additional params, or platform specific steps for example
func Build(ctx context.Context) error {
	mg.CtxDeps(ctx, Target)

	mg.Deps(InstallDeps)
	fmt.Println("Building...")
	return nil
	//cmd := exec.Command("go", "build", "-o", "MyApp", ".")
	//return cmd.Run()
}

// Install A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	return nil
	//return os.Rename("./MyApp", "/usr/bin/MyApp")
}

// InstallDeps Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	return nil
	//cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
	//return cmd.Run()
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	return
	//os.RemoveAll("MyApp")
}

// Target The first sentence in the comment will be the short help text shown with mage -l.
// The rest of the comment is long help text that will be shown with mage -h <target>
func Target() {
	fmt.Println("Hi!")
}
